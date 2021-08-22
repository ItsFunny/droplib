/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:10 上午
# @File : evidence.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sort"
	"strings"
	"time"
)
type EvidenceList []transfer.Evidence
func (evl EvidenceList) Hash() []byte {
	// These allocations are required because Evidence is not of type Bytes, and
	// golang slices can't be typed cast. This shouldn't be a performance problem since
	// the Evidence size is capped.
	evidenceBzs := make([][]byte, len(evl))
	for i := 0; i < len(evl); i++ {
		evidenceBzs[i] = evl[i].Bytes()
	}
	return merkle.HashFromByteSlices(evidenceBzs)
}
// EvidenceData contains any evidence of malicious wrong-doing by validators
type EvidenceData struct {
	Evidence EvidenceList `json:"evidence"`

	// Volatile. Used as cache
	hash     libs.HexBytes
	byteSize int64
}
func (data *EvidenceData) ByteSize() int64 {
	if data.byteSize == 0 && len(data.Evidence) != 0 {
		pb, err := data.ToEvidenceListProto()
		if err != nil {
			panic(err)
		}
		data.byteSize = int64(pb.XXX_Size())
	}
	return data.byteSize
}
func (data *EvidenceData) ToEvidenceListProto() (*types.EvidenceListProto, error) {
	if data == nil {
		return nil, errors.New("nil evidence data")
	}

	evi := new(types.EvidenceListProto)
	eviBzs := make([]*types.EvidenceProto, len(data.Evidence))
	for i := range data.Evidence {
		protoEvi, err := EvidenceToProto(data.Evidence[i])
		if err != nil {
			return nil, err
		}
		eviBzs[i] = protoEvi
	}
	evi.Evidence = eviBzs

	return evi, nil
}
func EvidenceToProto(evidence transfer.Evidence) (*types.EvidenceProto, error) {
	if evidence == nil {
		return nil, errors.New("nil evidence")
	}

	switch evi := evidence.(type) {
	case *DuplicateVoteEvidence:
		pbev := evi.ToProto()
		return &types.EvidenceProto{
			Sum: &types.EvidenceProto_DuplicateVoteEvidence{
				DuplicateVoteEvidence: pbev,
			},
		}, nil

	case *LightClientAttackEvidence:
		pbev, err := evi.ToProto()
		if err != nil {
			return nil, err
		}
		return &types.EvidenceProto{
			Sum: &types.EvidenceProto_LightClientAttackEvidence{
				LightClientAttackEvidence: pbev,
			},
		}, nil

	default:
		return nil, fmt.Errorf("toproto: evidence is not recognized: %T", evi)
	}
}
func (data *EvidenceData) FromProto(eviData *types.EvidenceListProto) error {
	if eviData == nil {
		return errors.New("nil evidenceData")
	}

	eviBzs := make(EvidenceList, len(eviData.Evidence))
	for i := range eviData.Evidence {
		evi, err := EvidenceFromProto(eviData.Evidence[i])
		if err != nil {
			return err
		}
		eviBzs[i] = evi
	}
	data.Evidence = eviBzs
	data.byteSize = int64(commonutils.ProtoSize(eviData))

	return nil
}
func (data *EvidenceData) Hash() libs.HexBytes {
	if data.hash == nil {
		data.hash = data.Evidence.Hash()
	}
	return data.hash
}



type DuplicateVoteEvidence struct {
	VoteA *Vote `json:"vote_a"`
	VoteB *Vote `json:"vote_b"`

	// abci specific information
	TotalVotingPower int64
	ValidatorPower   int64
	Timestamp        time.Time
}

var _ transfer.Evidence = &DuplicateVoteEvidence{}

