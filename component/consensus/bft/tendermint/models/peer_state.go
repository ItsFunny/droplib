/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:35 上午
# @File : peer_state.go
# @Description :
# @Attention :
*/
package models

import (
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	libsync "github.com/hyperledger/fabric-droplib/libs/sync"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
	"time"
)


type VoteSetReader interface {
	GetHeight() uint64
	GetRound() int32
	Type() byte
	Size() int
	BitArray() *libs.BitArray
	GetByIndex(int32) *Vote
	IsCommit() bool
}

type PeerState struct {
	NodeID      types2.PeerIDPretty
	logger      v2.Logger
	mtx         sync.RWMutex
	running     bool
	BroadcastWG sync.WaitGroup
	Stats       *peerStateStats `json:"stats"`

	PeerRoudState PeerRoundState

	Closer *libsync.Closer
}

func (ps *PeerState) PickVoteToSend(votes VoteSetReader) (*Vote, bool) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if votes.Size() == 0 {
		return nil, false
	}

	var (
		height    = votes.GetHeight()
		round     = votes.GetRound()
		votesType = types3.SignedMsgType(votes.Type())
		size      = votes.Size()
	)

	// lazily set data using 'votes'
	if votes.IsCommit() {
		ps.ensureCatchupCommitRound(height, round, size)
	}

	ps.ensureVoteBitArrays(height, size)

	psVotes := ps.getVoteBitArray(height, round, votesType)
	if psVotes == nil {
		return nil, false // not something worth sending
	}

	if index, ok := votes.BitArray().Sub(psVotes).PickRandom(); ok {
		return votes.GetByIndex(int32(index)), true
	}

	return nil, false
}
func (ps *PeerState) GetRoundState() *PeerRoundState {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	PeerRoudState := ps.PeerRoudState // copy
	return &PeerRoudState
}

// SetRunning sets the running state of the peer.
func (ps *PeerState) SetRunning(v bool) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.running = v
}
func (ps *PeerState) IsRunning() bool {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()

	return ps.running
}
func NewPeerState(logger v2.Logger, peerID types2.NodeID) *PeerState {
	return &PeerState{
		NodeID: types2.PeerIDPretty(peer.ID(peerID).Pretty()),
		logger: logger,
		Closer: libsync.NewCloser(),
		PeerRoudState: PeerRoundState{
			Round:              -1,
			ProposalPOLRound:   -1,
			LastCommitRound:    -1,
			CatchupCommitRound: -1,
		},
		Stats: &peerStateStats{},
	}
}

// peerStateStats holds internal statistics for a peer.
type peerStateStats struct {
	Votes      int `json:"votes"`
	BlockParts int `json:"block_parts"`
}

func (ps *PeerState) RecordVote() int {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.Stats.Votes++

	return ps.Stats.Votes
}

// RecordBlockPart increments internal block part related statistics for this peer.
// It returns the total number of added block parts.
func (ps *PeerState) RecordBlockPart() int {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.Stats.BlockParts++
	return ps.Stats.BlockParts
}

func (ps *PeerState) ApplyNewRoundStepMessage(msg *NewRoundStepMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// ignore duplicates or decreases
	if CompareHRS(msg.Height, msg.Round, msg.Step, ps.PeerRoudState.Height, ps.PeerRoudState.Round, ps.PeerRoudState.Step) <= 0 {
		return
	}

	var (
		// 当前process的height
		psHeight = ps.PeerRoudState.Height
		// 当前process的round
		psRound = ps.PeerRoudState.Round
		// 当前process的已经commit的round
		psCatchupCommitRound = ps.PeerRoudState.CatchupCommitRound

		psCatchupCommit = ps.PeerRoudState.CatchupCommit
		startTime       = libs.Now().Add(-1 * time.Duration(msg.SecondsSinceStartTime) * time.Second)
	)

	// 节点的step变更是通过msg感知到的
	// 重新赋值
	ps.PeerRoudState.Height = msg.Height
	ps.PeerRoudState.Round = msg.Round
	ps.PeerRoudState.Step = msg.Step
	ps.PeerRoudState.StartTime = startTime

	if psHeight != msg.Height || psRound != msg.Round {
		ps.PeerRoudState.Proposal = false
		// ps.PeerRoudState.ProposalBlockPartSetHeader = types.PartSetHeader{}
		// ps.PeerRoudState.ProposalBlockParts = nil
		ps.PeerRoudState.ProposalPOLRound = -1
		ps.PeerRoudState.ProposalPOL = nil

		// we'll update the BitArray capacity later
		// 因为这一步置空了,所以后续都会有ensure
		ps.PeerRoudState.Prevotes = nil
		ps.PeerRoudState.Precommits = nil
	}

	if psHeight == msg.Height && psRound != msg.Round && msg.Round == psCatchupCommitRound {
		// Peer caught up to CatchupCommitRound.
		// Preserve psCatchupCommit!
		// NOTE: We prefer to use PeerRoudState.Precommits if
		// pr.Round matches pr.CatchupCommitRound.
		ps.PeerRoudState.Precommits = psCatchupCommit
	}

	if psHeight != msg.Height {
		// shift Precommits to LastCommit
		if psHeight+1 == msg.Height && psRound == msg.LastCommitRound {
			ps.PeerRoudState.LastCommitRound = msg.LastCommitRound
			ps.PeerRoudState.LastCommit = ps.PeerRoudState.Precommits
		} else {
			ps.PeerRoudState.LastCommitRound = msg.LastCommitRound
			ps.PeerRoudState.LastCommit = nil
		}

		// we'll update the BitArray capacity later
		ps.PeerRoudState.CatchupCommitRound = -1
		ps.PeerRoudState.CatchupCommit = nil
	}
}

