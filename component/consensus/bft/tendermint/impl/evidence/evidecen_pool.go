/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 11:44 上午
# @File : evidecen_pool.go
# @Description :
# @Attention :
*/
package evidence

import (
	"bytes"
	"errors"
	"fmt"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/google/orderedcode"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/light"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
	"github.com/hyperledger/fabric-droplib/libs/clist"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// prefixes are unique across all tm db's
	prefixCommitted = int64(9)
	prefixPending   = int64(10)
)

type Pool struct {
	sync.Mutex

	logger logplugin.Logger

	evidenceStore services.DB
	evidenceList  *clist.CList // concurrent linked-list of evidence
	evidenceSize  uint32       // amount of pending evidence

	// needed to load validators to verify evidence
	stateDB services.IStateStore
	// needed to load headers and commits to verify evidence
	blockStore services.IBlockStore

	// latest state
	state models.LatestState

	// evidence from consensus is buffered to this slice, awaiting until the next height
	// before being flushed to the pool. This prevents broadcasting and proposing of
	// evidence before the height with which the evidence happened is finished.
	consensusBuffer []duplicateVoteSet

	pruningHeight int64
	pruningTime   time.Time
}

// NewPool creates an evidence pool. If using an existing evidence store,
// it will add all pending evidence to the concurrent list.
func NewDefaultEvidencePool(evidenceDB services.DB, stateDB services.IStateStore, blockStore services.IBlockStore) (*Pool, error) {
	state, err := stateDB.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	pool := &Pool{
		stateDB:         stateDB,
		blockStore:      blockStore,
		state:           state,
		logger:          logplugin.GlobalLogger("EVIDENCE_POOL"),
		evidenceStore:   evidenceDB,
		evidenceList:    clist.New(),
		consensusBuffer: make([]duplicateVoteSet, 0),
	}

	// If pending evidence already in db, in event of prior failure, then check
	// for expiration, update the size and load it back to the evidenceList.
	pool.pruningHeight, pool.pruningTime = pool.removeExpiredPendingEvidence()
	evList, _, err := pool.listEvidence(prefixPending, -1)
	if err != nil {
		return nil, err
	}

	atomic.StoreUint32(&pool.evidenceSize, uint32(len(evList)))

	for _, ev := range evList {
		pool.evidenceList.PushBack(ev)
	}

	return pool, nil
}

func (evpool *Pool) PendingEvidence(maxBytes int64) ([]transfer.Evidence, int64) {
	if evpool.Size() == 0 {
		return []transfer.Evidence{}, 0
	}
	evidence, size, err := evpool.listEvidence(prefixPending, maxBytes)
	if err != nil {
		evpool.logger.Error("failed to retrieve pending evidence", "err", err)
	}

	return evidence, size
}
func (evpool *Pool) Update(state models.LatestState, ev models.EvidenceList) {
	// sanity check
	if state.LastBlockHeight <= evpool.state.LastBlockHeight {
		panic(fmt.Sprintf(
			"failed EvidencePool.Update new state height is less than or equal to previous state height: %d <= %d",
			state.LastBlockHeight,
			evpool.state.LastBlockHeight,
		))
	}

	evpool.logger.Debug(
		"updating evidence pool",
		"last_block_height", state.LastBlockHeight,
		"last_block_time", state.LastBlockTime,
	)

	// flush conflicting vote pairs from the buffer, producing DuplicateVoteEvidence and
	// adding it to the pool
	evpool.processConsensusBuffer(state)
	// update state
	evpool.updateState(state)

	// move committed evidence out from the pending pool and into the committed pool
	evpool.markEvidenceAsCommitted(ev, int64(state.LastBlockHeight))

	// Prune pending evidence when it has expired. This also updates when the next
	// evidence will expire.
	if evpool.Size() > 0 && int64(state.LastBlockHeight) > evpool.pruningHeight &&
		state.LastBlockTime.After(evpool.pruningTime) {
		evpool.pruningHeight, evpool.pruningTime = evpool.removeExpiredPendingEvidence()
	}
}
func (evpool *Pool) AddEvidence(ev transfer.Evidence) error {
	evpool.logger.Debug("attempting to add evidence", "evidence", ev)

	// We have already verified this piece of evidence - no need to do it again
	if evpool.isPending(ev) {
		evpool.logger.Debug("evidence already pending; ignoring", "evidence", ev)
		return nil
	}

	// check that the evidence isn't already committed
	if evpool.isCommitted(ev) {
		// This can happen if the peer that sent us the evidence is behind so we
		// shouldn't punish the peer.
		evpool.logger.Debug("evidence was already committed; ignoring", "evidence", ev)
		return nil
	}

	// 1) Verify against state.
	if err := evpool.verify(ev); err != nil {
		return err
	}

	// 2) Save to store.
	if err := evpool.addPendingEvidence(ev); err != nil {
		return fmt.Errorf("failed to add evidence to pending list: %w", err)
	}

	// 3) Add evidence to clist.
	evpool.evidenceList.PushBack(ev)

	evpool.logger.Info("verified new evidence of byzantine behavior", "evidence", ev)
	return nil
}
func (evpool *Pool) ReportConflictingVotes(voteA, voteB *models.Vote) {
	evpool.Lock()
	defer evpool.Unlock()
	evpool.consensusBuffer = append(evpool.consensusBuffer, duplicateVoteSet{
		VoteA: voteA,
		VoteB: voteB,
	})
}

