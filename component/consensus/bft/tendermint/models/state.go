/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:48 上午
# @File : state.go
# @Description :
# @Attention :
*/
package models

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	types3 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	"time"
)

type MsgInfo struct {
	Msg    services.IMessage `json:"msg"`
	PeerID types2.NodeID   `json:"peer_key"`
}

type NewRoundStepMessage struct {
	Height                uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Round                 int32 `protobuf:"varint,2,opt,name=round,proto3" json:"round,omitempty"`
	Step                  types3.RoundStepType
	SecondsSinceStartTime int64 `protobuf:"varint,4,opt,name=seconds_since_start_time,json=secondsSinceStartTime,proto3" json:"seconds_since_start_time,omitempty"`
	LastCommitRound       int32 `protobuf:"varint,5,opt,name=last_commit_round,json=lastCommitRound,proto3" json:"last_commit_round,omitempty"`
}

func (m *NewRoundStepMessage) ValidateBasic() error {
	return nil
}

// ValidateHeight validates the height given the chain's initial height.
func (m *NewRoundStepMessage) ValidateHeight(initialHeight uint64) error {
	if m.Height < initialHeight {
		return fmt.Errorf("invalid Height %v (lower than initial height %v)",
			m.Height, initialHeight)
	}
	if m.Height == initialHeight && m.LastCommitRound != -1 {
		return fmt.Errorf("invalid LastCommitRound %v (must be -1 for initial height %v)",
			m.LastCommitRound, initialHeight)
	}
	if m.Height > initialHeight && m.LastCommitRound < 0 {
		return fmt.Errorf("LastCommitRound can only be negative for initial height %v", // nolint
			initialHeight)
	}
	return nil
}

type PeerRoundState struct {
	Height    uint64               `json:"height"` // Height peer is at
	Round     int32               `json:"round"`  // Round peer is at, -1 if unknown.
	Step      types3.RoundStepType `json:"step"`   // Step peer is at
	StartTime time.Time           `json:"start_time"`

	// FIXME
	Proposal bool `json:"proposal"`

	// ///
	ProposalPOLRound           int32          `json:"proposal_pol_round"`
	ProposalBlockPartSetHeader PartSetHeader  `json:"proposal_block_part_set_header"`
	ProposalBlockParts         *libs.BitArray `json:"proposal_block_parts"`

	ProposalPOL     *libs.BitArray `json:"proposal_pol"`
	Prevotes        *libs.BitArray `json:"prevotes"`          // All votes peer has for this round
	Precommits      *libs.BitArray `json:"precommits"`        // All precommits peer has for this round
	LastCommitRound int32          `json:"last_commit_round"` // Round of commit for last height. -1 if none.
	LastCommit      *libs.BitArray `json:"last_commit"`       // All commit precommits of commit for last height.

	// /////
	CatchupCommitRound int32 `json:"catchup_commit_round"`

	// ///
	CatchupCommit *libs.BitArray `json:"catchup_commit"`
}

type RoundState struct {
	Height    uint64               `json:"height"` // Height we are working on
	Round     int32               `json:"round"`
	Step      types3.RoundStepType `json:"step"`
	StartTime time.Time           `json:"start_time"`

	// Subjective time when +2/3 precommits for Block at Round were found
	CommitTime         time.Time     `json:"commit_time"`
	Validators         *ValidatorSet `json:"validators"`
	Proposal           *Proposal     `json:"proposal"`
	ProposalBlockParts *PartSet      `json:"proposal_block_parts"`
	// fabric block
	ProposalBlock *TendermintBlockWrapper `json:"proposalBlock"`

	LockedRound int32                   `json:"locked_round"`
	LockedBlock *TendermintBlockWrapper `json:"locked_block"`

	LockedBlockParts *PartSet `json:"locked_block_parts"`

	// Last known round with POL for non-nil valid block.
	ValidRound int32                   `json:"valid_round"`
	ValidBlock *TendermintBlockWrapper `json:"valid_block"` // Last known block of POL mentioned above.

	// Last known block parts of POL mentioned above.
	ValidBlockParts           *PartSet       `json:"valid_block_parts"`
	Votes                     *HeightVoteSet `json:"votes"`
	CommitRound               int32          `json:"commit_round"` //
	LastCommit                *VoteSet       `json:"last_commit"`  // Last precommits at Height-1
	LastValidators            *ValidatorSet  `json:"last_validators"`
	TriggeredTimeoutPrecommit bool           `json:"triggered_timeout_precommit"`
}

// func (rs *RoundState) RoundStateEvent() events2.EventDataRoundState {
// 	return events2.EventDataRoundState{
// 		Height: rs.Height,
// 		Round:  rs.Round,
// 		Step:   rs.Step.String(),
// 	}
// }

// func (rs *RoundState) NewRoundEvent() events2.EventDataNewRound {
// 	addr := rs.Validators.GetProposer().Address
// 	idx, _ := rs.Validators.GetByAddress(addr)
//
// 	return events2.EventDataNewRound{
// 		Height: rs.Height,
// 		Round:  rs.Round,
// 		Step:   rs.Step.String(),
// 		Proposer: ValidatorInfo{
// 			Address: addr,
// 			Index:   idx,
// 		},
// 	}
// }

// func (rs *RoundState) CompleteProposalEvent() events2.EventDataCompleteProposal {
// 	// We must construct BlockID from ProposalBlock and ProposalBlockParts
// 	// cs.Proposal is not guaranteed to be set when this function is called
// 	blockID := BlockID{
// 		Hash:          rs.ProposalBlock.Hash(),
// 		PartSetHeader: rs.ProposalBlockParts.Header(),
// 	}
//
// 	return events2.EventDataCompleteProposal{
// 		Height:  rs.Height,
// 		Round:   rs.Round,
// 		Step:    rs.Step.String(),
// 		BlockID: blockID,
// 	}
// }