func (ps *PeerState) ApplyNewValidBlockMessage(msg *NewValidBlockMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.PeerRoudState.Height != msg.Height {
		return
	}
	// 对于prevote的2/3 只有round 匹配上才会继续,否则如果是commit的话,则会直接更新
	// 因为可能会多轮才会commit 一个block
	if ps.PeerRoudState.Round != msg.Round && !msg.IsCommit {
		return
	}

	ps.PeerRoudState.ProposalBlockPartSetHeader = msg.BlockPartSetHeader
	ps.PeerRoudState.ProposalBlockParts = msg.BlockParts
}

func (ps *PeerState) ApplyHasVoteMessage(msg *HasVoteMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	if ps.PeerRoudState.Height != msg.Height {
		return
	}
}

func (ps *PeerState) SetHasProposal(proposal *Proposal) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.PeerRoudState.Height != proposal.Height || ps.PeerRoudState.Round != proposal.Round {
		return
	}

	if ps.PeerRoudState.Proposal {
		return
	}

	ps.PeerRoudState.Proposal = true

	// ps.PeerRoudState.ProposalBlockParts is set due to NewValidBlockMessage
	// FIXME
	// if ps.PeerRoudState.ProposalBlockParts != nil {
	// 	return
	// }

	// ps.PeerRoudState.ProposalBlockPartSetHeader = proposal.BlockID.PartSetHeader
	// ps.PeerRoudState.ProposalBlockParts = libs.NewBitArray(int(proposal.BlockID.PartSetHeader.Total))
	ps.PeerRoudState.ProposalPOLRound = proposal.POLRound
	ps.PeerRoudState.ProposalPOL = nil // Nil until ProposalPOLMessage received.
}

func (ps *PeerState) ApplyProposalPOLMessage(msg *ProposalPOLMessage) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.PeerRoudState.Height != msg.Height {
		return
	}
	if ps.PeerRoudState.ProposalPOLRound != msg.ProposalPOLRound {
		return
	}

	// TODO: Merge onto existing ps.PeerRoudState.ProposalPOL?
	// We might have sent some prevotes in the meantime.
	ps.PeerRoudState.ProposalPOL = msg.ProposalPOL
}
func (ps *PeerState) SetHasProposalBlockPart(height uint64, round int32, index int) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.PeerRoudState.Height != height || ps.PeerRoudState.Round != round {
		return
	}

	ps.PeerRoudState.ProposalBlockParts.SetIndex(index, true)
}

func (ps *PeerState) EnsureVoteBitArrays(height uint64, numValidators int) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.ensureVoteBitArrays(height, numValidators)
}
func (ps *PeerState) ensureVoteBitArrays(height uint64, numValidators int) {
	if ps.PeerRoudState.Height == height {
		if ps.PeerRoudState.Prevotes == nil {
			ps.PeerRoudState.Prevotes = libs.NewBitArray(numValidators)
		}
		if ps.PeerRoudState.Precommits == nil {
			ps.PeerRoudState.Precommits = libs.NewBitArray(numValidators)
		}
		if ps.PeerRoudState.CatchupCommit == nil {
			ps.PeerRoudState.CatchupCommit = libs.NewBitArray(numValidators)
		}
		if ps.PeerRoudState.ProposalPOL == nil {
			ps.PeerRoudState.ProposalPOL = libs.NewBitArray(numValidators)
		}
	} else if ps.PeerRoudState.Height == height+1 {
		if ps.PeerRoudState.LastCommit == nil {
			ps.PeerRoudState.LastCommit = libs.NewBitArray(numValidators)
		}
	}
}