func (evpool *Pool) CheckEvidence(evList models.EvidenceList) error {
	hashes := make([][]byte, len(evList))
	for idx, ev := range evList {

		ok := evpool.fastCheck(ev)

		if !ok {
			// check that the evidence isn't already committed
			if evpool.isCommitted(ev) {
				return &commonerrors.ErrInvalidEvidence{Evidence: ev, Reason: errors.New("evidence was already committed")}
			}

			err := evpool.verify(ev)
			if err != nil {
				return err
			}

			if err := evpool.addPendingEvidence(ev); err != nil {
				// Something went wrong with adding the evidence but we already know it is valid
				// hence we log an error and continue
				evpool.logger.Error("failed to add evidence to pending list", "err", err, "evidence", ev)
			}

			evpool.logger.Info("verified new evidence of byzantine behavior", "evidence", ev)
		}

		// check for duplicate evidence. We cache hashes so we don't have to work them out again.
		hashes[idx] = ev.Hash()
		for i := idx - 1; i >= 0; i-- {
			if bytes.Equal(hashes[i], hashes[idx]) {
				return &commonerrors.ErrInvalidEvidence{Evidence: ev, Reason: errors.New("duplicate evidence")}
			}
		}
	}

	return nil
}

// EvidenceFront goes to the first evidence in the clist
func (evpool *Pool) EvidenceFront() *clist.CElement {
	return evpool.evidenceList.Front()
}

