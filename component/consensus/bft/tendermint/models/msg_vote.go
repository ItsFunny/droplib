/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 10:01 上午
# @File : msg_vote.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	"github.com/hyperledger/fabric-droplib/libs/protoio"
	types2 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"time"
)

type Vote struct {
	Type             types2.SignedMsgType `json:"type"`
	Height           uint64                    `json:"height"`
	Round            int32                    `json:"round"`    // assume there will not be greater than 2_147_483_647 rounds
	BlockID          BlockID                  `json:"block_id"` // zero if vote is nil.
	Timestamp        time.Time                `json:"timestamp"`
	ValidatorAddress cryptolibs.Address          `json:"validator_address"`
	ValidatorIndex   int32                    `json:"validator_index"`
	Signature        []byte                   `json:"signature"`
}
func NewCommitSigAbsent() CommitSig {
	return CommitSig{
		BlockIDFlag: BlockIDFlagAbsent,
	}
}
func (vote *Vote) CommitSig() CommitSig {
	if vote == nil {
		return NewCommitSigAbsent()
	}

	var blockIDFlag BlockIDFlag
	switch {
	case vote.BlockID.IsComplete():
		blockIDFlag = BlockIDFlagCommit
	case vote.BlockID.IsZero():
		blockIDFlag = BlockIDFlagNil
	default:
		panic(fmt.Sprintf("Invalid vote %v - expected BlockID to be either empty or complete", vote))
	}

	return CommitSig{
		BlockIDFlag:      blockIDFlag,
		ValidatorAddress: vote.ValidatorAddress,
		Timestamp:        vote.Timestamp,
		Signature:        vote.Signature,
	}
}
func (vote *Vote) ToBytes() []byte {
	panic("implement me")
}

func (vote *Vote) Verify(chainID string, pubKey cryptolibs.IPublicKey) error {
	if !bytes.Equal(pubKey.Address(), vote.ValidatorAddress) {
		return commonerrors.ErrVoteInvalidValidatorAddress
	}
	v := vote.ToProto()
	if !pubKey.VerifySignature(VoteSignBytes(chainID, v), vote.Signature) {
		return commonerrors.ErrVoteInvalidSignature
	}
	return nil
}
func VoteSignBytes(chainID string, vote *types2.VoteProto) []byte {
	pb := CanonicalizeVote(chainID, vote)
	bz, err := protoio.MarshalDelimited(&pb)
	if err != nil {
		panic(err)
	}

	return bz
}
func CanonicalizeVote(chainID string, vote *types2.VoteProto) types2.CanonicalVoteProto {
	return types2.CanonicalVoteProto{
		Type:      vote.Type,
		Height:    vote.Height,       // encoded as sfixed64
		Round:     int64(vote.Round), // encoded as sfixed64
		BlockId:   CanonicalizeBlockID(vote.BlockId),
		Timestamp: vote.Timestamp,
		ChainId:   chainID,
	}
}
func (vote *Vote) ToProto() *types2.VoteProto {
	if vote == nil {
		return nil
	}
	return &types2.VoteProto{
		Type:             vote.Type,
		Height:           int64(vote.Height),
		Round:            vote.Round,
		BlockId:          vote.BlockID.ToProto(),
		Timestamp:        &timestamp.Timestamp{
			Seconds: vote.Timestamp.UnixNano(),
			Nanos:   int32(vote.Timestamp.Nanosecond()),
		},
		ValidatorAddress: vote.ValidatorAddress,
		ValidatorIndex:   vote.ValidatorIndex,
		Signature:        vote.Signature,
	}
}
func (vote *Vote) ValidateBasic() error {
	if !types.IsVoteTypeValid(vote.Type) {
		return errors.New("invalid Type")
	}

	if vote.Height < 0 {
		return errors.New("negative Height")
	}

	if vote.Round < 0 {
		return errors.New("negative Round")
	}

	// NOTE: Timestamp validation is subtle and handled elsewhere.

	if err := vote.BlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockID: %v", err)
	}

	// BlockID.ValidateBasic would not err if we for instance have an empty hash but a
	// non-empty PartsSetHeader:
	if !vote.BlockID.IsZero() && !vote.BlockID.IsComplete() {
		return fmt.Errorf("blockID must be either empty or complete, got: %v", vote.BlockID)
	}

	if len(vote.ValidatorAddress) != cryptolibs.AddressSize {
		return fmt.Errorf("expected ValidatorAddress size to be %d bytes, got %d bytes",
			cryptolibs.AddressSize,
			len(vote.ValidatorAddress),
		)
	}
	if vote.ValidatorIndex < 0 {
		return errors.New("negative ValidatorIndex")
	}
	if len(vote.Signature) == 0 {
		return errors.New("signature is missing")
	}

	if len(vote.Signature) > types.MaxSignatureSize {
		return fmt.Errorf("signature is too big (max: %d)", types.MaxSignatureSize)
	}

	return nil
}