// NewDuplicateVoteEvidence creates DuplicateVoteEvidence with right ordering given
// two conflicting votes. If one of the votes is nil, evidence returned is nil as well
func NewDuplicateVoteEvidence(vote1, vote2 *Vote, blockTime time.Time, valSet *ValidatorSet) *DuplicateVoteEvidence {
	var voteA, voteB *Vote
	if vote1 == nil || vote2 == nil || valSet == nil {
		return nil
	}
	idx, val := valSet.GetByAddress(vote1.ValidatorAddress)
	if idx == -1 {
		return nil
	}

	if strings.Compare(vote1.BlockID.Key(), vote2.BlockID.Key()) == -1 {
		voteA = vote1
		voteB = vote2
	} else {
		voteA = vote2
		voteB = vote1
	}
	return &DuplicateVoteEvidence{
		VoteA:            voteA,
		VoteB:            voteB,
		TotalVotingPower: valSet.TotalVotingPower(),
		ValidatorPower:   val.VotingPower,
		Timestamp:        blockTime,
	}
}

// Bytes returns the proto-encoded evidence as a byte array.
func (dve *DuplicateVoteEvidence) Bytes() []byte {
	pbe := dve.ToProto()
	bz, err := commonutils.Marshal(pbe)
	// bz, err := pbe.Marshal()
	if err != nil {
		panic(err)
	}

	return bz
}

// Hash returns the hash of the evidence.
func (dve *DuplicateVoteEvidence) Hash() []byte {
	return libs.Sum(dve.Bytes())
}

// Height returns the height of the infraction
func (dve *DuplicateVoteEvidence) Height() int64 {
	return int64(dve.VoteA.Height)
}

// String returns a string representation of the evidence.
func (dve *DuplicateVoteEvidence) String() string {
	return fmt.Sprintf("DuplicateVoteEvidence{VoteA: %v, VoteB: %v}", dve.VoteA, dve.VoteB)
}

// Time returns the time of the infraction
func (dve *DuplicateVoteEvidence) Time() time.Time {
	return dve.Timestamp
}

// ValidateBasic performs basic validation.
func (dve *DuplicateVoteEvidence) ValidateBasic() error {
	if dve == nil {
		return errors.New("empty duplicate vote evidence")
	}

	if dve.VoteA == nil || dve.VoteB == nil {
		return fmt.Errorf("one or both of the votes are empty %v, %v", dve.VoteA, dve.VoteB)
	}
	if err := dve.VoteA.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid VoteA: %w", err)
	}
	if err := dve.VoteB.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid VoteB: %w", err)
	}
	// Enforce Votes are lexicographically sorted on blockID
	if strings.Compare(dve.VoteA.BlockID.Key(), dve.VoteB.BlockID.Key()) >= 0 {
		return errors.New("duplicate votes in invalid order")
	}
	return nil
}

// ToProto encodes DuplicateVoteEvidence to protobuf
func (dve *DuplicateVoteEvidence) ToProto() *types.DuplicateVoteEvidenceProto {
	voteB := dve.VoteB.ToProto()
	voteA := dve.VoteA.ToProto()
	tp := types.DuplicateVoteEvidenceProto{
		VoteA:            voteA,
		VoteB:            voteB,
		TotalVotingPower: dve.TotalVotingPower,
		ValidatorPower:   dve.ValidatorPower,
		Timestamp:        commonutils.ConvtTime2GoogleTime(dve.Timestamp),
	}
	return &tp
}


// LightClientAttackEvidence is a generalized evidence that captures all forms of known attacks on
// a light client such that a full node can verify, propose and commit the evidence on-chain for
// punishment of the malicious validators. There are three forms of attacks: Lunatic, Equivocation
// and Amnesia. These attacks are exhaustive. You can find a more detailed overview of this at
// tendermint/docs/architecture/adr-047-handling-evidence-from-light-client.md
type LightClientAttackEvidence struct {
	ConflictingBlock *LightBlock
	CommonHeight     int64

	// abci specific information
	ByzantineValidators []*Validator // validators in the validator set that misbehaved in creating the conflicting block
	TotalVotingPower    int64        // total voting power of the validator set at the common height
	Timestamp           time.Time    // timestamp of the block at the common height
}

var _ transfer.Evidence = &LightClientAttackEvidence{}