// EvidenceWaitChan is a channel that closes once the first evidence in the list
// is there. i.e Front is not nil.
func (evpool *Pool) EvidenceWaitChan() <-chan struct{} {
	return evpool.evidenceList.WaitChan()
}
func (evpool *Pool) fastCheck(ev transfer.Evidence) bool {
	if lcae, ok := ev.(*models.LightClientAttackEvidence); ok {
		key := keyPending(ev)
		evBytes, err := evpool.evidenceStore.Get(key)
		if evBytes == nil { // the evidence is not in the nodes pending list
			return false
		}

		if err != nil {
			evpool.logger.Error("failed to load light client attack evidence", "err", err, "key(height/hash)", key)
			return false
		}

		// LightClientAttackEvidence
		var trustedPb types.LightClientAttackEvidenceProto

		if err = commonutils.UnMarshal(evBytes, &trustedPb); err != nil {
			evpool.logger.Error(
				"failed to convert light client attack evidence from bytes",
				"key(height/hash)", key,
				"err", err,
			)
			return false
		}

		trustedEv, err := models.LightClientAttackEvidenceFromProto(&trustedPb)
		if err != nil {
			evpool.logger.Error(
				"failed to convert light client attack evidence from protobuf",
				"key(height/hash)", key,
				"err", err,
			)
			return false
		}

		// Ensure that all the byzantine validators that the evidence pool has match
		// the byzantine validators in this evidence.
		if trustedEv.ByzantineValidators == nil && lcae.ByzantineValidators != nil {
			return false
		}

		if len(trustedEv.ByzantineValidators) != len(lcae.ByzantineValidators) {
			return false
		}

		byzValsCopy := make([]*models.Validator, len(lcae.ByzantineValidators))
		for i, v := range lcae.ByzantineValidators {
			byzValsCopy[i] = v.Copy()
		}

		// ensure that both validator arrays are in the same order
		sort.Sort(models.ValidatorsByVotingPower(byzValsCopy))

		for idx, val := range trustedEv.ByzantineValidators {
			if !bytes.Equal(byzValsCopy[idx].Address, val.Address) {
				return false
			}
			if byzValsCopy[idx].VotingPower != val.VotingPower {
				return false
			}
		}

		return true
	}

	// For all other evidence the evidence pool just checks if it is already in
	// the pending db.
	return evpool.isPending(ev)
}

func (evpool *Pool) markEvidenceAsCommitted(evidence models.EvidenceList, height int64) {
	blockEvidenceMap := make(map[string]struct{}, len(evidence))
	batch := evpool.evidenceStore.NewBatch()
	defer batch.Close()

	for _, ev := range evidence {
		if evpool.isPending(ev) {
			if err := batch.Delete(keyPending(ev)); err != nil {
				evpool.logger.Error("failed to batch pending evidence", "err", err)
			}
			blockEvidenceMap[evMapKey(ev)] = struct{}{}
		}

		// Add evidence to the committed list. As the evidence is stored in the block store
		// we only need to record the height that it was saved at.
		key := keyCommitted(ev)

		h := gogotypes.Int64Value{Value: height}
		evBytes, err := commonutils.Marshal(&h)
		if err != nil {
			evpool.logger.Error("failed to marshal committed evidence", "key(height/hash)", key, "err", err)
			continue
		}

		if err := evpool.evidenceStore.Set(key, evBytes); err != nil {
			evpool.logger.Error("failed to save committed evidence", "key(height/hash)", key, "err", err)
		}

		evpool.logger.Debug("marked evidence as committed", "evidence", ev)
	}

	// check if we need to remove any pending evidence
	if len(blockEvidenceMap) == 0 {
		return
	}

	// remove committed evidence from pending bucket
	if err := batch.WriteSync(); err != nil {
		evpool.logger.Error("failed to batch delete pending evidence", "err", err)
		return
	}

	// remove committed evidence from the clist
	evpool.removeEvidenceFromList(blockEvidenceMap)

	// update the evidence size
	atomic.AddUint32(&evpool.evidenceSize, ^uint32(len(blockEvidenceMap)-1))
}
func (evpool *Pool) Size() uint32 {
	return atomic.LoadUint32(&evpool.evidenceSize)
}
func (evpool *Pool) listEvidence(prefixKey int64, maxBytes int64) ([]transfer.Evidence, int64, error) {
	var (
		evSize    int64
		totalSize int64
		evidence  []transfer.Evidence
		evList    types.EvidenceListProto // used for calculating the bytes size
	)

	iter, err := services.IteratePrefix(evpool.evidenceStore, prefixToBytes(prefixKey))
	if err != nil {
		return nil, totalSize, fmt.Errorf("database error: %v", err)
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var evpb types.EvidenceProto

		if err := commonutils.UnMarshal(iter.Value(), &evpb); nil != err {
			return evidence, totalSize, err
		}

		evList.Evidence = append(evList.Evidence, &evpb)
		evSize = int64(evList.XXX_Size())

		if maxBytes != -1 && evSize > maxBytes {
			if err := iter.Error(); err != nil {
				return evidence, totalSize, err
			}
			return evidence, totalSize, nil
		}
		ev, err := models.EvidenceFromProto(&evpb)
		if err != nil {
			return nil, totalSize, err
		}

		totalSize = evSize
		evidence = append(evidence, ev)
	}

	if err := iter.Error(); err != nil {
		return evidence, totalSize, err
	}

	return evidence, totalSize, nil
}
func (evpool *Pool) removeExpiredPendingEvidence() (int64, time.Time) {
	batch := evpool.evidenceStore.NewBatch()
	defer batch.Close()

	height, time, blockEvidenceMap := evpool.batchExpiredPendingEvidence(batch)

	// if we haven't removed any evidence then return early
	if len(blockEvidenceMap) == 0 {
		return height, time
	}

	evpool.logger.Debug("removing expired evidence",
		"height", evpool.State().LastBlockHeight,
		"time", evpool.State().LastBlockTime,
		"expired evidence", len(blockEvidenceMap),
	)

	// remove expired evidence from pending bucket
	if err := batch.WriteSync(); err != nil {
		evpool.logger.Error("failed to batch delete pending evidence", "err", err)
		return int64(evpool.State().LastBlockHeight), evpool.State().LastBlockTime
	}

	// remove evidence from the clist
	evpool.removeEvidenceFromList(blockEvidenceMap)

	// update the evidence size
	atomic.AddUint32(&evpool.evidenceSize, ^uint32(len(blockEvidenceMap)-1))

	return height, time
}

