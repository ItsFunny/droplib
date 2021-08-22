/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:08 上午 
# @File : proposal.go
# @Description : 
# @Attention : 
*/
package models

import (
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/libs/protoio"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"time"
)

type Proposal struct {
	Type      types.SignedMsgType
	Height    uint64     `json:"height"`
	Round     int32     `json:"round"`     // there can not be greater than 2_147_483_647 rounds
	POLRound  int32     `json:"pol_round"` // -1 if null.
	BlockID   BlockID   `json:"block_id"`
	Timestamp time.Time `json:"timestamp"`
	Signature []byte    `json:"signature"`
}

func ProposalSignBytes(chainID string, p *types.ProposalProto) []byte {
	pb := CanonicalizeProposal(chainID, p)
	bz, err := protoio.MarshalDelimited(&pb)
	if err != nil {
		panic(err)
	}

	return bz
}
func CanonicalizeProposal(chainID string, proposal *types.ProposalProto) types.CanonicalProposalProto {
	return types.CanonicalProposalProto{
		Type:      types.SignedMsgType_SIGNED_MSG_TYPE_PROPOSAL,
		Height:    proposal.Height,       // encoded as sfixed64
		Round:     int64(proposal.Round), // encoded as sfixed64
		PolRound:  int64(proposal.PolRound),
		BlockId:   CanonicalizeBlockID(proposal.BlockId),
		Timestamp: proposal.Timestamp,
		ChainId:   chainID,
	}
}
func (p *Proposal) ToBytes() []byte {
	return nil
}

func NewProposal(height int64, round int32, polRound int32, blockID BlockID) *Proposal {
	return &Proposal{
		Type:      types.SignedMsgType_SIGNED_MSG_TYPE_PROPOSAL,
		Height:    uint64(height),
		Round:     round,
		BlockID:   blockID,
		POLRound:  polRound,
		Timestamp: libs.Now(),
	}
}

func (p *Proposal) ToProto() *types.ProposalProto {
	if p == nil {
		return &types.ProposalProto{}
	}
	pb := new(types.ProposalProto)

	pb.BlockId = p.BlockID.ToProto()
	pb.Type = p.Type
	pb.Height = int64(p.Height)
	pb.Round = p.Round
	pb.PolRound = p.POLRound
	pb.Timestamp = commonutils.ConvtTime2GoogleTime(p.Timestamp)
	pb.Signature = p.Signature

	return pb
}