/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 3:06 下午
# @File : state.go
# @Description :
# @Attention :
*/
package state

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services4 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/ticker"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/wal"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	models2 "github.com/hyperledger/fabric-droplib/component/event/models"
	services3 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

var (
	_ services.IStateComponent = (*State)(nil)
)

var (
	ErrInvalidProposalSignature   = errors.New("error invalid proposal signature")
	ErrInvalidProposalPOLRound    = errors.New("error invalid proposal POL round")
	ErrAddingVote                 = errors.New("error adding vote")
	ErrSignatureFoundInPastBlocks = errors.New("found signature from the same key")
	errPubKeyIsNotSet             = errors.New("pubkey is not set. Look for \"Can't get private validator pubkey\" errors")
)

// interface to the evidence pool
type evidencePool interface {
	// reports conflicting votes to the evidence pool to be processed into evidence
	ReportConflictingVotes(voteA, voteB *models.Vote)
}

// interface to the mempool
type txNotifier interface {
	TxsAvailable() <-chan struct{}
}
type StateOption func(*State)

type State struct {
	*impl.BaseServiceImpl
	sync.RWMutex
	models.RoundState

	config *config.ConsensusConfiguration

	state models.LatestState

	signer    cryptolibs.ISigner
	publicKey cryptolibs.IPublicKey

	statsMsgQueue    chan models.MsgInfo
	peerMsgQueue     chan models.MsgInfo
	internalMsgQueue chan models.MsgInfo

	blockStore    services.IBlockStore
	blockExecutor services.IBlockExecutor

	wal          services.IWAL
	doWALCatchup bool // determines if we even try to do the catchup
	replayMode   bool

	// test
	nSteps int

	decideProposal func(height uint64, round int32)
	setProposal    func(proposal *models.Proposal, extraData []byte) error
	doPrevote      func(height uint64, round int32)

	eventBus     services.IConsensusEventBusComponent
	eventManager base.IEventComponent

	// tx
	txNotifier txNotifier
	evpool     evidencePool

	tikToker services.ITimeoutTicker

	done chan struct{}
}

var (
	msgQueueSize = 1000
)