func (evpool *Pool) batchExpiredPendingEvidence(batch services.Batch) (int64, time.Time, map[string]struct{}) {
	blockEvidenceMap := make(map[string]struct{})
	iter, err := services.IteratePrefix(evpool.evidenceStore, prefixToBytes(prefixPending))
	if err != nil {
		evpool.logger.Error("failed to iterate over pending evidence", "err", err)
		return int64(evpool.State().LastBlockHeight), evpool.State().LastBlockTime, blockEvidenceMap
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		ev, err := bytesToEv(iter.Value())
		if err != nil {
			evpool.logger.Error("failed to transition evidence from protobuf", "err", err, "ev", ev)
			continue
		}

		// if true, we have looped through all expired evidence
		if !evpool.isExpired(ev.Height(), ev.Time()) {
			// Return the height and time with which this evidence will have expired
			// so we know when to prune next.
			return ev.Height() + evpool.State().ConsensusParams.Evidence.MaxAgeNumBlocks + 1,
				ev.Time().Add(evpool.State().ConsensusParams.Evidence.MaxAgeDuration).Add(time.Second),
				blockEvidenceMap
		}

		// else add to the batch
		if err := batch.Delete(iter.Key()); err != nil {
			evpool.logger.Error("failed to batch evidence", "err", err, "ev", ev)
			continue
		}

		// and add to the map to remove the evidence from the clist
		blockEvidenceMap[evMapKey(ev)] = struct{}{}
	}

	return int64(evpool.State().LastBlockHeight), evpool.State().LastBlockTime, blockEvidenceMap
}

func (evpool *Pool) State() models.LatestState {
	evpool.Lock()
	defer evpool.Unlock()
	return evpool.state
}

func prefixToBytes(prefix int64) []byte {
	key, err := orderedcode.Append(nil, prefix)
	if err != nil {
		panic(err)
	}
	return key
}

func (evpool *Pool) removeEvidenceFromList(
	blockEvidenceMap map[string]struct{}) {

	for e := evpool.evidenceList.Front(); e != nil; e = e.Next() {
		// Remove from clist
		ev := e.Value.(transfer.Evidence)
		if _, ok := blockEvidenceMap[evMapKey(ev)]; ok {
			evpool.evidenceList.Remove(e)
			e.DetachPrev()
		}
	}
}

