/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 12:42 下午
# @File : message.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"fmt"
	types3 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	types2 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

type NewValidBlockMessage struct {
	Height             uint64
	Round              int32
	BlockPartSetHeader PartSetHeader
	BlockParts         *libs.BitArray
	IsCommit           bool
}

func (m *NewValidBlockMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if err := m.BlockPartSetHeader.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockPartSetHeader: %v", err)
	}
	if m.BlockParts.Size() == 0 {
		return errors.New("empty blockParts")
	}
	if m.BlockParts.Size() != int(m.BlockPartSetHeader.Total) {
		return fmt.Errorf("blockParts bit array size %d not equal to BlockPartSetHeader.Total %d",
			m.BlockParts.Size(),
			m.BlockPartSetHeader.Total)
	}
	if m.BlockParts.Size() > int(MaxBlockPartsCount) {
		return fmt.Errorf("blockParts bit array is too big: %d, max: %d", m.BlockParts.Size(), MaxBlockPartsCount)
	}
	return nil
}


// HasVoteMessage is sent to indicate that a particular vote has been received.
type HasVoteMessage struct {
	Height uint64
	Round  int32
	Type   types2.SignedMsgType
	Index  int32
}

// ValidateBasic performs basic validation.
func (m *HasVoteMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if !types3.IsVoteTypeValid(m.Type) {
		return errors.New("invalid Type")
	}
	if m.Index < 0 {
		return errors.New("negative Index")
	}
	return nil
}

// String returns a string representation.
func (m *HasVoteMessage) String() string {
	return fmt.Sprintf("[HasVote VI:%v V:{%v/%02d/%v}]", m.Index, m.Height, m.Round, m.Type)
}

type VoteSetMaj23Message struct {
	Height  int64
	Round   int32
	Type    types2.SignedMsgType
	BlockID BlockID
}

func (v VoteSetMaj23Message) ValidateBasic() error {
	return nil
}




//////////// data-channel
type ProposalMessage struct {
	Proposal *Proposal
	ExtraData []byte `json:"extraData"`
}

func (p ProposalMessage) ValidateBasic() error {
	return nil
}



// ProposalPOLMessage is sent when a previous proposal is re-proposed.
type ProposalPOLMessage struct {
	Height           uint64
	ProposalPOLRound int32
	ProposalPOL      *libs.BitArray
}

// ValidateBasic performs basic validation.
func (m *ProposalPOLMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.ProposalPOLRound < 0 {
		return errors.New("negative ProposalPOLRound")
	}
	if m.ProposalPOL.Size() == 0 {
		return errors.New("empty ProposalPOL bit array")
	}
	if m.ProposalPOL.Size() > 10000 {
		return fmt.Errorf("proposalPOL bit array is too big: %d, max: %d", m.ProposalPOL.Size(), 10000)
	}
	return nil
}


type BlockPartMessage struct {
	Height uint64
	Round  int32
	Part   *Part
}

// ValidateBasic performs basic validation.
func (m *BlockPartMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if err := m.Part.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong Part: %v", err)
	}
	return nil
}

// String returns a string representation.
func (m *BlockPartMessageWrapper) String() string {
	return fmt.Sprintf("[BlockPart H:%v R:%v P:%v]", m.Height, m.Round, m.Part)
}




type VoteMessage struct {
	Vote *Vote
}

// ValidateBasic performs basic validation.
func (m *VoteMessage) ValidateBasic() error {
	return m.Vote.ValidateBasic()
}

// String returns a string representation.
func (m *VoteMessage) String() string {
	return fmt.Sprintf("[Vote %v]", m.Vote)
}



// VoteSetBitsMessage is sent to communicate the bit-array of votes seen for the
// BlockID.
type VoteSetBitsMessage struct {
	Height  uint64
	Round   int32
	Type    types2.SignedMsgType
	BlockID BlockID
	Votes   *libs.BitArray
}

// ValidateBasic performs basic validation.
func (m *VoteSetBitsMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if !types3.IsVoteTypeValid(m.Type) {
		return errors.New("invalid Type")
	}
	if err := m.BlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockID: %v", err)
	}

	// NOTE: Votes.Size() can be zero if the node does not have any
	if m.Votes.Size() > types3.MaxVotesCount {
		return fmt.Errorf("votes bit array is too big: %d, max: %d", m.Votes.Size(), types3.MaxVotesCount)
	}

	return nil
}

// String returns a string representation.
func (m *VoteSetBitsMessage) String() string {
	return fmt.Sprintf("[VSB %v/%02d/%v %v %v]", m.Height, m.Round, m.Type, m.BlockID, m.Votes)
}
