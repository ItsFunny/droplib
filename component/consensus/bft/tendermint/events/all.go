/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 4:23 下午
# @File : all.go
# @Description :
# @Attention :
*/
package events

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
)

// NOTE: This goes into the replay WAL
type EventDataRoundState struct {
	Height uint64  `json:"height"`
	Round  int32  `json:"round"`
	Step   string `json:"step"`
}

type EventDataNewRound struct {
	Height uint64  `json:"height"`
	Round  int32  `json:"round"`
	Step   string `json:"step"`

	Proposer models.ValidatorInfo `json:"proposer"`
}
type EventDataCompleteProposal struct {
	Height uint64  `json:"height"`
	Round  int32  `json:"round"`
	Step   string `json:"step"`

	BlockID models.BlockID `json:"block_id"`
}
type EventDataVote struct {
	Vote *models.Vote
}


type EventDataNewBlock struct {
	Block *models.TendermintBlockWrapper `json:"block"`
}

type EventDataNewBlockHeader struct {
	Header models.TendermintBlockHeader `json:"header"`

	NumTxs int64 `json:"num_txs"` // Number of txs in a block
}
type EventDataNewEvidence struct {
	Evidence transfer.Evidence `json:"evidence"`

	Height int64 `json:"height"`
}
// All txs fire EventDataTx
type EventDataTx struct {
	// abci.TxResult
	Height int64             `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Index  uint32            `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
	Tx     []byte            `protobuf:"bytes,3,opt,name=tx,proto3" json:"tx,omitempty"`
}

type EventDataValidatorSetUpdates struct {
	ValidatorUpdates []*models.Validator `json:"validator_updates"`
}