func evMapKey(ev transfer.Evidence) string {
	return string(ev.Hash())
}
func bytesToEv(evBytes []byte) (transfer.Evidence, error) {
	var evpb types.EvidenceProto
	err := commonutils.UnMarshal(evBytes, &evpb)
	if err != nil {
		return &models.DuplicateVoteEvidence{}, err
	}

	return models.EvidenceFromProto(&evpb)
}
func (evpool *Pool) isExpired(height int64, time time.Time) bool {
	var (
		params       = evpool.State().ConsensusParams.Evidence
		ageDuration  = evpool.State().LastBlockTime.Sub(time)
		ageNumBlocks = evpool.State().LastBlockHeight - height
	)
	return int64(ageNumBlocks) > params.MaxAgeNumBlocks &&
		ageDuration > params.MaxAgeDuration
}

func (evpool *Pool) processConsensusBuffer(state models.LatestState) {
	evpool.Lock()
	defer evpool.Unlock()
	for _, voteSet := range evpool.consensusBuffer {
		// Check the height of the conflicting votes and fetch the corresponding time and validator set
		// to produce the valid evidence
		var dve *models.DuplicateVoteEvidence
		switch {
		case int64(voteSet.VoteA.Height) == state.LastBlockHeight:
			dve = models.NewDuplicateVoteEvidence(
				voteSet.VoteA,
				voteSet.VoteB,
				state.LastBlockTime,
				state.LastValidators,
			)

		case int64(voteSet.VoteA.Height) < state.LastBlockHeight:
			valSet, err := evpool.stateDB.LoadValidators(int64(voteSet.VoteA.Height))
			if err != nil {
				evpool.logger.Error("failed to load validator set for conflicting votes",
					"height", voteSet.VoteA.Height, "err", err)
				continue
			}
			blockMeta := evpool.blockStore.LoadBlockMeta(voteSet.VoteA.Height)
			if blockMeta == nil {
				evpool.logger.Error("failed to load block time for conflicting votes", "height", voteSet.VoteA.Height)
				continue
			}
			dve = models.NewDuplicateVoteEvidence(
				voteSet.VoteA,
				voteSet.VoteB,
				blockMeta.Header.Time,
				valSet,
			)

		default:
			// evidence pool shouldn't expect to get votes from consensus of a height that is above the current
			// state. If this error is seen then perhaps consider keeping the votes in the buffer and retry
			// in following heights
			evpool.logger.Error("inbound duplicate votes from consensus are of a greater height than current state",
				"duplicate vote height", voteSet.VoteA.Height,
				"state.LastBlockHeight", state.LastBlockHeight)
			continue
		}

		// check if we already have this evidence
		if evpool.isPending(dve) {
			evpool.logger.Debug("evidence already pending; ignoring", "evidence", dve)
			continue
		}

		// check that the evidence is not already committed on chain
		if evpool.isCommitted(dve) {
			evpool.logger.Debug("evidence already committed; ignoring", "evidence", dve)
			continue
		}

		if err := evpool.addPendingEvidence(dve); err != nil {
			evpool.logger.Error("failed to flush evidence from consensus buffer to pending list: %w", err)
			continue
		}

		evpool.evidenceList.PushBack(dve)

		evpool.logger.Info("verified new evidence of byzantine behavior", "evidence", dve)
	}
	// reset consensus buffer
	evpool.consensusBuffer = make([]duplicateVoteSet, 0)
}
func (evpool *Pool) addPendingEvidence(ev transfer.Evidence) error {
	evpb, err := models.EvidenceToProto(ev)
	if err != nil {
		return fmt.Errorf("failed to convert to proto: %w", err)
	}

	evBytes, err := commonutils.Marshal(evpb)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	key := keyPending(ev)

	err = evpool.evidenceStore.Set(key, evBytes)
	if err != nil {
		return fmt.Errorf("failed to persist evidence: %w", err)
	}

	atomic.AddUint32(&evpool.evidenceSize, 1)
	return nil
}
func (evpool *Pool) isCommitted(evidence transfer.Evidence) bool {
	key := keyCommitted(evidence)
	ok, err := evpool.evidenceStore.Has(key)
	if err != nil {
		evpool.logger.Error("failed to find committed evidence", "err", err)
	}
	return ok
}
func (evpool *Pool) isPending(evidence transfer.Evidence) bool {
	key := keyPending(evidence)
	ok, err := evpool.evidenceStore.Has(key)
	if err != nil {
		evpool.logger.Error("failed to find pending evidence", "err", err)
	}
	return ok
}
func keyPending(evidence transfer.Evidence) []byte {
	var height int64 = evidence.Height()
	key, err := orderedcode.Append(nil, prefixPending, height, string(evidence.Hash()))
	if err != nil {
		panic(err)
	}
	return key
}

