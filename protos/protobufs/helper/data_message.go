/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/5 9:28 上午
# @File : data_message.go
# @Description :
# @Attention :
*/
package helper

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	consensus2 "github.com/hyperledger/fabric-droplib/protos/protobufs/consensus"
)

var (
	_ services.IChannelMessageWrapper = (*ConsensusVoteChannelMessageWrapper)(nil)
	_ services.IChannelMessageWrapper = (*ConsensusStateChannelMessageWrapper)(nil)
	_ services.IChannelMessageWrapper = (*ConsensusDataChannelMessageWrapper)(nil)
	_ services.IChannelMessageWrapper = (*ConsensusVoteSetBitsChannelMessageWrapper)(nil)
)

func (this *ConsensusVoteChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &ConsensusVoteChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *consensus2.VoteProtoWrapper:

		r.Sum = &ConsensusVoteChannelDataMessageWrapper_Vote{Vote: msg}
	default:
		return nil, errors.New("未知类型")
	}

	return proto.Marshal(r)
}
func (this *ConsensusVoteChannelMessageWrapper) Decode(bytes []byte) (proto.Message, services.IMessage, error) {
	r := &ConsensusVoteChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil, nil, err
	}
	switch msg := r.Sum.(type) {
	case *ConsensusVoteChannelDataMessageWrapper_Vote:
		v := msg.Vote.Vote
		fromProto, err := models.VoteFromProto(v)
		if nil != err {
			return nil, nil, err
		}
		logic := models.VoteMessage{
			Vote: fromProto,
		}
		return msg.Vote, &logic, nil
	default:
		return nil, nil, errors.New("asd")
	}
}

// ////////
/*
  NewRoundStepProtoWrapper
  NewValidBlockProtoWrapper
  HasVoteProtoWrapper
  VoteSetMaj23ProtoWrapper
*/
func (this *ConsensusStateChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &ConsensusStateChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *consensus2.NewRoundStepProtoWrapper:
		r.Sum = &ConsensusStateChannelDataMessageWrapper_NewRoundStep{NewRoundStep: msg}
	case *consensus2.NewValidBlockProtoWrapper:
		r.Sum = &ConsensusStateChannelDataMessageWrapper_NewValidBlock{
			NewValidBlock: msg,
		}
	case *consensus2.HasVoteProtoWrapper:
		r.Sum = &ConsensusStateChannelDataMessageWrapper_HasVote{HasVote: msg}
	case *consensus2.VoteSetMaj23ProtoWrapper:
		r.Sum = &ConsensusStateChannelDataMessageWrapper_VoteSetMaj23{VoteSetMaj23: msg}
	default:
		return nil, errors.New("未知类型")
	}

	return proto.Marshal(r)
}

/*
  NewRoundStepProtoWrapper
  NewValidBlockProtoWrapper
  HasVoteProtoWrapper
  VoteSetMaj23ProtoWrapper
*/
func (this *ConsensusStateChannelMessageWrapper) Decode(bytes []byte) (proto.Message, services.IMessage, error) {
	r := &ConsensusStateChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil, nil, err
	}
	var (
		protoMsg proto.Message
		logicMsg services.IMessage
		err      error
	)
	switch msg := r.Sum.(type) {
	case *ConsensusStateChannelDataMessageWrapper_NewRoundStep:
		protoMsg = msg.NewRoundStep
		logicMsg, err = models.NewRoundStepFromProto(msg.NewRoundStep)
	case *ConsensusStateChannelDataMessageWrapper_NewValidBlock:
		protoMsg = msg.NewValidBlock
		logicMsg, err = models.NewValidBlockFromProto(msg.NewValidBlock)
	case *ConsensusStateChannelDataMessageWrapper_HasVote:
		protoMsg = msg.HasVote
		logicMsg, _ = models.HasVoteFromProto(msg.HasVote)
	case *ConsensusStateChannelDataMessageWrapper_VoteSetMaj23:
		protoMsg = msg.VoteSetMaj23
		logicMsg, err = models.VoteSetMAJ23FromProto(msg.VoteSetMaj23)
	default:
		return nil, nil, errors.New("asd")
	}
	return protoMsg, logicMsg, err
}

// /////////////

/*
  ProposalProtoWrapper
  ProposalPOLProtoWrapper
  BlockPartProtoWrapper
*/

func (this *ConsensusDataChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &ConsensusDataChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *consensus2.ProposalProtoWrapper:
		r.Sum = &ConsensusDataChannelDataMessageWrapper_Proposal{Proposal: msg}
	case *consensus2.ProposalPOLProtoWrapper:
		r.Sum = &ConsensusDataChannelDataMessageWrapper_ProposalPol{
			ProposalPol: msg,
		}
	case *consensus2.BlockPartProtoWrapper:
		r.Sum = &ConsensusDataChannelDataMessageWrapper_BlockPart{BlockPart: msg}
	default:
		return nil, errors.New("未知类型")
	}

	return proto.Marshal(r)
}

/*
  ProposalProtoWrapper
  ProposalPOLProtoWrapper
  BlockPartProtoWrapper
*/
func (this *ConsensusDataChannelMessageWrapper) Decode(bytes []byte) (proto.Message, services.IMessage, error) {
	r := &ConsensusDataChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil, nil, err
	}
	var (
		protoMsg proto.Message
		logicMsg services.IMessage
		err      error
	)
	switch msg := r.Sum.(type) {
	case *ConsensusDataChannelDataMessageWrapper_Proposal:
		protoMsg = msg.Proposal
		logicMsg, err = models.ProposalFromProto(msg.Proposal)
	case *ConsensusDataChannelDataMessageWrapper_ProposalPol:
		protoMsg = msg.ProposalPol
		logicMsg, err = models.ProposalPolFromProto(msg.ProposalPol)
	case *ConsensusDataChannelDataMessageWrapper_BlockPart:
		protoMsg = msg.BlockPart
		logicMsg, err = models.BlockPartFromProto(msg.BlockPart)
	default:
		return nil, nil, errors.New("asd")
	}
	return protoMsg, logicMsg, err
}

// ////
/*
	VoteSetBitsProtoWrapper
*/
func (this *ConsensusVoteSetBitsChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &ConsensusVoteSetBitsChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *consensus2.VoteSetBitsProtoWrapper:
		r.Sum = &ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits{VoteSetBits: msg}
	default:
		return nil, errors.New("未知类型")
	}

	return proto.Marshal(r)
}

func (this *ConsensusVoteSetBitsChannelMessageWrapper) Decode(bytes []byte) (proto.Message, services.IMessage, error) {
	r := &ConsensusVoteSetBitsChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil, nil, err
	}
	var (
		protoMsg proto.Message
		logicMsg services.IMessage
		err      error
	)
	switch msg := r.Sum.(type) {
	case *ConsensusVoteSetBitsChannelDataMessageWrapper_VoteSetBits:
		protoMsg = msg.VoteSetBits
		logicMsg, err = models.VoteSetBitFromProto(msg.VoteSetBits)
	default:
		return nil, nil, errors.New("asd")
	}
	return protoMsg, logicMsg, err
}