func NewState(config *config.ConsensusConfiguration,
	state models.LatestState,
	blockExec services.IBlockExecutor,
	signer cryptolibs.ISigner,
	blockStore services.IBlockStore,
	notifier txNotifier,
	evPool evidencePool,
	eventComponent base.IEventComponent,
	eventBus services.IConsensusEventBusComponent,
	options ...StateOption) *State {

	cs := &State{
		config:           config,
		state:            state,
		signer:           signer,
		publicKey:        signer.GetPublicKey(),
		statsMsgQueue:    make(chan models.MsgInfo, msgQueueSize),
		peerMsgQueue:     make(chan models.MsgInfo, msgQueueSize),
		internalMsgQueue: make(chan models.MsgInfo, msgQueueSize),
		blockStore:       blockStore,
		blockExecutor:    blockExec,
		wal:              wal.NilWAL{},
		doWALCatchup:     true,
		eventBus:         eventBus,
		eventManager:     eventComponent,
		txNotifier:       notifier,
		evpool:           evPool,
		tikToker:         ticker.NewTimeoutTicker(),
		done:             make(chan struct{}),
	}

	// set function defaults (may be overwritten before calling Start)
	cs.decideProposal = cs.defaultDecideProposal
	cs.doPrevote = cs.defaultDoPrevote
	cs.setProposal = cs.defaultSetProposal

	// We have no votes, so reconstruct LastCommit from SeenCommit.
	if state.LastBlockHeight > 0 {
		cs.reconstructLastCommit(state)
	}

	cs.updateToState(state)

	// Don't call scheduleRound0 yet.
	// We do that upon Start().

	cs.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_STATE, cs)
	for _, option := range options {
		option(cs)
	}

	return cs
}
func (cs *State) reconstructLastCommit(state models.LatestState) {
	seenCommit := cs.blockStore.LoadSeenCommit(uint64(state.LastBlockHeight))
	if seenCommit == nil {
		panic(fmt.Sprintf(
			"failed to reconstruct last commit; seen commit for height %v not found",
			state.LastBlockHeight,
		))
	}

	lastPrecommits := models.CommitToVoteSet(state.ChainID, seenCommit, state.LastValidators)
	if !lastPrecommits.HasTwoThirdsMajority() {
		panic("failed to reconstruct last commit; does not have +2/3 maj")
	}

	cs.LastCommit = lastPrecommits
}
func (cs *State) defaultSetProposal(proposal *models.Proposal, extraData []byte) error {
	// Already have one
	// TODO: possibly catch double proposals
	if cs.Proposal != nil {
		return nil
	}

	// Does not apply
	if proposal.Height != cs.Height || proposal.Round != cs.Round {
		return nil
	}

	// Verify POLRound, which must be -1 or in range [0, proposal.Round).
	if proposal.POLRound < -1 ||
		(proposal.POLRound >= 0 && proposal.POLRound >= proposal.Round) {
		return ErrInvalidProposalPOLRound
	}

	p := proposal.ToProto()
	// Verify signature
	if !cs.Validators.GetProposer().PubKey.VerifySignature(
		models.ProposalSignBytes(cs.state.ChainID, p), proposal.Signature,
	) {
		return ErrInvalidProposalSignature
	}

	proposal.Signature = p.Signature
	cs.Proposal = proposal
	// We don't update cs.ProposalBlockParts if it is already set.
	// This happens if we're already in cstypes.RoundStepCommit or if there is a valid block in the current round.
	// TODO: We can check if Proposal is for a different block as this is a sign of misbehavior!
	if cs.ProposalBlockParts == nil {
		cs.ProposalBlockParts = models.NewPartSetFromHeader(proposal.BlockID.PartSetHeader)
	}

	cs.Logger.Info("received proposal", "proposal", proposal)
	return nil
}
func (cs *State) defaultDoPrevote(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	// If a block is locked, prevote that.
	if cs.LockedBlock != nil {
		logger.Debug("prevote step; already locked on a block; prevoting locked block")
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, cs.LockedBlock.Hash(), cs.LockedBlockParts.Header())
		return
	}

	// If ProposalBlock is nil, prevote nil.
	if cs.ProposalBlock == nil {
		logger.Debug("prevote step: ProposalBlock is nil")
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, nil, models.PartSetHeader{})
		return
	}

	// Validate proposal block
	err := cs.blockExecutor.ValidateBlock(cs.state, cs.ProposalBlock)
	if err != nil {
		// ProposalBlock is invalid, prevote nil.
		logger.Error("prevote step: ProposalBlock is invalid", "err", err)
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, nil, models.PartSetHeader{})
		return
	}

	// Prevote cs.ProposalBlock
	// NOTE: the proposal signature is validated when it is received,
	// and the proposal block parts are validated as they are received (against the merkle hash in the proposal)
	logger.Debug("prevote step: ProposalBlock is valid")
	cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, cs.ProposalBlock.Hash(), cs.ProposalBlockParts.Header())
}
func (cs *State) defaultDecideProposal(height uint64, round int32) {
	var block *models.TendermintBlockWrapper
	var blockParts *models.PartSet

	// Decide on block
	// 对应的是19行,如果validValue不为空
	if cs.ValidBlock != nil {
		// If there is valid block, choose that.
		block, blockParts = cs.ValidBlock, cs.ValidBlockParts
	} else {
		// Create a new proposal block from state/txs from the mempool.
		// 否则的话从缓存中获取
		block, blockParts = cs.createProposalBlock()
		if block == nil {
			return
		}
	}

	// Flush the WAL. Otherwise, we may not recompute the same proposal to sign,
	// and the privValidator will refuse to sign anything.
	if err := cs.wal.FlushAndSync(); err != nil {
		cs.Logger.Error("Error flushing to disk")
	}

	// Make proposal
	propBlockID := models.BlockID{Hash: block.Hash(), PartSetHeader: blockParts.Header()}

	proposal := models.NewProposal(int64(height), round, cs.ValidRound, propBlockID)
	// p := proposal.ToProto()
	if sig, err := cs.signer.SignProposal(cs.state.ChainID, proposal); err == nil {
		proposal.Signature = sig
		// send proposal and block parts on internal msg queue
		cs.sendInternalMessage(models.MsgInfo{&models.ProposalMessage{Proposal: proposal}, ""})
		for i := 0; i < int(blockParts.Total()); i++ {
			part := blockParts.GetPart(i)
			cs.sendInternalMessage(models.MsgInfo{&models.BlockPartMessage{cs.Height, cs.Round, part}, ""})
		}
		cs.Logger.Info("Signed proposal", "height:", height, "round:", round, "proposal:", proposal)
		cs.Logger.Debug(fmt.Sprintf("Signed proposal block: %v", block))
	} else if !cs.replayMode {
		cs.Logger.Error("enterPropose: Error signing proposal", "height", height, "round", round, "err", err)
	}
}
func (cs *State) createProposalBlock() (block *models.TendermintBlockWrapper, blockParts *models.PartSet) {
	if cs.signer == nil {
		panic("entered createProposalBlock with privValidator being nil")
	}

	var commit *models.Commit
	switch {
	case cs.Height == cs.state.InitialHeight:
		// We're creating a proposal for the first block.
		// The commit is empty, but not nil.
		commit = models.NewCommit(0, 0, models.BlockID{}, nil)

	case cs.LastCommit.HasTwoThirdsMajority():
		// Make the commit from LastCommit
		commit = cs.LastCommit.MakeCommit()

	default: // This shouldn't happen.
		cs.Logger.Error("propose step; cannot propose anything without commit for the previous block")
		return
	}

	if cs.signer == nil {
		// If this node is a validator & proposer in the current round, it will
		// miss the opportunity to create a block.
		cs.Logger.Error("propose step; empty priv validator public key", "err", errPubKeyIsNotSet)
		return
	}

	proposerAddr := cs.publicKey.Address()

	return cs.blockExecutor.CreateProposalBlock(int64(cs.Height), cs.state, commit, proposerAddr)
}
func (cs *State) OnStart() error {
	if _, ok := cs.wal.(wal.NilWAL); ok {
		if err := cs.loadWalFile(); err != nil {
			return err
		}
	}

	if cs.doWALCatchup {
		repairAttempted := false
	LOOP:
		for {
			err := cs.catchupReplay(cs.Height)
			switch {
			case err == nil:
				break LOOP
			case !services.IsDataCorruptionError(err):
				cs.Logger.Error("Error on catchup replay. Proceeding to start State anyway", "err", err)
				break LOOP
			case repairAttempted:
				return err
			}

			cs.Logger.Error("WAL file is corrupted, attempting repair", "err", err)

			// 1) prep work
			if err := cs.wal.Stop(); err != nil {
				return err
			}
			repairAttempted = true

			// 2) backup original WAL file
			corruptedFile := fmt.Sprintf("%s.CORRUPTED", cs.config.WalFile())
			if err := libs.CopyFile(cs.config.WalFile(), corruptedFile); err != nil {
				return err
			}
			cs.Logger.Info("Backed up WAL file", "src", cs.config.WalFile(), "dst", corruptedFile)

			// 3) try to repair (WAL file will be overwritten!)
			if err := repairWalFile(corruptedFile, cs.config.WalFile()); err != nil {
				cs.Logger.Error("WAL repair failed", "err", err)
				return err
			}
			cs.Logger.Info("Successful repair")

			// reload WAL file
			if err := cs.loadWalFile(); err != nil {
				return err
			}
		}
	}

	if err := cs.eventManager.Start(services4.ASYNC_START); nil != err {
		cs.Logger.Error("启动event manager 失败:" + err.Error())
		return err
	}
	// we need the timeoutRoutine for replay so
	// we don't block on the tick chan.
	// NOTE: we will get a build up of garbage go routines
	// firing on the tockChan until the receiveRoutine is started
	// to deal with them (by that point, at most one will be valid)
	if err := cs.tikToker.Start(services4.ASYNC_START); err != nil {
		return err
	}

	go cs.receiveRoutine(0)

	return nil
}
func (cs *State) receiveRoutine(maxSteps int) {
	onExit := func(cs *State) {
		// NOTE: the internalMsgQueue may have signed messages from our
		// priv_val that haven't hit the WAL, but its ok because
		// priv_val tracks LastSig

		// close wal now that we're done writing to it
		if err := cs.wal.Stop(); err != nil {
			cs.Logger.Error("error trying to stop wal", "error", err)
		}
		cs.wal.Wait()

		close(cs.done)
	}

	defer func() {
		if r := recover(); r != nil {
			cs.Logger.Error("CONSENSUS FAILURE!!!", "err", r, "stack", string(debug.Stack()))
			// stop gracefully
			//
			// NOTE: We most probably shouldn't be running any further when there is
			// some unexpected panic. Some unknown error happened, and so we don't
			// know if that will result in the validator signing an invalid thing. It
			// might be worthwhile to explore a mechanism for manual resuming via
			// some console or secure RPC system, but for now, halting the chain upon
			// unexpected consensus bugs sounds like the better option.
			onExit(cs)
		}
	}()

	for {
		if maxSteps > 0 {
			if cs.nSteps >= maxSteps {
				cs.Logger.Info("reached max steps. exiting receive routine")
				cs.nSteps = 0
				return
			}
		}
		rs := cs.RoundState
		var mi models.MsgInfo

		select {
		case <-cs.txNotifier.TxsAvailable():
			cs.handleTxsAvailable()
		case mi = <-cs.peerMsgQueue:
			if err := cs.wal.Write(mi); err != nil {
				cs.Logger.Error("Error writing to wal", "err", err)
			}
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			cs.handleMsg(mi)
		case mi = <-cs.internalMsgQueue:
			err := cs.wal.WriteSync(mi) // NOTE: fsync
			if err != nil {
				panic(fmt.Sprintf("Failed to write %v msg to consensus wal due to %v. Check your FS and restart the node", mi, err))
			}

			if _, ok := mi.Msg.(*models.VoteMessage); ok {
				// we actually want to simulate failing during
				// the previous WriteSync, but this isn't easy to do.
				// Equivalent would be to fail here and manually remove
				// some bytes from the end of the wal.

				// fail.Fail() // XXX
			}

			// handles proposals, block parts, votes
			cs.handleMsg(mi)
		case ti := <-cs.tikToker.Chan(): // tockChan:
			// 保存日志记录
			if err := cs.wal.Write(ti); err != nil {
				cs.Logger.Error("Error writing to wal", "err", err)
			}
			// if the timeout is relevant to the rs
			// go to the next step
			cs.handleTimeout(ti, rs)
		case <-cs.Quit():
			onExit(cs)
			return
		}
	}
}
func (cs *State) handleTxsAvailable() {
	cs.Lock()
	defer cs.Unlock()

	// We only need to do this for round 0.
	if cs.Round != 0 {
		return
	}

	switch cs.Step {
	case types.RoundStepNewHeight: // timeoutCommit phase
		if cs.needProofBlock(cs.Height) {
			// enterPropose will be called by enterNewRound
			return
		}

		// +1ms to ensure RoundStepNewRound timeout always happens after RoundStepNewHeight
		timeoutCommit := cs.StartTime.Sub(libs.Now()) + 1*time.Millisecond
		cs.scheduleTimeout(timeoutCommit, cs.Height, 0, types.RoundStepNewRound)
	case types.RoundStepNewRound: // after timeoutCommit
		cs.enterPropose(cs.Height, 0)
	}
}
func repairWalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	var (
		dec = services.NewWALDecoder(in)
		enc = services.NewWALEncoder(out)
	)

	// best-case repair (until first error is encountered)
	for {
		msg, err := dec.Decode()
		if err != nil {
			break
		}

		err = enc.Encode(msg)
		if err != nil {
			return fmt.Errorf("failed to encode msg: %w", err)
		}
	}

	return nil
}
func (cs *State) catchupReplay(csHeight uint64) error {

	// Set replayMode to true so we don't log signing errors.
	cs.replayMode = true
	defer func() { cs.replayMode = false }()

	// Ensure that #ENDHEIGHT for this height doesn't exist.
	// NOTE: This is just a sanity check. As far as we know things work fine
	// without it, and Handshake could reuse State if it weren't for
	// this check (since we can crash after writing #ENDHEIGHT).
	//
	// Ignore data corruption errors since this is a sanity check.
	gr, found, err := cs.wal.SearchForEndHeight(csHeight, &services.WALSearchOptions{IgnoreDataCorruptionErrors: true})
	if err != nil {
		return err
	}
	if gr != nil {
		if err := gr.Close(); err != nil {
			return err
		}
	}
	if found {
		return fmt.Errorf("wal should not contain #ENDHEIGHT %d", csHeight)
	}

	// Search for last height marker.
	//
	// Ignore data corruption errors in previous heights because we only care about last height
	if csHeight < cs.state.InitialHeight {
		return fmt.Errorf("cannot replay height %v, below initial height %v", csHeight, cs.state.InitialHeight)
	}
	endHeight := csHeight - 1
	if csHeight == cs.state.InitialHeight {
		endHeight = 0
	}
	gr, found, err = cs.wal.SearchForEndHeight(endHeight, &services.WALSearchOptions{IgnoreDataCorruptionErrors: true})
	if err == io.EOF {
		cs.Logger.Error("Replay: wal.group.Search returned EOF", "#ENDHEIGHT", endHeight)
	} else if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("cannot replay height %d. WAL does not contain #ENDHEIGHT for %d", csHeight, endHeight)
	}
	defer gr.Close()

	cs.Logger.Info("Catchup by replaying consensus messages", "height", csHeight)

	var msg *services.TimedWALMessage
	dec := services.WALDecoder{gr}