func (evpool *Pool) updateState(state models.LatestState) {
	evpool.Lock()
	defer evpool.Unlock()
	evpool.state = state
}
func keyCommitted(evidence transfer.Evidence) []byte {
	var height int64 = evidence.Height()
	key, err := orderedcode.Append(nil, prefixCommitted, height, string(evidence.Hash()))
	if err != nil {
		panic(err)
	}
	return key
}

func (evpool *Pool) verify(evidence transfer.Evidence) error {
	var (
		state          = evpool.State()
		height         = state.LastBlockHeight
		evidenceParams = state.ConsensusParams.Evidence
		ageNumBlocks   = int64(height) - evidence.Height()
	)

	// ensure we have the block for the evidence height
	//
	// NOTE: It is currently possible for a peer to send us evidence we're not
	// able to process because we're too far behind (e.g. syncing), so we DO NOT
	// return an invalid evidence error because we do not want the peer to
	// disconnect or signal an error in this particular case.
	blockMeta := evpool.blockStore.LoadBlockMeta(uint64(evidence.Height()))
	if blockMeta == nil {
		return fmt.Errorf("failed to verify evidence; missing block for height %d", evidence.Height())
	}

	// verify the time of the evidence
	evTime := blockMeta.Header.Time
	if evidence.Time() != evTime {
		return commonerrors.NewErrInvalidEvidence(
			evidence,
			fmt.Errorf(
				"evidence has a different time to the block it is associated with (%v != %v)",
				evidence.Time(), evTime,
			),
		)
	}

	ageDuration := state.LastBlockTime.Sub(evTime)

	// check that the evidence hasn't expired
	if ageDuration > evidenceParams.MaxAgeDuration && ageNumBlocks > evidenceParams.MaxAgeNumBlocks {
		return commonerrors.NewErrInvalidEvidence(
			evidence,
			fmt.Errorf(
				"evidence from height %d (created at: %v) is too old; min height is %d and evidence can not be older than %v",
				evidence.Height(),
				evTime,
				int64(height)-evidenceParams.MaxAgeNumBlocks,
				state.LastBlockTime.Add(evidenceParams.MaxAgeDuration),
			),
		)
	}

	// apply the evidence-specific verification logic
	switch ev := evidence.(type) {
	case *models.DuplicateVoteEvidence:
		valSet, err := evpool.stateDB.LoadValidators(evidence.Height())
		if err != nil {
			return err
		}

		if err := VerifyDuplicateVote(ev, state.ChainID, valSet); err != nil {
			return commonerrors.NewErrInvalidEvidence(evidence, err)
		}

		return nil

	case *models.LightClientAttackEvidence:
		commonHeader, err := getSignedHeader(evpool.blockStore, evidence.Height())
		if err != nil {
			return err
		}

		commonVals, err := evpool.stateDB.LoadValidators(evidence.Height())
		if err != nil {
			return err
		}

		trustedHeader := commonHeader

		// in the case of lunatic the trusted header is different to the common header
		if evidence.Height() != ev.ConflictingBlock.Height {
			trustedHeader, err = getSignedHeader(evpool.blockStore, ev.ConflictingBlock.Height)
			if err != nil {
				return err
			}
		}

		err = VerifyLightClientAttack(
			ev,
			commonHeader,
			trustedHeader,
			commonVals,
			state.LastBlockTime,
			state.ConsensusParams.Evidence.MaxAgeDuration,
		)
		if err != nil {
			return commonerrors.NewErrInvalidEvidence(evidence, err)
		}

		// Find out what type of attack this was and thus extract the malicious
		// validators. Note, in the case of an Amnesia attack we don't have any
		// malicious validators.
		validators := ev.GetByzantineValidators(commonVals, trustedHeader)

		// Ensure this matches the validators that are listed in the evidence. They
		// should be ordered based on power.
		if validators == nil && ev.ByzantineValidators != nil {
			return commonerrors.NewErrInvalidEvidence(
				evidence,
				fmt.Errorf(
					"expected nil validators from an amnesia light client attack but got %d",
					len(ev.ByzantineValidators),
				),
			)
		}

		if exp, got := len(validators), len(ev.ByzantineValidators); exp != got {
			return commonerrors.NewErrInvalidEvidence(
				evidence,
				fmt.Errorf("expected %d byzantine validators from evidence but got %d", exp, got),
			)
		}

		// ensure that both validator arrays are in the same order
		sort.Sort(models.ValidatorsByVotingPower(ev.ByzantineValidators))

		for idx, val := range validators {
			if !bytes.Equal(ev.ByzantineValidators[idx].Address, val.Address) {
				return commonerrors.NewErrInvalidEvidence(
					evidence,
					fmt.Errorf(
						"evidence contained an unexpected byzantine validator address; expected: %v, got: %v",
						val.Address, ev.ByzantineValidators[idx].Address,
					),
				)
			}

			if ev.ByzantineValidators[idx].VotingPower != val.VotingPower {
				return commonerrors.NewErrInvalidEvidence(
					evidence,
					fmt.Errorf(
						"evidence contained unexpected byzantine validator power; expected %d, got %d",
						val.VotingPower, ev.ByzantineValidators[idx].VotingPower,
					),
				)
			}
		}

		return nil

	default:
		return commonerrors.NewErrInvalidEvidence(evidence, fmt.Errorf("unrecognized evidence type: %T", evidence))
	}
}

