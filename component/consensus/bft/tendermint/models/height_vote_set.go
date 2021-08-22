/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:13 上午
# @File : height_vote_set.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"fmt"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sync"
)

type RoundVoteSet struct {
	Prevotes   *VoteSet
	Precommits *VoteSet
}

type HeightVoteSet struct {
	chainID string
	height  uint64
	valSet  *ValidatorSet

	mtx               sync.Mutex
	round             int32                             // max tracked round
	roundVoteSets     map[int32]RoundVoteSet            // keys: [0...round]
	peerCatchupRounds map[types2.PeerIDPretty][]int32 // keys: peer.ID; values: at most 2 rounds
}

func NewHeightVoteSet(chainID string, height uint64, valSet *ValidatorSet) *HeightVoteSet {
	hvs := &HeightVoteSet{
		chainID: chainID,
	}
	hvs.Reset(height, valSet)
	return hvs
}
func (hvs *HeightVoteSet) Reset(height uint64, valSet *ValidatorSet) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()

	hvs.height = height
	hvs.valSet = valSet
	hvs.roundVoteSets = make(map[int32]RoundVoteSet)
	hvs.peerCatchupRounds = make(map[types2.PeerIDPretty][]int32)

	hvs.addRound(0)
	hvs.round = 0
}
var (
	ErrGotVoteFromUnwantedRound = errors.New(
		"peer has sent a vote that does not match our round for more than one round",
	)
)
func (hvs *HeightVoteSet) POLInfo() (polRound int32, polBlockID BlockID) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	for r := hvs.round; r >= 0; r-- {
		rvs := hvs.getVoteSet(r, types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE)
		polBlockID, ok := rvs.TwoThirdsMajority()
		if ok {
			return r, polBlockID
		}
	}
	return -1, BlockID{}
}
func (hvs *HeightVoteSet) AddVote(vote *Vote, peerID types2.PeerIDPretty) (added bool, err error) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	if !types.IsVoteTypeValid(vote.Type) {
		return
	}
	voteSet := hvs.getVoteSet(vote.Round, vote.Type)
	if voteSet == nil {
		if rndz := hvs.peerCatchupRounds[peerID]; len(rndz) < 2 {
			hvs.addRound(vote.Round)
			voteSet = hvs.getVoteSet(vote.Round, vote.Type)
			hvs.peerCatchupRounds[peerID] = append(rndz, vote.Round)
		} else {
			// punish peer
			err = ErrGotVoteFromUnwantedRound
			return
		}
	}
	added, err = voteSet.AddVote(vote)
	return
}
func (hvs *HeightVoteSet) SetRound(round int32) {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	newRound := libs.SafeSubInt32(hvs.round, 1)
	if hvs.round != 0 && (round < newRound) {
		panic("SetRound() must increment hvs.round")
	}
	for r := newRound; r <= round; r++ {
		if _, ok := hvs.roundVoteSets[r]; ok {
			continue // Already exists because peerCatchupRounds.
		}
		hvs.addRound(r)
	}
	hvs.round = round
}
func (hvs *HeightVoteSet) addRound(round int32) {
	if _, ok := hvs.roundVoteSets[round]; ok {
		panic("addRound() for an existing round")
	}
	// log.Debug("addRound(round)", "round", round)
	prevotes :=NewVoteSet(hvs.chainID, hvs.height, round, types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, hvs.valSet)
	precommits := NewVoteSet(hvs.chainID, hvs.height, round, types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, hvs.valSet)
	hvs.roundVoteSets[round] = RoundVoteSet{
		Prevotes:   prevotes,
		Precommits: precommits,
	}
}

func (hvs *HeightVoteSet) Precommits(round int32) *VoteSet {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.getVoteSet(round, types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT)
}
func (hvs *HeightVoteSet) Prevotes(round int32) *VoteSet {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	return hvs.getVoteSet(round, types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE)
}
// If a peer claims that it has 2/3 majority for given blockKey, call this.
// NOTE: if there are too many peers, or too much peer churn,
// this can cause memory issues.
// TODO: implement ability tve peers too
func (hvs *HeightVoteSet) SetPeerMaj23(
	round int32,
	voteType types3.SignedMsgType,
	peerID types2.PeerIDPretty,
	blockID BlockID) error {
	hvs.mtx.Lock()
	defer hvs.mtx.Unlock()
	if !types.IsVoteTypeValid(voteType) {
		return fmt.Errorf("setPeerMaj23: Invalid vote type %X", voteType)
	}
	// 根据类型挑选定义着的voteSet
	voteSet := hvs.getVoteSet(round, voteType)
	if voteSet == nil {
		return nil // something we don't know about yet
	}
	// 表明,这个block 收到了 2/3 的msg,可能是prevote,也可能是precommit
	return voteSet.SetPeerMaj23(peerID, blockID)
}
func (hvs *HeightVoteSet) getVoteSet(round int32, voteType types3.SignedMsgType) *VoteSet {
	rvs, ok := hvs.roundVoteSets[round]
	if !ok {
		return nil
	}
	switch voteType {
	case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
		return rvs.Prevotes
	case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
		return rvs.Precommits
	default:
		panic(fmt.Sprintf("Unexpected vote type %X", voteType))
	}
}
