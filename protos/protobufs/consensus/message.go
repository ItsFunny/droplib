/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/4 4:03 下午
# @File : message.go
# @Description :
# @Attention :
*/
package consensus

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

// Wrap implements the p2p Wrapper interface and wraps a consensus proto message.
func (m *MessageProtoWrapper) Wrap(pb proto.Message) error {
	switch msg := pb.(type) {
	// case *NewRoundStep:
	// 	m.Sum = &MessageProtoWrapper_NewRoundStep{NewRoundStep: msg}

	// case *NewValidBlock:
	// 	m.Sum = &Message_NewValidBlock{NewValidBlock: msg}
	//
	// case *Proposal:
	// 	m.Sum = &Message_Proposal{Proposal: msg}
	//
	// case *ProposalPOL:
	// 	m.Sum = &Message_ProposalPol{ProposalPol: msg}
	//
	// case *BlockPart:
	// 	m.Sum = &Message_BlockPart{BlockPart: msg}
	//
	// case *Vote:
	// 	m.Sum = &Message_Vote{Vote: msg}
	//
	// case *HasVote:
	// 	m.Sum = &Message_HasVote{HasVote: msg}
	//
	// case *VoteSetMaj23:
	// 	m.Sum = &Message_VoteSetMaj23{VoteSetMaj23: msg}
	//
	// case *VoteSetBits:
	// 	m.Sum = &Message_VoteSetBits{VoteSetBits: msg}

	default:
		return fmt.Errorf("unknown message: %T", msg)
	}

	return nil
}