func VerifyDuplicateVote(e *models.DuplicateVoteEvidence, chainID string, valSet *models.ValidatorSet) error {
	_, val := valSet.GetByAddress(e.VoteA.ValidatorAddress)
	if val == nil {
		return fmt.Errorf("address %X was not a validator at height %d", e.VoteA.ValidatorAddress, e.Height())
	}
	pubKey := val.PubKey

	// H/R/S must be the same
	if e.VoteA.Height != e.VoteB.Height ||
		e.VoteA.Round != e.VoteB.Round ||
		e.VoteA.Type != e.VoteB.Type {
		return fmt.Errorf("h/r/s does not match: %d/%d/%v vs %d/%d/%v",
			e.VoteA.Height, e.VoteA.Round, e.VoteA.Type,
			e.VoteB.Height, e.VoteB.Round, e.VoteB.Type)
	}

	// Address must be the same
	if !bytes.Equal(e.VoteA.ValidatorAddress, e.VoteB.ValidatorAddress) {
		return fmt.Errorf("validator addresses do not match: %X vs %X",
			e.VoteA.ValidatorAddress,
			e.VoteB.ValidatorAddress,
		)
	}

	// BlockIDs must be different
	if e.VoteA.BlockID.Equals(e.VoteB.BlockID) {
		return fmt.Errorf(
			"block IDs are the same (%v) - not a real duplicate vote",
			e.VoteA.BlockID,
		)
	}

	// pubkey must match address (this should already be true, sanity check)
	addr := e.VoteA.ValidatorAddress
	if !bytes.Equal(pubKey.Address(), addr) {
		return fmt.Errorf("address (%X) doesn't match pubkey (%v - %X)",
			addr, pubKey, pubKey.Address())
	}

	// validator voting power and total voting power must match
	if val.VotingPower != e.ValidatorPower {
		return fmt.Errorf("validator power from evidence and our validator set does not match (%d != %d)",
			e.ValidatorPower, val.VotingPower)
	}
	if valSet.TotalVotingPower() != e.TotalVotingPower {
		return fmt.Errorf("total voting power from the evidence and our validator set does not match (%d != %d)",
			e.TotalVotingPower, valSet.TotalVotingPower())
	}

	va := e.VoteA.ToProto()
	vb := e.VoteB.ToProto()
	// Signatures must be valid
	if !pubKey.VerifySignature(models.VoteSignBytes(chainID, va), e.VoteA.Signature) {
		return fmt.Errorf("verifying VoteA: %w", commonerrors.ErrVoteInvalidSignature)
	}
	if !pubKey.VerifySignature(models.VoteSignBytes(chainID, vb), e.VoteB.Signature) {
		return fmt.Errorf("verifying VoteB: %w", commonerrors.ErrVoteInvalidSignature)
	}

	return nil
}