func (ps *PeerState) SetHasVote(vote *Vote) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.setHasVote(vote.Height, vote.Round, vote.Type, vote.ValidatorIndex)
}

func (ps *PeerState) setHasVote(height uint64, round int32, voteType types3.SignedMsgType, index int32) {
	// logger := ps.logger.With(
	// 	"peerH/R", fmt.Sprintf("%d/%d", ps.PeerRoudState.Height, ps.PeerRoudState.Round),
	// 	"H/R", fmt.Sprintf("%d/%d", height, round),
	// )
	logger := ps.logger

	logger.Debug("setHasVote", "type", voteType, "index", index)

	// NOTE: some may be nil BitArrays -> no side effects
	psVotes := ps.getVoteBitArray(height, round, voteType)
	if psVotes != nil {
		// 则更新这个节点也已经投票了
		psVotes.SetIndex(int(index), true)
	}
}

func (ps *PeerState) getVoteBitArray(height uint64, round int32, votesType types3.SignedMsgType) *libs.BitArray {
	if !types.IsVoteTypeValid(votesType) {
		return nil
	}

	if ps.PeerRoudState.Height == height {
		// 如果是正常的process,round 匹配,则根据 msg的类型判断获取 这个节点已经收到的相关数据
		if ps.PeerRoudState.Round == round {
			switch votesType {
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
				return ps.PeerRoudState.Prevotes

			case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
				return ps.PeerRoudState.Precommits
			}
		}

		// 说明高度不同,
		if ps.PeerRoudState.CatchupCommitRound == round {
			// 则代表的是,收到的是其他节点的重放消息(例如A网络堵塞,A重新作为proposer之后,发出提案,但是之前改节点已经commit
			// ....?
			switch votesType {
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
				return nil

			case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
				return ps.PeerRoudState.CatchupCommit
			}
		}

		if ps.PeerRoudState.ProposalPOLRound == round {
			switch votesType {
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
				return ps.PeerRoudState.ProposalPOL

			case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
				return nil
			}
		}

		return nil
	}
	if ps.PeerRoudState.Height == height+1 {
		if ps.PeerRoudState.LastCommitRound == round {
			switch votesType {
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
				return nil

			case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
				return ps.PeerRoudState.LastCommit
			}
		}

		return nil
	}

	return nil
}

func (ps *PeerState) ApplyVoteSetBitsMessage(msg *VoteSetBitsMessage, ourVotes *libs.BitArray) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	nodeVotesInfo := ps.getVoteBitArray(msg.Height, msg.Round, msg.Type)
	if nodeVotesInfo != nil {
		if ourVotes == nil {
			nodeVotesInfo.Update(msg.Votes)
		} else {
			otherVotes := nodeVotesInfo.Sub(ourVotes)
			hasVotes := otherVotes.Or(msg.Votes)
			nodeVotesInfo.Update(hasVotes)
		}
	}
}
func (ps *PeerState) InitProposalBlockParts(partSetHeader PartSetHeader) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.PeerRoudState.ProposalBlockParts != nil {
		return
	}

	ps.PeerRoudState.ProposalBlockPartSetHeader = partSetHeader
	ps.PeerRoudState.ProposalBlockParts = libs.NewBitArray(int(partSetHeader.Total))
}

func (ps *PeerState) ensureCatchupCommitRound(height uint64, round int32, numValidators int) {
	if ps.PeerRoudState.Height != height {
		return
	}

	/*
		NOTE: This is wrong, 'round' could change.
		e.g. if orig round is not the same as block LastCommit round.
		if ps.CatchupCommitRound != -1 && ps.CatchupCommitRound != round {
			panic(fmt.Sprintf(
				"Conflicting CatchupCommitRound. Height: %v,
				Orig: %v,
				New: %v",
				height,
				ps.CatchupCommitRound,
				round))
		}
	*/

	if ps.PeerRoudState.CatchupCommitRound == round {
		return // Nothing to do!
	}

	ps.PeerRoudState.CatchupCommitRound = round
	if round == ps.PeerRoudState.Round {
		ps.PeerRoudState.CatchupCommit = ps.PeerRoudState.Precommits
	} else {
		ps.PeerRoudState.CatchupCommit = libs.NewBitArray(numValidators)
	}
}
func CompareHRS(h1 uint64, r1 int32, s1 types.RoundStepType, h2 uint64, r2 int32, s2 types.RoundStepType) int {
	if h1 < h2 {
		return -1
	} else if h1 > h2 {
		return 1
	}
	if r1 < r2 {
		return -1
	} else if r1 > r2 {
		return 1
	}
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}
