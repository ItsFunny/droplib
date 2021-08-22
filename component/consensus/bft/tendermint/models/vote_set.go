/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:13 上午 
# @File : vote_set.go
# @Description : 
# @Attention : 
*/
package models

import (
	"bytes"
	"fmt"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sync"
)


type blockVotes struct {
	peerMaj23 bool           // peer claims to have maj23
	bitArray  *libs.BitArray // valIndex -> hasVote?
	votes     []*Vote        // valIndex -> *Vote
	sum       int64          // vote sum
}
func (vs *blockVotes) addVerifiedVote(vote *Vote, votingPower int64) {
	valIndex := vote.ValidatorIndex
	if existing := vs.votes[valIndex]; existing == nil {
		vs.bitArray.SetIndex(int(valIndex), true)
		vs.votes[valIndex] = vote
		vs.sum += votingPower
	}
}
func (vs *blockVotes) getByIndex(index int32) *Vote {
	if vs == nil {
		return nil
	}
	return vs.votes[index]
}
type VoteSet struct {
	chainID       string
	height        uint64
	round         int32
	signedMsgType types.SignedMsgType
	valSet        *ValidatorSet

	mtx           sync.Mutex
	votesBitArray *libs.BitArray
	votes         []*Vote                // Primary votes to share
	sum           int64                  // Sum of voting power for seen votes, discounting conflicts
	maj23         *BlockID               // First 2/3 majority seen
	votesByBlock  map[string]*blockVotes // string(blockHash|blockParts) -> blockVotes
	peerMaj23s    map[types2.PeerIDPretty]BlockID      // Maj23 for each peer
}
func NewCommit(height uint64, round int32, blockID BlockID, commitSigs []CommitSig) *Commit {
	return &Commit{
		Height:     height,
		Round:      round,
		BlockID:    blockID,
		Signatures: commitSigs,
	}
}
func (voteSet *VoteSet) MakeCommit() *Commit {
	if voteSet.signedMsgType != types.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT {
		panic("Cannot MakeCommit() unless VoteSet.Type is PrecommitType")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	// Make sure we have a 2/3 majority
	if voteSet.maj23 == nil {
		panic("Cannot MakeCommit() unless a blockhash has +2/3")
	}

	// For every validator, get the precommit
	commitSigs := make([]CommitSig, len(voteSet.votes))
	for i, v := range voteSet.votes {
		commitSig := v.CommitSig()
		// if block ID exists but doesn't match, exclude sig
		if commitSig.ForBlock() && !v.BlockID.Equals(*voteSet.maj23) {
			commitSig = NewCommitSigAbsent()
		}
		commitSigs[i] = commitSig
	}

	return NewCommit(voteSet.GetHeight(), voteSet.GetRound(), *voteSet.maj23, commitSigs)
}
func (voteSet *VoteSet) HasTwoThirdsMajority() bool {
	if voteSet == nil {
		return false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.maj23 != nil
}
func (voteSet *VoteSet) HasTwoThirdsAny() bool {
	if voteSet == nil {
		return false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.sum > voteSet.valSet.TotalVotingPower()*2/3
}
func (voteSet *VoteSet) HasAll() bool {
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.sum == voteSet.valSet.TotalVotingPower()
}
func (voteSet *VoteSet) StringShort() string {
	if voteSet == nil {
		return "nil-VoteSet"
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	_, _, frac := voteSet.sumTotalFrac()
	return fmt.Sprintf(`VoteSet{H:%v R:%v T:%v +2/3:%v(%v) %v %v}`,
		voteSet.height, voteSet.round, voteSet.signedMsgType, voteSet.maj23, frac, voteSet.votesBitArray, voteSet.peerMaj23s)
}
func (voteSet *VoteSet) sumTotalFrac() (int64, int64, float64) {
	voted, total := voteSet.sum, voteSet.valSet.TotalVotingPower()
	fracVoted := float64(voted) / float64(total)
	return voted, total, fracVoted
}
// Constructs a new VoteSet struct used to accumulate votes for given height/round.
func NewVoteSet(chainID string, height uint64, round int32,
	signedMsgType types.SignedMsgType, valSet *ValidatorSet) *VoteSet {
	// if height == 0 {
	// 	panic("Cannot make VoteSet for height == 0, doesn't make sense.")
	// }
	return &VoteSet{
		chainID:       chainID,
		height:        height,
		round:         round,
		signedMsgType: signedMsgType,
		valSet:        valSet,
		votesBitArray: libs.NewBitArray(valSet.Size()),
		votes:         make([]*Vote, valSet.Size()),
		sum:           0,
		maj23:         nil,
		votesByBlock:  make(map[string]*blockVotes, valSet.Size()),
		peerMaj23s:    make(map[types2.PeerIDPretty]BlockID),
	}
}

func (voteSet *VoteSet) TwoThirdsMajority() (blockID BlockID, ok bool) {
	if voteSet == nil {
		return BlockID{}, false
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	if voteSet.maj23 != nil {
		return *voteSet.maj23, true
	}
	return BlockID{}, false
}
func (voteSet *VoteSet) GetByIndex(i int32) *Vote {
	panic("implement me")
}

func (voteSet *VoteSet) IsCommit() bool {
	panic("implement me")
}

func (voteSet *VoteSet) BitArray() *libs.BitArray {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	return voteSet.votesBitArray.Copy()
}
func (voteSet *VoteSet) BitArrayByBlockID(blockID BlockID) *libs.BitArray {
	if voteSet == nil {
		return nil
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()
	votesByBlock, ok := voteSet.votesByBlock[blockID.Key()]
	if ok {
		return votesByBlock.bitArray.Copy()
	}
	return nil
}


// Implements VoteSetReader.
func (voteSet *VoteSet) GetHeight() uint64 {
	if voteSet == nil {
		return 0
	}
	return voteSet.height
}

// Implements VoteSetReader.
func (voteSet *VoteSet) GetRound() int32 {
	if voteSet == nil {
		return -1
	}
	return voteSet.round
}

// Implements VoteSetReader.
func (voteSet *VoteSet) Type() byte {
	if voteSet == nil {
		return 0x00
	}
	return byte(voteSet.signedMsgType)
}

// Implements VoteSetReader.
func (voteSet *VoteSet) Size() int {
	if voteSet == nil {
		return 0
	}
	return voteSet.valSet.Size()
}

// Returns added=true if vote is valid and new.
// Otherwise returns err=ErrVote[
//		UnexpectedStep | InvalidIndex | InvalidAddress |
//		InvalidSignature | InvalidBlockHash | ConflictingVotes ]
// Duplicate votes return added=false, err=nil.
// Conflicting votes return added=*, err=ErrVoteConflictingVotes.
// NOTE: vote should not be mutated after adding.
// NOTE: VoteSet must not be nil
// NOTE: Vote must not be nil
func (voteSet *VoteSet) AddVote(vote *Vote) (added bool, err error) {
	if voteSet == nil {
		panic("AddVote() on nil VoteSet")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	return voteSet.addVote(vote)
}
func (voteSet *VoteSet) addVerifiedVote(
	vote *Vote,
	blockKey string,
	votingPower int64,
) (added bool, conflicting *Vote) {
	valIndex := vote.ValidatorIndex

	// Already exists in voteSet.votes?
	if existing := voteSet.votes[valIndex]; existing != nil {
		if existing.BlockID.Equals(vote.BlockID) {
			panic("addVerifiedVote does not expect duplicate votes")
		} else {
			conflicting = existing
		}
		// Replace vote if blockKey matches voteSet.maj23.
		if voteSet.maj23 != nil && voteSet.maj23.Key() == blockKey {
			voteSet.votes[valIndex] = vote
			voteSet.votesBitArray.SetIndex(int(valIndex), true)
		}
		// Otherwise don't add it to voteSet.votes
	} else {
		// Add to voteSet.votes and incr .sum
		voteSet.votes[valIndex] = vote
		voteSet.votesBitArray.SetIndex(int(valIndex), true)
		voteSet.sum += votingPower
	}

	votesByBlock, ok := voteSet.votesByBlock[blockKey]
	if ok {
		if conflicting != nil && !votesByBlock.peerMaj23 {
			// There's a conflict and no peer claims that this block is special.
			return false, conflicting
		}
		// We'll add the vote in a bit.
	} else {
		// .votesByBlock doesn't exist...
		if conflicting != nil {
			// ... and there's a conflicting vote.
			// We're not even tracking this blockKey, so just forget it.
			return false, conflicting
		}
		// ... and there's no conflicting vote.
		// Start tracking this blockKey
		votesByBlock = newBlockVotes(false, voteSet.valSet.Size())
		voteSet.votesByBlock[blockKey] = votesByBlock
		// We'll add the vote in a bit.
	}

	// Before adding to votesByBlock, see if we'll exceed quorum
	origSum := votesByBlock.sum
	quorum := voteSet.valSet.TotalVotingPower()*2/3 + 1

	// Add vote to votesByBlock
	votesByBlock.addVerifiedVote(vote, votingPower)

	// If we just crossed the quorum threshold and have 2/3 majority...
	if origSum < quorum && quorum <= votesByBlock.sum {
		// Only consider the first quorum reached
		if voteSet.maj23 == nil {
			maj23BlockID := vote.BlockID
			voteSet.maj23 = &maj23BlockID
			// And also copy votes over to voteSet.votes
			for i, vote := range votesByBlock.votes {
				if vote != nil {
					voteSet.votes[i] = vote
				}
			}
		}
	}

	return true, conflicting
}
// NOTE: Validates as much as possible before attempting to verify the signature.
func (voteSet *VoteSet) addVote(vote *Vote) (added bool, err error) {
	if vote == nil {
		return false, commonerrors.ErrVoteNil
	}
	valIndex := vote.ValidatorIndex
	valAddr := vote.ValidatorAddress
	blockKey := vote.BlockID.Key()

	// Ensure that validator index was set
	if valIndex < 0 {
		return false, fmt.Errorf("index < 0: %w", commonerrors.ErrVoteInvalidValidatorIndex)
	} else if len(valAddr) == 0 {
		return false, fmt.Errorf("empty address: %w", commonerrors.ErrVoteInvalidValidatorAddress)
	}

	// Make sure the step matches.
	if (vote.Height != voteSet.height) ||
		(vote.Round != voteSet.round) ||
		(vote.Type != voteSet.signedMsgType) {
		return false, fmt.Errorf("expected %d/%d/%d, but got %d/%d/%d: %w",
			voteSet.height, voteSet.round, voteSet.signedMsgType,
			vote.Height, vote.Round, vote.Type, commonerrors.ErrVoteUnexpectedStep)
	}

	// Ensure that signer is a validator.
	lookupAddr, val := voteSet.valSet.GetByIndex(valIndex)
	if val == nil {
		return false, fmt.Errorf(
			"cannot find validator %d in valSet of size %d: %w",
			valIndex, voteSet.valSet.Size(), commonerrors.ErrVoteInvalidValidatorIndex)
	}

	// Ensure that the signer has the right address.
	if !bytes.Equal(valAddr, lookupAddr) {
		return false, fmt.Errorf(
			"vote.ValidatorAddress (%X) does not match address (%X) for vote.ValidatorIndex (%d)\n"+
				"Ensure the genesis file is correct across all validators: %w",
			valAddr, lookupAddr, valIndex,commonerrors.ErrVoteInvalidValidatorAddress)
	}

	// If we already know of this vote, return false.
	if existing, ok := voteSet.getVote(valIndex, blockKey); ok {
		if bytes.Equal(existing.Signature, vote.Signature) {
			return false, nil // duplicate
		}
		return false, fmt.Errorf("existing vote: %v; new vote: %v: %w", existing, vote, commonerrors.ErrVoteNonDeterministicSignature)
	}

	// Check signature.
	if err := vote.Verify(voteSet.chainID, val.PubKey); err != nil {
		return false, fmt.Errorf("failed to verify vote with ChainID %s and PubKey %s: %w", voteSet.chainID, val.PubKey, err)
	}

	// Add vote and get conflicting vote if any.
	added, conflicting := voteSet.addVerifiedVote(vote, blockKey, val.VotingPower)
	if conflicting != nil {
		return added, NewConflictingVoteError(conflicting, vote)
	}
	if !added {
		panic("Expected to add non-conflicting vote")
	}
	return added, nil
}

type ErrVoteConflictingVotes struct {
	VoteA *Vote
	VoteB *Vote
}

func (err *ErrVoteConflictingVotes) Error() string {
	return fmt.Sprintf("conflicting votes from validator %X", err.VoteA.ValidatorAddress)
}
func NewConflictingVoteError(vote1, vote2 *Vote) *ErrVoteConflictingVotes {
	return &ErrVoteConflictingVotes{
		VoteA: vote1,
		VoteB: vote2,
	}
}
func (voteSet *VoteSet) getVote(valIndex int32, blockKey string) (vote *Vote, ok bool) {
	if existing := voteSet.votes[valIndex]; existing != nil && existing.BlockID.Key() == blockKey {
		return existing, true
	}
	if existing := voteSet.votesByBlock[blockKey].getByIndex(valIndex); existing != nil {
		return existing, true
	}
	return nil, false
}

func (voteSet *VoteSet) SetPeerMaj23(peerID types2.PeerIDPretty, blockID BlockID) error {
	if voteSet == nil {
		panic("SetPeerMaj23() on nil VoteSet")
	}
	voteSet.mtx.Lock()
	defer voteSet.mtx.Unlock()

	blockKey := blockID.Key()

	// Make sure peer hasn't already told us something.
	if existing, ok := voteSet.peerMaj23s[peerID]; ok {
		if existing.Equals(blockID) {
			return nil // Nothing to do
		}
		return fmt.Errorf("setPeerMaj23: Received conflicting blockID from peer %v. Got %v, expected %v",
			peerID, blockID, existing)
	}
	voteSet.peerMaj23s[peerID] = blockID

	// Create .votesByBlock entry if needed.
	votesByBlock, ok := voteSet.votesByBlock[blockKey]
	if ok {
		// 说明是一个重放消息
		if votesByBlock.peerMaj23 {
			return nil // Nothing to do
		}
		votesByBlock.peerMaj23 = true
		// No need to copy votes, already there.
	} else {
		// 表明收到了2/3
		votesByBlock = newBlockVotes(true, voteSet.valSet.Size())
		voteSet.votesByBlock[blockKey] = votesByBlock
		// No need to copy votes, no votes to copy over.
	}
	return nil
}
func newBlockVotes(peerMaj23 bool, numValidators int) *blockVotes {
	return &blockVotes{
		peerMaj23: peerMaj23,
		bitArray:  libs.NewBitArray(numValidators),
		votes:     make([]*Vote, numValidators),
		sum:       0,
	}
}