// ABCI forms an array of abci evidence for each byzantine validator
// func (l *LightClientAttackEvidence) ABCI() []abci.Evidence {
// 	abciEv := make([]abci.Evidence, len(l.ByzantineValidators))
// 	for idx, val := range l.ByzantineValidators {
// 		abciEv[idx] = abci.Evidence{
// 			Type:             abci.EvidenceType_LIGHT_CLIENT_ATTACK,
// 			Validator:        TM2PB.Validator(val),
// 			Height:           l.Height(),
// 			Time:             l.Timestamp,
// 			TotalVotingPower: l.TotalVotingPower,
// 		}
// 	}
// 	return abciEv
// }

// Bytes returns the proto-encoded evidence as a byte array
func (l *LightClientAttackEvidence) Bytes() []byte {
	pbe, err := l.ToProto()
	if err != nil {
		panic(err)
	}
	// bz, err := pbe.Marshal()
	bz,err:=commonutils.Marshal(pbe)
	if err != nil {
		panic(err)
	}
	return bz
}

// GetByzantineValidators finds out what style of attack LightClientAttackEvidence was and then works out who
// the malicious validators were and returns them. This is used both for forming the ByzantineValidators
// field and for validating that it is correct. Validators are ordered based on validator power
func (l *LightClientAttackEvidence) GetByzantineValidators(commonVals *ValidatorSet,
	trusted *SignedHeader) []*Validator {
	var validators []*Validator
	// First check if the header is invalid. This means that it is a lunatic attack and therefore we take the
	// validators who are in the commonVals and voted for the lunatic header
	if l.ConflictingHeaderIsInvalid(trusted.TendermintBlockHeader) {
		for _, commitSig := range l.ConflictingBlock.Commit.Signatures {
			if !commitSig.ForBlock() {
				continue
			}

			_, val := commonVals.GetByAddress(commitSig.ValidatorAddress)
			if val == nil {
				// validator wasn't in the common validator set
				continue
			}
			validators = append(validators, val)
		}
		sort.Sort(ValidatorsByVotingPower(validators))
		return validators
	} else if trusted.Commit.Round == l.ConflictingBlock.Commit.Round {
		// This is an equivocation attack as both commits are in the same round. We then find the validators
		// from the conflicting light block validator set that voted in both headers.
		// Validator hashes are the same therefore the indexing order of validators are the same and thus we
		// only need a single loop to find the validators that voted twice.
		for i := 0; i < len(l.ConflictingBlock.Commit.Signatures); i++ {
			sigA := l.ConflictingBlock.Commit.Signatures[i]
			if sigA.Absent() {
				continue
			}

			sigB := trusted.Commit.Signatures[i]
			if sigB.Absent() {
				continue
			}

			_, val := l.ConflictingBlock.ValidatorSet.GetByAddress(sigA.ValidatorAddress)
			validators = append(validators, val)
		}
		sort.Sort(ValidatorsByVotingPower(validators))
		return validators
	}
	// if the rounds are different then this is an amnesia attack. Unfortunately, given the nature of the attack,
	// we aren't able yet to deduce which are malicious validators and which are not hence we return an
	// empty validator set.
	return validators
}

// ConflictingHeaderIsInvalid takes a trusted header and matches it againt a conflicting header
// to determine whether the conflicting header was the product of a valid state transition
// or not. If it is then all the deterministic fields of the header should be the same.
// If not, it is an invalid header and constitutes a lunatic attack.
func (l *LightClientAttackEvidence) ConflictingHeaderIsInvalid(trustedHeader *TendermintBlockHeader) bool {
	return !bytes.Equal(trustedHeader.ValidatorsHash, l.ConflictingBlock.ValidatorsHash) ||
		!bytes.Equal(trustedHeader.NextValidatorsHash, l.ConflictingBlock.NextValidatorsHash) ||
		!bytes.Equal(trustedHeader.ConsensusHash, l.ConflictingBlock.ConsensusHash) ||
		// !bytes.Equal(trustedHeader.AppHash, l.ConflictingBlock.AppHash) ||
		!bytes.Equal(trustedHeader.LastResultsHash, l.ConflictingBlock.LastResultsHash)

}

