/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 2:00 下午
# @File : validator.go
# @Description :
# @Attention :
*/
package mock

import (
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	types2 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

type MockValidator struct {
	// KILL 充当假节点功能
	Index       int32 // Validator index. NOTE: we don't assume validator set changes.
	Height      int64
	Round       int32
	signer      cryptolibs.ISigner
	VotingPower int64
}

func (f *MockValidator) Sign(msg []byte) ([]byte, error) {
	panic("implement me")
}

func (f *MockValidator) SignVote(chainID string, bytes cryptolibs.IToBytes) ([]byte, error) {
	return f.signer.SignVote(chainID, bytes)
}

func (f *MockValidator) SignProposal(chainID string, bytes cryptolibs.IToBytes) ([]byte, error) {
	return f.signer.SignProposal(chainID, bytes)
}

func (f *MockValidator) GetPublicKey() cryptolibs.IPublicKey {
	return f.signer.GetPublicKey()
}

var testMinPower int64 = 10

func NewMockValidator(siger cryptolibs.ISigner, valIndex int32) *MockValidator {
	return &MockValidator{
		Index:       valIndex,
		signer:      siger,
		VotingPower: testMinPower,
	}
}

func (vs *MockValidator) MockSignVote(globalCfg *config.GlobalConfiguration, voteType types2.SignedMsgType, hash []byte, header models.PartSetHeader) (*models.Vote, error) {
	pubKey := vs.signer.GetPublicKey()

	vote := &models.Vote{
		ValidatorIndex:   vs.Index,
		ValidatorAddress: pubKey.Address(),
		Height:           uint64(vs.Height),
		Round:            vs.Round,
		Timestamp:        libs.Now(),
		Type:             voteType,
		BlockID:          models.BlockID{Hash: hash, PartSetHeader: header},
	}
	// v := vote.ToProto()
	sig, err := vs.signer.SignVote(globalCfg.ChainID(), vote)
	vote.Signature = sig

	return vote, err
}
