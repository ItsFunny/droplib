/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:07 上午
# @File : state.go
# @Description :
# @Attention :
*/
package events

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"time"
)

// FIXME 命名
type EventChanData struct {
	Data interface{}
}
type EventVote struct {
	Type             types3.SignedMsgType  `json:"type"`
	Height           int64                 `json:"height"`
	Round            int32                 `json:"round"`    // assume there will not be greater than 2_147_483_647 rounds
	BlockID          models2.BlockID               `json:"block_id"` // zero if vote is nil.
	Timestamp        time.Time             `json:"timestamp"`
	ValidatorAddress types2.PeerIDPretty `json:"validator_address"`
	ValidatorIndex   int32                 `json:"validator_index"`
	Signature        []byte                `json:"signature"`
}


type EventRoundState struct {
	Height    int64         `json:"height"` // Height we are working on
	Round     int32         `json:"round"`
	Step      types.RoundStepType `json:"step"`
	StartTime time.Time     `json:"start_time"`

	// Subjective time when +2/3 precommits for Block at Round were found
	CommitTime time.Time     `json:"commit_time"`
	Validators *models2.ValidatorSet `json:"validators"`
	Proposal   *models2.Proposal     `json:"proposal"`
	ProposalBlockParts *models2.PartSet      `json:"proposal_block_parts"`
	// fabric block
	ProposalBlock    *models2.TendermintBlockWrapper `json:"proposalBlock"`

	LockedRound      int32                      `json:"locked_round"`
	LockedBlock      *models2.TendermintBlockWrapper `json:"locked_block"`

	LockedBlockParts *models2.PartSet                   `json:"locked_block_parts"`

	// Last known round with POL for non-nil valid block.
	ValidRound int32                      `json:"valid_round"`
	ValidBlock *models2.TendermintBlockWrapper `json:"valid_block"` // Last known block of POL mentioned above.

	// Last known block parts of POL mentioned above.
	ValidBlockParts           *models2.PartSet                  `json:"valid_block_parts"`
	Votes                     *models2.HeightVoteSet `json:"votes"`
	CommitRound               int32                     `json:"commit_round"` //
	LastCommit                *models2.VoteSet                  `json:"last_commit"`  // Last precommits at Height-1
	LastValidators            *models2.ValidatorSet             `json:"last_validators"`
	TriggeredTimeoutPrecommit bool                      `json:"triggered_timeout_precommit"`
}


type EventNewValidBlock struct {
	Height    int64         `json:"height"` // Height we are working on
	Round     int32         `json:"round"`
	ProposalBlockParts *models2.PartSet      `json:"proposal_block_parts"`
	Step      types.RoundStepType `json:"step"`
}