func getSignedHeader(blockStore services.IBlockStore, height int64) (*models.SignedHeader, error) {
	blockMeta := blockStore.LoadBlockMeta(uint64(height))
	if blockMeta == nil {
		return nil, fmt.Errorf("don't have header at height #%d", height)
	}
	commit := blockStore.LoadBlockCommit(uint64(height))
	if commit == nil {
		return nil, fmt.Errorf("don't have commit at height #%d", height)
	}
	return &models.SignedHeader{
		TendermintBlockHeader: &blockMeta.Header,
		Commit:                commit,
	}, nil
}

func VerifyLightClientAttack(e *models.LightClientAttackEvidence, commonHeader, trustedHeader *models.SignedHeader,
	commonVals *models.ValidatorSet, now time.Time, trustPeriod time.Duration) error {
	// In the case of lunatic attack we need to perform a single verification jump between the
	// common header and the conflicting one
	if commonHeader.Height != trustedHeader.Height {

		err := light.Verify(commonHeader, commonVals, e.ConflictingBlock.SignedHeader, e.ConflictingBlock.ValidatorSet,
			trustPeriod, now, 0*time.Second, light.DefaultTrustLevel)
		if err != nil {
			return fmt.Errorf("skipping verification from common to conflicting header failed: %w", err)
		}
	} else {
		// in the case of equivocation and amnesia we expect some header hashes to be correctly derived
		if isInvalidHeader(trustedHeader.TendermintBlockHeader, e.ConflictingBlock.TendermintBlockHeader) {
			return errors.New("common height is the same as conflicting block height so expected the conflicting" +
				" block to be correctly derived yet it wasn't")
		}
		// ensure that 2/3 of the validator set did vote for this block
		if err := e.ConflictingBlock.ValidatorSet.VerifyCommitLight(trustedHeader.ChainID, e.ConflictingBlock.Commit.BlockID,
			e.ConflictingBlock.Height, e.ConflictingBlock.Commit); err != nil {
			return fmt.Errorf("invalid commit from conflicting block: %w", err)
		}
	}

	if evTotal, valsTotal := e.TotalVotingPower, commonVals.TotalVotingPower(); evTotal != valsTotal {
		return fmt.Errorf("total voting power from the evidence and our validator set does not match (%d != %d)",
			evTotal, valsTotal)
	}

	if bytes.Equal(trustedHeader.Hash(), e.ConflictingBlock.Hash()) {
		return fmt.Errorf("trusted header hash matches the evidence's conflicting header hash: %X",
			trustedHeader.Hash())
	}

	return nil
}