LOOP:
	for {
		msg, err = dec.Decode()
		switch {
		case err == io.EOF:
			break LOOP
		case services.IsDataCorruptionError(err):
			cs.Logger.Error("data has been corrupted in last height of consensus WAL", "err", err, "height", csHeight)
			return err
		case err != nil:
			return err
		}

		// NOTE: since the priv key is set when the msgs are received
		// it will attempt to eg double sign but we can just ignore it
		// since the votes will be replayed and we'll get to the next step
		if err := cs.readReplayMessage(msg, nil); err != nil {
			return err
		}
	}
	cs.Logger.Info("Replay: Done")
	return nil
}
func (cs *State) readReplayMessage(msg *services.TimedWALMessage, newStepSub services3.Subscription) error {
	// Skip meta messages which exist for demarcating boundaries.
	if _, ok := msg.Msg.(services.EndHeightMessage); ok {
		return nil
	}

	// for logging
	switch m := msg.Msg.(type) {
	case events.EventDataRoundState:
		cs.Logger.Info("Replay: New Step", "height", m.Height, "round", m.Round, "step", m.Step)
		// these are playback checks
		ticker := time.After(time.Second * 2)
		if newStepSub != nil {
			select {
			case stepMsg := <-newStepSub.Out():
				m2 := stepMsg.Data().(events.EventDataRoundState)
				if m.Height != m2.Height || m.Round != m2.Round || m.Step != m2.Step {
					return fmt.Errorf("roundState mismatch. Got %v; Expected %v", m2, m)
				}
			case <-newStepSub.Canceled():
				return fmt.Errorf("failed to read off newStepSub.Out(). newStepSub was canceled")
			case <-ticker:
				return fmt.Errorf("failed to read off newStepSub.Out()")
			}
		}
	case models.MsgInfo:
		peerID := m.PeerID
		if peerID == "" {
			peerID = "local"
		}
		switch msg := m.Msg.(type) {
		case *models.ProposalMessage:
			p := msg.Proposal
			cs.Logger.Info("Replay: Proposal", "height", p.Height, "round", p.Round, "header",
				p.BlockID.PartSetHeader, "pol", p.POLRound, "peer", peerID)
		case *models.BlockPartMessage:
			cs.Logger.Info("Replay: BlockPart", "height", msg.Height, "round", msg.Round, "peer", peerID)
		case *models.VoteMessage:
			v := msg.Vote
			cs.Logger.Info("Replay: Vote", "height", v.Height, "round", v.Round, "type", v.Type,
				"blockID", v.BlockID, "peer", peerID)
		}

		cs.handleMsg(m)
	case models.TimeoutInfo:
		cs.Logger.Info("Replay: Timeout", "height", m.Height, "round", m.Round, "step", m.Step, "dur", m.Duration)
		cs.handleTimeout(m, cs.RoundState)
	default:
		return fmt.Errorf("replay: Unknown TimedWALMessage type: %v", reflect.TypeOf(msg.Msg))
	}
	return nil
}
func (cs *State) handleTimeout(ti models.TimeoutInfo, rs models.RoundState) {
	cs.Logger.Debug("Received tock", "timeout", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)

	// timeouts must be for current height, round, step
	if ti.Height != rs.Height || ti.Round < rs.Round || (ti.Round == rs.Round && ti.Step < rs.Step) {
		cs.Logger.Debug("Ignoring tock because we're ahead", "height", rs.Height, "round", rs.Round, "step", rs.Step)
		return
	}

	// the timeout will now cause a state transition
	cs.Lock()
	defer cs.Unlock()

	switch ti.Step {
	case types.RoundStepNewHeight:
		// NewRound event fired from enterNewRound.
		// XXX: should we fire timeout here (for timeout commit)?
		cs.enterNewRound(ti.Height, 0)
	case types.RoundStepNewRound:
		cs.enterPropose(ti.Height, 0)
	case types.RoundStepPropose:
		if err := cs.eventBus.PublishEventTimeoutPropose(events.RoundStateEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing timeout propose", "err", err)
		}
		cs.enterPrevote(ti.Height, ti.Round)
	case types.RoundStepPrevoteWait:
		if err := cs.eventBus.PublishEventTimeoutWait(events.RoundStateEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing timeout wait", "err", err)
		}
		cs.enterPrecommit(ti.Height, ti.Round)
	case types.RoundStepPrecommitWait:
		if err := cs.eventBus.PublishEventTimeoutWait(events.RoundStateEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing timeout wait", "err", err)
		}
		cs.enterPrecommit(ti.Height, ti.Round)
		cs.enterNewRound(ti.Height, ti.Round+1)
	default:
		panic(fmt.Sprintf("Invalid timeout step: %v", ti.Step))
	}

}
func (cs *State) enterPrevote(height uint64, round int32) {
	if cs.Height != height || round < cs.Round || (cs.Round == round && types.RoundStepPrevote <= cs.Step) {
		cs.Logger.Debug(fmt.Sprintf(
			"enterPrevote(%v/%v): Invalid args. Current step: %v/%v/%v",
			height,
			round,
			cs.Height,
			cs.Round,
			cs.Step))
		return
	}

	defer func() {
		// Done enterPrevote:
		cs.updateRoundStep(round, types.RoundStepPrevote)
		cs.newStep()
	}()

	cs.Logger.Debug(fmt.Sprintf("enterPrevote(%v/%v); current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	// Sign and broadcast vote as necessary
	cs.doPrevote(height, round)

	// Once `addVote` hits any +2/3 prevotes, we will go to PrevoteWait
	// (so we have more time to try and collect +2/3 prevotes for a single block)
}
func (cs *State) newStep() {
	rs := events.RoundStateEvent(cs.RoundState)
	if err := cs.wal.Write(rs); err != nil {
		cs.Logger.Error("Error writing to wal", "err", err)
	}
	cs.nSteps++
	// newStep is called by updateToState in NewState before the eventBus is set!
	if cs.eventBus != nil {
		if err := cs.eventBus.PublishEventNewRoundStep(rs); err != nil {
			cs.Logger.Error("Error publishing new round step", "err", err)
		}
		cs.eventManager.Fire(models2.DefaultEventWrapper{
			NameSpace: types.EventNewRoundStep,
			EventData: &cs.RoundState,
		})
	}
}
func (cs *State) enterPropose(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	if cs.Height != height || round < cs.Round || (cs.Round == round && types.RoundStepPropose <= cs.Step) {
		logger.Debug(fmt.Sprintf(" 试图开始proposal ,enterPropose(%v/%v): 参数错误. 当前step信息为: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}
	logger.Info(fmt.Sprintf("开始proposal:enterPropose(%v/%v). 当前step信息为: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPropose:
		cs.updateRoundStep(round, types.RoundStepPropose)
		cs.newStep()

		// If we have the whole proposal + POL, then goto Prevote now.
		// else, we'll enterPrevote when the rest of the proposal is received (in AddProposalBlockPart),
		// or else after timeoutPropose
		if cs.isProposalComplete() {
			cs.enterPrevote(height, cs.Round)
		}
	}()

	// If we don't get the proposal and all block parts quick enough, enterPrevote
	cs.scheduleTimeout(cs.config.Propose(round), height, round, types.RoundStepPropose)

	// Nothing more to do if we're not a validator
	if cs.signer == nil {
		logger.Debug("This node is not a validator")
		return
	}
	logger.Debug("This node is a validator")

	if cs.publicKey == nil {
		// If this node is a validator & proposer in the current round, it will
		// miss the opportunity to create a block.
		logger.Error(fmt.Sprintf("enterPropose: %v", errPubKeyIsNotSet))
		return
	}
	address := cs.publicKey.Address()

	// if not a validator, we're done
	if !cs.Validators.HasAddress(address) {
		logger.Debug("This node is not a validator", "addr", address, "vals", cs.Validators)
		return
	}

	if cs.isProposer(address) {
		logger.Debug("enterPropose: our turn to propose",
			"proposer", address,
			"privValidator", cs.signer,
		)
		cs.decideProposal(height, round)
	} else {
		logger.Debug("enterPropose: not our turn to propose",
			"proposer", cs.Validators.GetProposer().Address,
			"privValidator", cs.signer,
		)
	}
}
func (cs *State) isProposer(address []byte) bool {
	return bytes.Equal(cs.Validators.GetProposer().Address, address)
}
func (cs *State) isProposalComplete() bool {
	if cs.Proposal == nil || cs.ProposalBlock == nil {
		return false
	}
	// we have the proposal. if there's a POLRound,
	// make sure we have the prevotes from it too
	if cs.Proposal.POLRound < 0 {
		return true
	}
	// if this is false the proposer is lying or we haven't received the POL yet
	return cs.Votes.Prevotes(cs.Proposal.POLRound).HasTwoThirdsMajority()

}
func (cs *State) enterNewRound(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	if cs.Height != height || round < cs.Round || (cs.Round == round && cs.Step != types.RoundStepNewHeight) {
		logger.Debug(fmt.Sprintf(
			"试图 enterNewRound(%v/%v):  参数错误,当前step相关信息为: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))
		return
	}

	if now := libs.Now(); cs.StartTime.After(now) {
		logger.Debug("need to set a buffer and log message here for sanity", "startTime", cs.StartTime, "now", now)
	}

	logger.Info(fmt.Sprintf("开始新的一轮:(%v/%v). 当前step信息为: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	// Increment validators if necessary
	validators := cs.Validators
	if cs.Round < round {
		validators = validators.Copy()
		validators.IncrementProposerPriority(libs.SafeSubInt32(round, cs.Round))
	}

	// Setup new round
	// we don't fire newStep for this step,
	// but we fire an event, so update the round step first
	cs.updateRoundStep(round, types.RoundStepNewRound)
	cs.Validators = validators
	if round == 0 {
		// We've already reset these upon new height,
		// and meanwhile we might have received a proposal
		// for round 0.
	} else {
		logger.Info("Resetting Proposal info")
		cs.Proposal = nil
		cs.ProposalBlock = nil
		cs.ProposalBlockParts = nil
	}
	cs.Votes.SetRound(libs.SafeAddInt32(round, 1)) // also track next round (round+1) to allow round-skipping
	cs.TriggeredTimeoutPrecommit = false

	if err := cs.eventBus.PublishEventNewRound(events.NewRoundEvent(cs.RoundState)); err != nil {
		cs.Logger.Error("Error publishing new round", "err", err)
	}
	// cs.metrics.Rounds.Set(float64(round))

	// Wait for txs to be available in the mempool
	// before we enterPropose in round 0. If the last block changed the app hash,
	// we may need an empty "proof" block, and enterPropose immediately.
	waitForTxs := cs.config.WaitForTxs() && round == 0 && !cs.needProofBlock(uint64(height))
	// FIXME should always true
	if waitForTxs {
		if cs.config.CreateEmptyBlocksInterval > 0 {
			cs.scheduleTimeout(cs.config.CreateEmptyBlocksInterval, height, round,
				types.RoundStepNewRound)
		}
	} else {
		cs.enterPropose(height, round)
	}
}
func (cs *State) needProofBlock(height uint64) bool {
	if height == uint64(cs.state.InitialHeight) {
		return true
	}
	// FIXME ,应该是不需要的
	lastBlockMeta := cs.blockStore.LoadBlockMeta(height - 1)
	if lastBlockMeta == nil {
		panic(fmt.Sprintf("needProofBlock: last block meta for height %d not found", height-1))
	}
	// return !bytes.Equal(cs.state.AppHash, lastBlockMeta.Header.AppHash)
	return true
}

func (cs *State) updateRoundStep(round int32, step types.RoundStepType) {
	cs.Round = round
	cs.Step = step
}

func (cs *State) handleMsg(mi models.MsgInfo) {
	cs.Lock()
	defer cs.Unlock()

	var (
		added bool
		err   error
	)
	msg, peerID := mi.Msg, mi.PeerID
	switch msg := msg.(type) {
	case *models.ProposalMessage:
		// will not cause transition.
		// once proposal is set, we can receive block parts
		err = cs.setProposal(msg.Proposal, msg.ExtraData)
	case *models.BlockPartMessage:
		// if the proposal is complete, we'll enterPrevote or tryFinalizeCommit
		added, err = cs.addProposalBlockPart(msg, peerID)
		if added {
			cs.statsMsgQueue <- mi
		}

		if err != nil && msg.Round != cs.Round {
			cs.Logger.Debug(
				"Received block part from wrong round",
				"height",
				cs.Height,
				"csRound",
				cs.Round,
				"blockRound",
				msg.Round)
			err = nil
		}
	case *models.VoteMessage:
		added, err = cs.tryAddVote(msg.Vote, peerID)
		if added {
			cs.statsMsgQueue <- mi
		}
	default:
		cs.Logger.Error("Unknown msg type", "type", reflect.TypeOf(msg))
		return
	}

	if err != nil {
		cs.Logger.Error("Error with msg", "height", cs.Height, "round", cs.Round,
			"peer", peerID, "err", err, "msg", msg)
	}
}
func (cs *State) tryAddVote(vote *models.Vote, peerID types2.NodeID) (bool, error) {
	added, err := cs.addVote(vote, peerID)
	if err != nil {
		// If the vote height is off, we'll just ignore it,
		// But if it's a conflicting sig, add it to the cs.evpool.
		// If it's otherwise invalid, punish peer.
		// nolint: gocritic
		if voteErr, ok := err.(*models.ErrVoteConflictingVotes); ok {
			if cs.signer == nil {
				return false, errPubKeyIsNotSet
			}

			if bytes.Equal(vote.ValidatorAddress, cs.publicKey.Address()) {
				cs.Logger.Error(
					"Found conflicting vote from ourselves. Did you unsafe_reset a validator?",
					"height",
					vote.Height,
					"round",
					vote.Round,
					"type",
					vote.Type)
				return added, err
			}
			// report conflicting votes to the evidence pool
			// FIXME ,双重签名还是需要的
			cs.evpool.ReportConflictingVotes(voteErr.VoteA, voteErr.VoteB)
			cs.Logger.Info("Found and sent conflicting votes to the evidence pool",
				"VoteA", voteErr.VoteA,
				"VoteB", voteErr.VoteB,
			)
			return added, err
		} else if err == commonerrors.ErrVoteNonDeterministicSignature {
			cs.Logger.Debug("Vote has non-deterministic signature", "err", err)
		} else {
			// Either
			// 1) bad peer OR
			// 2) not a bad peer? this can also err sometimes with "Unexpected step" OR
			// 3) tmkms use with multiple validators connecting to a single tmkms instance
			// 		(https://github.com/tendermint/tendermint/issues/3839).
			cs.Logger.Error("Error attempting to add vote", "err", err)
			return added, ErrAddingVote
		}
	}
	return added, nil
}
func (cs *State) addProposalBlockPart(msg *models.BlockPartMessage, peerID types2.NodeID) (added bool, err error) {
	height, round, part := msg.Height, msg.Round, msg.Part

	// Blocks might be reused, so round mismatch is OK
	if cs.Height != height {
		cs.Logger.Debug("Received block part from wrong height", "height", height, "round", round)
		return false, nil
	}

	// We're not expecting a block part.
	if cs.ProposalBlockParts == nil {
		// NOTE: this can happen when we've gone to a higher round and
		// then receive parts from the previous round - not necessarily a bad peer.
		cs.Logger.Info("Received a block part when we're not expecting any",
			"height", height, "round", round, "index", part.Index, "peer", peerID)
		return false, nil
	}

	added, err = cs.ProposalBlockParts.AddPart(part)
	if err != nil {
		return added, err
	}
	if cs.ProposalBlockParts.ByteSize() > cs.state.ConsensusParams.Block.MaxBytes {
		return added, fmt.Errorf("total size of proposal block parts exceeds maximum block bytes (%d > %d)",
			cs.ProposalBlockParts.ByteSize(), cs.state.ConsensusParams.Block.MaxBytes,
		)
	}
	if added && cs.ProposalBlockParts.IsComplete() {
		bz, err := ioutil.ReadAll(cs.ProposalBlockParts.GetReader())
		if err != nil {
			return added, err
		}

		var pbb = &types3.TendermintBlockProtoWrapper{}
		err = commonutils.UnMarshal(bz, pbb)
		if err != nil {
			return added, err
		}

		block, err := models.BlockFromProto(pbb)
		if err != nil {
			return added, err
		}

		cs.ProposalBlock = block
		// NOTE: it's possible to receive complete proposal blocks for future rounds without having the proposal
		cs.Logger.Info("Received complete proposal block", "height", cs.ProposalBlock.Height, "hash", cs.ProposalBlock.Hash())
		if err := cs.eventBus.PublishEventCompleteProposal(events.CompleteProposalEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing event complete proposal", "err", err)
		}

		// Update Valid* if we can.
		prevotes := cs.Votes.Prevotes(cs.Round)
		blockID, hasTwoThirds := prevotes.TwoThirdsMajority()
		if hasTwoThirds && !blockID.IsZero() && (cs.ValidRound < cs.Round) {
			if cs.ProposalBlock.HashesTo(blockID.Hash) {
				cs.Logger.Info("Updating valid block to new proposal block",
					"valid-round", cs.Round, "valid-block-hash", cs.ProposalBlock.Hash())
				cs.ValidRound = cs.Round
				cs.ValidBlock = cs.ProposalBlock
				cs.ValidBlockParts = cs.ProposalBlockParts
			}
			// TODO: In case there is +2/3 majority in Prevotes set for some
			// block and cs.ProposalBlock contains different block, either
			// proposer is faulty or voting power of faulty processes is more
			// than 1/3. We should trigger in the future accountability
			// procedure at this point.
		}

		if cs.Step <= types.RoundStepPropose && cs.isProposalComplete() {
			// Move onto the next step
			cs.enterPrevote(height, cs.Round)
			if hasTwoThirds { // this is optimisation as this will be triggered when prevote is added
				cs.enterPrecommit(height, cs.Round)
			}
		} else if cs.Step == types.RoundStepCommit {
			// If we're waiting on the proposal block...
			cs.tryFinalizeCommit(height)
		}
		return added, nil
	}
	return added, nil
}
func (cs *State) tryFinalizeCommit(height uint64) {
	// logger := cs.Logger.With("height", height)
	logger := cs.Logger

	if cs.Height != height {
		panic(fmt.Sprintf("tryFinalizeCommit() cs.Height: %v vs height: %v", cs.Height, height))
	}

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	if !ok || len(blockID.Hash) == 0 {
		logger.Error("Attempt to finalize failed. There was no +2/3 majority, or +2/3 was for <nil>.")
		return
	}
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		// TODO: this happens every time if we're not a validator (ugly logs)
		// TODO: ^^ wait, why does it matter that we're a validator?
		logger.Debug(
			"attempt to finalize failed; we do not have the commit block",
			"proposal-block", cs.ProposalBlock.Hash(),
			"commit-block", blockID.Hash,
		)
		return
	}

	//	go
	cs.finalizeCommit(height)
}
func (cs *State) finalizeCommit(height uint64) {
	if cs.Height != height || cs.Step != types.RoundStepCommit {
		cs.Logger.Debug(fmt.Sprintf(
			"finalizeCommit(%v): Invalid args. Current step: %v/%v/%v",
			height,
			cs.Height,
			cs.Round,
			cs.Step))
		return
	}

	blockID, ok := cs.Votes.Precommits(cs.CommitRound).TwoThirdsMajority()
	block, blockParts := cs.ProposalBlock, cs.ProposalBlockParts

	if !ok {
		panic("Cannot finalizeCommit, commit does not have two thirds majority")
	}
	if !blockParts.HasHeader(blockID.PartSetHeader) {
		panic("Expected ProposalBlockParts header to be commit header")
	}
	if !block.HashesTo(blockID.Hash) {
		panic("Cannot finalizeCommit, ProposalBlock does not hash to commit hash")
	}
	if err := cs.blockExecutor.ValidateBlock(cs.state, block); err != nil {
		panic(fmt.Errorf("+2/3 committed an invalid block: %w", err))
	}

	cs.Logger.Info("finalizing commit of block with N txs",
		"height", block.Height,
		"hash", block.Hash(),
		// "root", block.AppHash,
		"N", len(block.Txs),
	)
	cs.Logger.Debug(fmt.Sprintf("%v", block))

	libs.Fail() // XXX

	// Save to blockStore.
	h := cs.blockStore.Height()
	if h == 0 || h < uint64(block.Height) {
		// NOTE: the seenCommit is local justification to commit this block,
		// but may differ from the LastCommit included in the next block
		precommits := cs.Votes.Precommits(cs.CommitRound)
		seenCommit := precommits.MakeCommit()
		cs.blockStore.SaveBlock(block, blockParts, seenCommit)
	} else {
		// Happens during replay if we already saved the block but didn't commit
		cs.Logger.Debug("calling finalizeCommit on already stored block", "height", block.Height)
	}

	libs.Fail() // XXX

	// Write EndHeightMessage{} for this height, implying that the blockstore
	// has saved the block.
	//
	// If we crash before writing this EndHeightMessage{}, we will recover by
	// running ApplyBlock during the ABCI handshake when we restart.  If we
	// didn't save the block to the blockstore before writing
	// EndHeightMessage{}, we'd have to change WAL replay -- currently it
	// complains about replaying for heights where an #ENDHEIGHT entry already
	// exists.
	//
	// Either way, the State should not be resumed until we
	// successfully call ApplyBlock (ie. later here, or in Handshake after
	// restart).
	// 持久化区块高度
	endMsg := services.EndHeightMessage{Height: height}
	if err := cs.wal.WriteSync(endMsg); err != nil { // NOTE: fsync
		panic(fmt.Sprintf("Failed to write %v msg to consensus wal due to %v. Check your FS and restart the node",
			endMsg, err))
	}

	libs.Fail() // XXX

	// Create a copy of the state for staging and an event cache for txs.
	stateCopy := cs.state.Copy()

	// Execute and commit the block, update and save the state, and update the mempool.
	// NOTE The block.AppHash wont reflect these txs until the next block.
	var err error
	var retainHeight uint64
	stateCopy, retainHeight, err = cs.blockExecutor.ApplyBlock(
		stateCopy,
		models.BlockID{Hash: block.Hash()},
		block)
	if err != nil {
		cs.Logger.Error("Error on ApplyBlock", "err", err)
		return
	}

	libs.Fail() // XXX

	// Prune old heights, if requested by ABCI app.
	if retainHeight > 0 {
		pruned, err := cs.pruneBlocks(retainHeight)
		if err != nil {
			cs.Logger.Error("Failed to prune blocks", "retainHeight", retainHeight, "err", err)
		} else {
			cs.Logger.Info("Pruned blocks", "pruned", pruned, "retainHeight", retainHeight)
		}
	}

	// must be called before we update state
	// cs.recordMetrics(height, block)

	// NewHeightStep!
	cs.updateToState(stateCopy)

	libs.Fail() // XXX

	// Private validator might have changed it's key pair => refetch pubkey.
	if err := cs.updatePrivValidatorPubKey(); err != nil {
		cs.Logger.Error("Can't get private validator pubkey", "err", err)
	}

	// cs.StartTime is already set.
	// Schedule Round0 to start soon.
	cs.scheduleRound0(&cs.RoundState)

	// By here,
	// * cs.Height has been increment to height+1
	// * cs.Step is now cstypes.RoundStepNewHeight
	// * cs.StartTime is set to when we will start round0.
}
func (cs *State) scheduleRound0(rs *models.RoundState) {
	// cs.Logger.Info("scheduleRound0", "now", tmtime.Now(), "startTime", cs.StartTime)
	sleepDuration := rs.StartTime.Sub(libs.Now())
	cs.scheduleTimeout(sleepDuration, rs.Height, 0, types.RoundStepNewHeight)
}
func (cs *State) updatePrivValidatorPubKey() error {
	if cs.signer == nil {
		return nil
	}

	pubKey := cs.signer.GetPublicKey()
	cs.publicKey = pubKey
	return nil
}
func (cs *State) pruneBlocks(retainHeight uint64) (uint64, error) {
	base := cs.blockStore.Base()
	if uint64(retainHeight) <= base {
		return 0, nil
	}
	pruned, err := cs.blockStore.PruneBlocks(retainHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to prune block store: %w", err)
	}

	// 与fabric 结合不需要prune states
	err = cs.blockExecutor.GetStateStore().PruneStates(retainHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to prune state store: %w", err)
	}
	return pruned, nil
}
func (cs *State) updateToState(state models.LatestState) {
	if cs.CommitRound > -1 && 0 < cs.Height && int64(cs.Height) != state.LastBlockHeight {
		panic(fmt.Sprintf("updateToState() expected state height of %v but found %v",
			cs.Height, state.LastBlockHeight))
	}
	if !cs.state.IsEmpty() {
		if cs.state.LastBlockHeight > 0 && cs.state.LastBlockHeight+1 != int64(cs.Height) {
			// This might happen when someone else is mutating cs.state.
			// Someone forgot to pass in state.Copy() somewhere?!
			panic(fmt.Sprintf("Inconsistent cs.state.LastBlockHeight+1 %v vs cs.Height %v",
				cs.state.LastBlockHeight+1, cs.Height))
		}
		if cs.state.LastBlockHeight > 0 && cs.Height == cs.state.InitialHeight {
			panic(fmt.Sprintf("Inconsistent cs.state.LastBlockHeight %v, expected 0 for initial height %v",
				cs.state.LastBlockHeight, cs.state.InitialHeight))
		}

		// If state isn't further out than cs.state, just ignore.
		// This happens when SwitchToConsensus() is called in the reactor.
		// We don't want to reset e.g. the Votes, but we still want to
		// signal the new round step, because other services (eg. txNotifier)
		// depend on having an up-to-date peer state!
		if state.LastBlockHeight > 0 && state.LastBlockHeight <= cs.state.LastBlockHeight {
			cs.Logger.Info(
				"Ignoring updateToState()",
				"newHeight",
				state.LastBlockHeight+1,
				"oldHeight",
				cs.state.LastBlockHeight+1)
			cs.newStep()
			return
		}
	}

	// Reset fields based on state.
	validators := state.Validators

	switch {
	case state.LastBlockHeight == -1: // Very first commit should be empty.
		cs.LastCommit = (*models.VoteSet)(nil)
	case cs.CommitRound > -1 && cs.Votes != nil: // Otherwise, use cs.Votes
		if !cs.Votes.Precommits(cs.CommitRound).HasTwoThirdsMajority() {
			panic(fmt.Sprintf("Wanted to form a Commit, but Precommits (H/R: %d/%d) didn't have 2/3+: %v",
				state.LastBlockHeight,
				cs.CommitRound,
				cs.Votes.Precommits(cs.CommitRound)))
		}
		cs.LastCommit = cs.Votes.Precommits(cs.CommitRound)
	case cs.LastCommit == nil:
		// NOTE: when Tendermint starts, it has no votes. reconstructLastCommit
		// must be called to reconstruct LastCommit from SeenCommit.
		panic(fmt.Sprintf("LastCommit cannot be empty after initial block (H:%d)",
			state.LastBlockHeight+1,
		))
	}

	// Next desired block height
	height := state.LastBlockHeight + 1
	if height == 0 {
		height = int64(state.InitialHeight)
	}

	// RoundState fields
	cs.updateHeight(uint64(height))
	cs.updateRoundStep(0, types.RoundStepNewHeight)
	if cs.CommitTime.IsZero() {
		// "Now" makes it easier to sync up dev nodes.
		// We add timeoutCommit to allow transactions
		// to be gathered for the first block.
		// And alternative solution that relies on clocks:
		// cs.StartTime = state.LastBlockTime.Add(timeoutCommit)
		cs.StartTime = cs.config.Commit(libs.Now())
	} else {
		cs.StartTime = cs.config.Commit(cs.CommitTime)
	}

	cs.Validators = validators
	cs.Proposal = nil
	cs.ProposalBlock = nil
	// cs.ProposalBlockParts = nil
	cs.LockedRound = -1
	cs.LockedBlock = nil
	cs.LockedBlockParts = nil
	cs.ValidRound = -1
	cs.ValidBlock = nil
	cs.ValidBlockParts = nil
	cs.Votes = models.NewHeightVoteSet(state.ChainID, uint64(height), validators)
	cs.CommitRound = -1
	cs.LastValidators = state.LastValidators
	cs.TriggeredTimeoutPrecommit = false

	cs.state = state

	// Finally, broadcast RoundState
	cs.newStep()
}
func (cs *State) updateHeight(height uint64) {
	// cs.metrics.Height.Set(float64(height))
	cs.Height = height
}
func (cs *State) enterPrecommit(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	if cs.Height != height || round < cs.Round || (cs.Round == round && types.RoundStepPrecommit <= cs.Step) {
		logger.Debug(fmt.Sprintf(
			"enterPrecommit(%v/%v): Invalid args. Current step: %v/%v/%v",
			height,
			round,
			cs.Height,
			cs.Round,
			cs.Step))
		return
	}

	logger.Debug(fmt.Sprintf("enterPrecommit(%v/%v); current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPrecommit:
		cs.updateRoundStep(round, types.RoundStepPrecommit)
		cs.newStep()
	}()

	// check for a polka
	blockID, ok := cs.Votes.Prevotes(round).TwoThirdsMajority()

	// If we don't have a polka, we must precommit nil.
	if !ok {
		if cs.LockedBlock != nil {
			logger.Debug("enterPrecommit: no +2/3 prevotes during enterPrecommit while we're locked. Precommitting nil")
		} else {
			logger.Debug("enterPrecommit: no +2/3 prevotes during enterPrecommit. Precommitting nil.")
		}
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, nil, models.PartSetHeader{})
		return
	}

	// At this point +2/3 prevoted for a particular block or nil.
	if err := cs.eventBus.PublishEventPolka(events.RoundStateEvent(cs.RoundState)); err != nil {
		cs.Logger.Error("Error publishing polka", "err", err)
	}

	// the latest POLRound should be this round.
	polRound, _ := cs.Votes.POLInfo()
	if polRound < round {
		panic(fmt.Sprintf("This POLRound should be %v but got %v", round, polRound))
	}

	// +2/3 prevoted nil. Unlock and precommit nil.
	if len(blockID.Hash) == 0 {
		if cs.LockedBlock == nil {
			logger.Debug("enterPrecommit: +2/3 prevoted for nil")
		} else {
			logger.Debug("enterPrecommit: +2/3 prevoted for nil; unlocking")
			cs.LockedRound = -1
			cs.LockedBlock = nil
			cs.LockedBlockParts = nil
			if err := cs.eventBus.PublishEventUnlock(events.RoundStateEvent(cs.RoundState)); err != nil {
				cs.Logger.Error("Error publishing event unlock", "err", err)
			}
		}
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, nil, models.PartSetHeader{})
		return
	}

	// At this point, +2/3 prevoted for a particular block.

	// If we're already locked on that block, precommit it, and update the LockedRound
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		logger.Debug("enterPrecommit: +2/3 prevoted locked block; relocking")
		cs.LockedRound = round
		if err := cs.eventBus.PublishEventRelock(events.RoundStateEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing event relock", "err", err)
		}
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, blockID.Hash, blockID.PartSetHeader)
		return
	}

	// If +2/3 prevoted for proposal block, stage and precommit it
	if cs.ProposalBlock.HashesTo(blockID.Hash) {
		logger.Debug("enterPrecommit: +2/3 prevoted proposal block; locking", "hash", blockID.Hash)
		// Validate the block.
		if err := cs.blockExecutor.ValidateBlock(cs.state, cs.ProposalBlock); err != nil {
			panic(fmt.Sprintf("enterPrecommit: +2/3 prevoted for an invalid block: %v", err))
		}
		cs.LockedRound = round
		cs.LockedBlock = cs.ProposalBlock
		// cs.LockedBlockParts = cs.ProposalBlockParts
		if err := cs.eventBus.PublishEventLock(events.RoundStateEvent(cs.RoundState)); err != nil {
			cs.Logger.Error("Error publishing event lock", "err", err)
		}
		cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, blockID.Hash, blockID.PartSetHeader)
		return
	}

	// There was a polka in this round for a block we don't have.
	// Fetch that block, unlock, and precommit nil.
	// The +2/3 prevotes for this round is the POL for our unlock.
	logger.Debug("enterPrecommit: +2/3 prevotes for a block we do not have; voting nil", "blockID", blockID)
	cs.LockedRound = -1
	cs.LockedBlock = nil
	cs.LockedBlockParts = nil
	// if !cs.ProposalBlockParts.HasHeader(blockID.PartSetHeader) {
	// 	cs.ProposalBlock = nil
	// 	cs.ProposalBlockParts = types.NewPartSetFromHeader(blockID.PartSetHeader)
	// }
	if err := cs.eventBus.PublishEventUnlock(events.RoundStateEvent(cs.RoundState)); err != nil {
		cs.Logger.Error("Error publishing event unlock", "err", err)
	}
	cs.signAddVote(types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, nil, models.PartSetHeader{})
}
func (cs *State) signAddVote(msgType types3.SignedMsgType, hash []byte, header models.PartSetHeader) *models.Vote {
	if cs.signer == nil { // the node does not have a key
		return nil
	}

	if cs.publicKey == nil {
		// Vote won't be signed, but it's not critical.
		cs.Logger.Error(fmt.Sprintf("signAddVote: %v", cs.publicKey))
		return nil
	}

	// If the node not in the validator set, do nothing.
	if !cs.Validators.HasAddress(cs.publicKey.Address()) {
		return nil
	}

	// TODO: pass pubKey to signVote
	vote, err := cs.signVote(msgType, hash, header)
	if err == nil {
		cs.sendInternalMessage(models.MsgInfo{&models.VoteMessage{Vote: vote}, ""})
		cs.Logger.Info("Signed and pushed vote", "height", cs.Height, "round", cs.Round, "vote", vote)
		return vote
	}
	// if !cs.replayMode {
	cs.Logger.Error("Error signing vote", "height", cs.Height, "round", cs.Round, "vote", vote, "err", err)
	// }
	return nil
}
func (cs *State) sendInternalMessage(mi models.MsgInfo) {
	select {
	case cs.internalMsgQueue <- mi:
	default:
		// NOTE: using the go-routine means our votes can
		// be processed out of order.
		// TODO: use CList here for strict determinism and
		// attempt push to internalMsgQueue in receiveRoutine
		cs.Logger.Info("Internal msg queue is full. Using a go-routine")
		go func() { cs.internalMsgQueue <- mi }()
	}
}

func (cs *State) signVote(
	msgType types3.SignedMsgType,
	hash []byte,
	header models.PartSetHeader,
) (*models.Vote, error) {
	// Flush the WAL. Otherwise, we may not recompute the same vote to sign,
	// and the privValidator will refuse to sign anything.
	if err := cs.wal.FlushAndSync(); err != nil {
		return nil, err
	}

	if cs.signer == nil {
		return nil, errPubKeyIsNotSet
	}
	addr := cs.publicKey.Address()
	valIdx, _ := cs.Validators.GetByAddress(addr)

	vote := &models.Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   valIdx,
		Height:           cs.Height,
		Round:            cs.Round,
		Timestamp:        cs.voteTime(),
		Type:             msgType,
		BlockID:          models.BlockID{Hash: hash, PartSetHeader: header},
	}
	// v := vote.ToProto()
	sig, err := cs.signer.SignVote(cs.state.ChainID, vote)
	if nil != err {
		return vote, err
	}
	vote.Signature = sig

	return vote, err
}
func (cs *State) voteTime() time.Time {
	now := libs.Now()
	minVoteTime := now
	// Minimum time increment between blocks
	const timeIota = time.Millisecond
	// TODO: We should remove next line in case we don't vote for v in case cs.ProposalBlock == nil,
	// even if cs.LockedBlock != nil. See https://docs.tendermint.com/master/spec/.
	if cs.LockedBlock != nil {
		// See the BFT time spec https://docs.tendermint.com/master/spec/consensus/bft-time.html
		minVoteTime = cs.LockedBlock.Time.Add(timeIota)
	} else if cs.ProposalBlock != nil {
		minVoteTime = cs.ProposalBlock.Time.Add(timeIota)
	}

	if now.After(minVoteTime) {
		return now
	}
	return minVoteTime
}
func (cs *State) addVote(vote *models.Vote, peerID types2.NodeID) (added bool, err error) {
	cs.Logger.Debug("addVote", "voteHeight", vote.Height, "voteType", vote.Type, "valIndex", vote.ValidatorIndex, "csHeight", cs.Height)

	// A precommit for the previous height?
	// These come in while we wait timeoutCommit
	// 可能是网络阻塞的消息
	if vote.Height+1 == cs.Height && vote.Type == types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT {
		// 如果是preCommit消息,此时只有当处于newHeight的时候才会处理
		if cs.Step != types.RoundStepNewHeight {
			// Late precommit at prior height is ignored
			cs.Logger.Debug("Precommit vote came in after commit timeout and has been ignored", "vote", vote)
			return
		}
		// 否则的话,也照旧更新投票信息
		added, err = cs.LastCommit.AddVote(vote)
		if !added {
			return
		}

		cs.Logger.Info(fmt.Sprintf("Added to lastPrecommits: %v", cs.LastCommit.StringShort()))
		// 发布event,其他组价订阅消费
		if err := cs.eventBus.PublishEventVote(events.EventDataVote{Vote: vote}); err != nil {
			return added, err
		}
		cs.eventManager.Fire(models2.DefaultEventWrapper{
			NameSpace: types.EventVote,
			EventData: vote,
		})

		// if we can skip timeoutCommit and have all the votes now,
		if cs.config.SkipTimeoutCommit && cs.LastCommit.HasAll() {
			// go straight to new round (skip timeout commit)
			// cs.scheduleTimeout(time.Duration(0), cs.Height, 0, cstypes.RoundStepNewHeight)
			// 开启新的一轮,同时高度是新的高度
			cs.enterNewRound(cs.Height, 0)
		}

		return
	}

	// Height mismatch is ignored.
	// Not necessarily a bad peer, but not favorable behavior.
	if vote.Height != cs.Height {
		cs.Logger.Debug("vote ignored and not added", "voteHeight", vote.Height, "csHeight", cs.Height, "peerID", peerID)
		return
	}

	height := cs.Height
	added, err = cs.Votes.AddVote(vote, peerID.ToPretty())
	if !added {
		// Either duplicate, or error upon cs.Votes.AddByIndex()
		return
	}

	if err := cs.eventBus.PublishEventVote(events.EventDataVote{Vote: vote}); err != nil {
		return added, err
	}
	cs.eventManager.Fire(models2.DefaultEventWrapper{
		NameSpace: types.EventVote,
		EventData: vote,
	})

	switch vote.Type {
	// 更新投票信息
	case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
		prevotes := cs.Votes.Prevotes(vote.Round)
		cs.Logger.Info("Added to prevote", "vote", vote, "prevotes", prevotes.StringShort())

		// If +2/3 prevotes for a block or nil for *any* round:
		if blockID, ok := prevotes.TwoThirdsMajority(); ok {
			// 收到了2/3
			// There was a polka!
			// If we're locked but this is a recent polka, unlock.
			// If it matches our ProposalBlock, update the ValidBlock

			// Unlock if `cs.LockedRound < vote.Round <= cs.Round`
			// NOTE: If vote.Round > cs.Round, we'll deal with it when we get to vote.Round
			if (cs.LockedBlock != nil) && (cs.LockedRound < vote.Round) && (vote.Round <= cs.Round) && !cs.LockedBlock.HashesTo(blockID.Hash) {
				cs.Logger.Info("Unlocking because of POL.", "lockedRound", cs.LockedRound, "POLRound", vote.Round)
				cs.LockedRound = -1
				cs.LockedBlock = nil
				cs.LockedBlockParts = nil
				if err := cs.eventBus.PublishEventUnlock(events.RoundStateEvent(cs.RoundState)); err != nil {
					return added, err
				}
			}

			// Update Valid* if we can.
			// NOTE: our proposal block may be nil or not what received a polka..
			if len(blockID.Hash) != 0 && (cs.ValidRound < vote.Round) && (vote.Round == cs.Round) {
				if cs.ProposalBlock.HashesTo(blockID.Hash) {
					cs.Logger.Info(
						"Updating ValidBlock because of POL.", "validRound", cs.ValidRound, "POLRound", vote.Round)
					cs.ValidRound = vote.Round
					cs.ValidBlock = cs.ProposalBlock
					cs.ValidBlockParts = cs.ProposalBlockParts
				} else {
					cs.Logger.Info(
						"valid block we do not know about; set ProposalBlock=nil",
						"proposal", cs.ProposalBlock.Hash(),
						"blockID", blockID.Hash,
					)

					// We're getting the wrong block.
					cs.ProposalBlock = nil
				}
				if !cs.ProposalBlockParts.HasHeader(blockID.PartSetHeader) {
					cs.ProposalBlockParts = models.NewPartSetFromHeader(blockID.PartSetHeader)
				}
				cs.eventManager.Fire(models2.DefaultEventWrapper{
					NameSpace: types.EventValidBlock,
					EventData: &cs.RoundState,
				})
				if err := cs.eventBus.PublishEventValidBlock(events.RoundStateEvent(cs.RoundState)); err != nil {
					return added, err
				}
			}
		}

		// If +2/3 prevotes for *anything* for future round:
		switch {
		case cs.Round < vote.Round && prevotes.HasTwoThirdsAny():
			// Round-skip if there is any 2/3+ of votes ahead of us
			cs.enterNewRound(height, vote.Round)
		case cs.Round == vote.Round && types.RoundStepPrevote <= cs.Step: // current round
			blockID, ok := prevotes.TwoThirdsMajority()
			if ok && (cs.isProposalComplete() || len(blockID.Hash) == 0) {
				cs.enterPrecommit(height, vote.Round)
			} else if prevotes.HasTwoThirdsAny() {
				cs.enterPrevoteWait(height, vote.Round)
			}
		case cs.Proposal != nil && 0 <= cs.Proposal.POLRound && cs.Proposal.POLRound == vote.Round:
			// If the proposal is now complete, enter prevote of cs.Round.
			if cs.isProposalComplete() {
				cs.enterPrevote(height, cs.Round)
			}
		}

	case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
		precommits := cs.Votes.Precommits(vote.Round)
		cs.Logger.Info("Added to precommit", "vote", vote, "precommits", precommits.StringShort())

		blockID, ok := precommits.TwoThirdsMajority()
		if ok {
			// Executed as TwoThirdsMajority could be from a higher round
			cs.enterNewRound(height, vote.Round)
			cs.enterPrecommit(height, vote.Round)
			if len(blockID.Hash) != 0 {
				cs.enterCommit(height, vote.Round)
				if cs.config.SkipTimeoutCommit && precommits.HasAll() {
					cs.enterNewRound(cs.Height, 0)
				}
			} else {
				cs.enterPrecommitWait(height, vote.Round)
			}
		} else if cs.Round <= vote.Round && precommits.HasTwoThirdsAny() {
			cs.enterNewRound(height, vote.Round)
			cs.enterPrecommitWait(height, vote.Round)
		}

	default:
		panic(fmt.Sprintf("Unexpected vote type %v", vote.Type))
	}

	return added, err
}
func (cs *State) enterPrecommitWait(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	if cs.Height != height || round < cs.Round || (cs.Round == round && cs.TriggeredTimeoutPrecommit) {
		logger.Debug(
			fmt.Sprintf(
				"enterPrecommitWait(%v/%v): Invalid args. "+
					"Current state is Height/Round: %v/%v/, TriggeredTimeoutPrecommit:%v",
				height, round, cs.Height, cs.Round, cs.TriggeredTimeoutPrecommit))
		return
	}
	if !cs.Votes.Precommits(round).HasTwoThirdsAny() {
		panic(fmt.Sprintf("enterPrecommitWait(%v/%v), but Precommits does not have any +2/3 votes", height, round))
	}
	logger.Debug(fmt.Sprintf("enterPrecommitWait(%v/%v). Current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPrecommitWait:
		cs.TriggeredTimeoutPrecommit = true
		cs.newStep()
	}()

	// Wait for some more precommits; enterNewRound
	cs.scheduleTimeout(cs.config.Precommit(round), height, round, types.RoundStepPrecommitWait)
}
func (cs *State) enterCommit(height uint64, commitRound int32) {
	// logger := cs.Logger.With("height", height, "commitRound", commitRound)
	logger := cs.Logger

	if cs.Height != height || types.RoundStepCommit <= cs.Step {
		logger.Debug(fmt.Sprintf(
			"enterCommit(%v/%v): Invalid args. Current step: %v/%v/%v",
			height,
			commitRound,
			cs.Height,
			cs.Round,
			cs.Step))
		return
	}
	logger.Info(fmt.Sprintf("enterCommit(%v/%v). Current: %v/%v/%v", height, commitRound, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterCommit:
		// keep cs.Round the same, commitRound points to the right Precommits set.
		cs.updateRoundStep(cs.Round, types.RoundStepCommit)
		cs.CommitRound = commitRound
		cs.CommitTime = libs.Now()
		cs.newStep()

		// Maybe finalize immediately.
		cs.tryFinalizeCommit(height)
	}()

	blockID, ok := cs.Votes.Precommits(commitRound).TwoThirdsMajority()
	if !ok {
		panic("RunActionCommit() expects +2/3 precommits")
	}

	// The Locked* fields no longer matter.
	// Move them over to ProposalBlock if they match the commit hash,
	// otherwise they'll be cleared in updateToState.
	if cs.LockedBlock.HashesTo(blockID.Hash) {
		logger.Debug("commit is for a locked block; set ProposalBlock=LockedBlock", "blockHash", blockID.Hash)
		cs.ProposalBlock = cs.LockedBlock
		// cs.ProposalBlockParts = cs.LockedBlockParts
	}

	// If we don't have the block being committed, set up to get it.
	if !cs.ProposalBlock.HashesTo(blockID.Hash) {
		if !cs.ProposalBlockParts.HasHeader(blockID.PartSetHeader) {
			logger.Info(
				"commit is for a block we do not know about; set ProposalBlock=nil",
				"proposal", cs.ProposalBlock.Hash(),
				"commit", blockID.Hash,
			)

			// We're getting the wrong block.
			// Set up ProposalBlockParts and keep waiting.
			cs.ProposalBlock = nil
			cs.ProposalBlockParts = models.NewPartSetFromHeader(blockID.PartSetHeader)
			if err := cs.eventBus.PublishEventValidBlock(events.RoundStateEvent(cs.RoundState)); err != nil {
				cs.Logger.Error("Error publishing valid block", "err", err)
			}
			cs.eventManager.Fire(models2.DefaultEventWrapper{
				NameSpace: types.EventValidBlock,
				EventData: &cs.RoundState,
			})
		}
	}
	// else {
	// We just need to keep waiting.
	// }
}
func (cs *State) enterPrevoteWait(height uint64, round int32) {
	// logger := cs.Logger.With("height", height, "round", round)
	logger := cs.Logger

	if cs.Height != height || round < cs.Round || (cs.Round == round && types.RoundStepPrevoteWait <= cs.Step) {
		logger.Debug(fmt.Sprintf(
			"enterPrevoteWait(%v/%v): Invalid args. Current step: %v/%v/%v",
			height,
			round,
			cs.Height,
			cs.Round,
			cs.Step))
		return
	}
	if !cs.Votes.Prevotes(round).HasTwoThirdsAny() {
		panic(fmt.Sprintf("enterPrevoteWait(%v/%v), but Prevotes does not have any +2/3 votes", height, round))
	}

	logger.Debug(fmt.Sprintf("enterPrevoteWait(%v/%v); current: %v/%v/%v", height, round, cs.Height, cs.Round, cs.Step))

	defer func() {
		// Done enterPrevoteWait:
		cs.updateRoundStep(round, types.RoundStepPrevoteWait)
		cs.newStep()
	}()

	// Wait for some more prevotes; enterPrecommit
	cs.scheduleTimeout(cs.config.Prevote(round), height, round, types.RoundStepPrevoteWait)
}
func (cs *State) scheduleTimeout(duration time.Duration, height uint64, round int32, step types.RoundStepType) {
	cs.tikToker.ScheduleTimeout(models.TimeoutInfo{Duration: duration, Height: height, Round: round, Step: step})
}
func (cs *State) loadWalFile() error {
	wal, err := cs.OpenWAL(cs.config.WalFile())
	if err != nil {
		cs.Logger.Error("Error loading State wal", "err", err)
		return err
	}
	cs.wal = wal
	return nil
}

// OpenWAL opens a file to log all consensus messages and timeouts for
// deterministic accountability.
func (cs *State) OpenWAL(walFile string) (services.IWAL, error) {
	wal, err := services.NewWAL(walFile)
	if err != nil {
		cs.Logger.Error("Failed to open WAL", "file", walFile, "err", err)
		return nil, err
	}
	wal.SetLogger(cs.Logger)
	// wal.SetLogger(cs.Logger.With("wal", walFile))
	if err := wal.Start(services4.ASYNC_START); err != nil {
		cs.Logger.Error("Failed to start WAL", "err", err)
		return nil, err
	}
	return wal, nil
}

func (s *State) StateNotify() <-chan models.MsgInfo {
	return s.statsMsgQueue
}

func (s *State) PeerMsgNotify() chan models.MsgInfo {
	return s.peerMsgQueue
}

func (s *State) LoadCommit(height uint64) *models.Commit {
	s.Lock()
	defer s.Unlock()

	if height == s.blockStore.Height() {
		return s.blockStore.LoadSeenCommit(height)
	}

	return s.blockStore.LoadBlockCommit(height)
}

// SetPrivValidator sets the private validator account for signing votes. It
// immediately requests pubkey and caches it.
func (cs *State) SetPrivValidator(priv cryptolibs.ISigner) {
	cs.Lock()
	defer cs.Unlock()

	cs.signer = priv

	if err := cs.updatePrivValidatorPubKey(); err != nil {
		cs.Logger.Error("failed to get private validator pubkey", "err", err)
	}
}

// SetTimeoutTicker sets the local timer. It may be useful to overwrite for
// testing.
func (cs *State) SetTimeoutTicker(timeoutTicker services.ITimeoutTicker) {
	cs.Lock()
	cs.tikToker = timeoutTicker
	cs.Unlock()
}

func (s *State) GetInitialHeight(lock bool) uint64 {
	s.RLock()
	defer s.RUnlock()
	return s.state.InitialHeight
}

func (s *State) GetCurrentHeightWithVote(lock bool) (uint64, *models.HeightVoteSet) {
	s.RLock()
	defer s.RUnlock()
	height, votes := s.Height, s.Votes
	return height, votes
}

func (s *State) GetSizeWithHeightWithLastCommitSize(lock bool) (uint64, int, int) {
	s.RLock()
	defer s.RUnlock()
	height, valSize, lastCommitSize := s.Height, s.Validators.Size(), s.LastCommit.Size()
	return height, valSize, lastCommitSize
}

func (s *State) GetRoundState() *models.RoundState {
	return &s.RoundState
}

func (s *State) GetBlockStore() services.IBlockStore {
	return s.blockStore
}

// timeoutRoutine: receive requests for timeouts on tickChan and fire timeouts on tockChan
// receiveRoutine: serializes processing of proposoals, block parts, votes; coordinates state transitions
func (cs *State) startRoutines(maxSteps int) {
	err := cs.tikToker.Start(services4.ASYNC_START)
	if err != nil {
		cs.Logger.Error("Error starting timeout ticker", "err", err)
		return
	}
	go cs.receiveRoutine(maxSteps)
}