// Hash returns the hash of the header and the commonHeight. This is designed to cause hash collisions
// with evidence that have the same conflicting header and common height but different permutations
// of validator commit signatures. The reason for this is that we don't want to allow several
// permutations of the same evidence to be committed on chain. Ideally we commit the header with the
// most commit signatures (captures the most byzantine validators) but anything greater than 1/3 is sufficient.
func (l *LightClientAttackEvidence) Hash() []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, l.CommonHeight)
	bz := make([]byte, libs.Size+n)
	copy(bz[:libs.Size-1], l.ConflictingBlock.Hash().Bytes())
	copy(bz[libs.Size:], buf)
	return libs.Sum(bz)
}

// Height returns the last height at which the primary provider and witness provider had the same header.
// We use this as the height of the infraction rather than the actual conflicting header because we know
// that the malicious validators were bonded at this height which is important for evidence expiry
func (l *LightClientAttackEvidence) Height() int64 {
	return l.CommonHeight
}

// String returns a string representation of LightClientAttackEvidence
func (l *LightClientAttackEvidence) String() string {
	return fmt.Sprintf("LightClientAttackEvidence{ConflictingBlock: %v, CommonHeight: %d}",
		l.ConflictingBlock.String(), l.CommonHeight)
}

// Time returns the time of the common block where the infraction leveraged off.
func (l *LightClientAttackEvidence) Time() time.Time {
	return l.Timestamp
}

// ValidateBasic performs basic validation such that the evidence is consistent and can now be used for verification.
func (l *LightClientAttackEvidence) ValidateBasic() error {
	if l.ConflictingBlock == nil {
		return errors.New("conflicting block is nil")
	}

	// this check needs to be done before we can run validate basic
	if l.ConflictingBlock.TendermintBlockHeader == nil {
		return errors.New("conflicting block missing header")
	}

	if err := l.ConflictingBlock.ValidateBasic(l.ConflictingBlock.ChainID); err != nil {
		return fmt.Errorf("invalid conflicting light block: %w", err)
	}

	if l.CommonHeight <= 0 {
		return errors.New("negative or zero common height")
	}

	// check that common height isn't ahead of the height of the conflicting block. It
	// is possible that they are the same height if the light node witnesses either an
	// amnesia or a equivocation attack.
	if l.CommonHeight > l.ConflictingBlock.Height {
		return fmt.Errorf("common height is ahead of the conflicting block height (%d > %d)",
			l.CommonHeight, l.ConflictingBlock.Height)
	}

	return nil
}

// ToProto encodes LightClientAttackEvidence to protobuf
func (l *LightClientAttackEvidence) ToProto() (*types.LightClientAttackEvidenceProto, error) {
	conflictingBlock, err := l.ConflictingBlock.ToProto()
	if err != nil {
		return nil, err
	}

	byzVals := make([]*types.ValidatorProto, len(l.ByzantineValidators))
	for idx, val := range l.ByzantineValidators {
		valpb, err := val.ToProto()
		if err != nil {
			return nil, err
		}
		byzVals[idx] = valpb
	}

	return &types.LightClientAttackEvidenceProto{
		ConflictingBlock:    conflictingBlock,
		CommonHeight:        l.CommonHeight,
		ByzantineValidators: byzVals,
		TotalVotingPower:    l.TotalVotingPower,
		Timestamp:           commonutils.ConvtTime2GoogleTime(l.Timestamp),
	}, nil
}

// ErrEvidenceOverflow is for when there the amount of evidence exceeds the max bytes.
type ErrEvidenceOverflow struct {
	Max int64
	Got int64
}

// NewErrEvidenceOverflow returns a new ErrEvidenceOverflow where got > max.
func NewErrEvidenceOverflow(max, got int64) *ErrEvidenceOverflow {
	return &ErrEvidenceOverflow{max, got}
}

// Error returns a string representation of the error.
func (err *ErrEvidenceOverflow) Error() string {
	return fmt.Sprintf("Too much evidence: Max %d, got %d", err.Max, err.Got)
}