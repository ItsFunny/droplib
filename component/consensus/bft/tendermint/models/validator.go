/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:02 上午
# @File : validator.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"fmt"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

type ValidatorInfo struct {
	Address cryptolibs.Address `json:"address"`
	Index   int32              `json:"index"`
}

// Volatile state for each Validator
// NOTE: The ProposerPriority is not included in Validator.Hash();
// make sure to update that method if changes are made here
type Validator struct {
	Address     cryptolibs.Address    `json:"address"`
	PubKey      cryptolibs.IPublicKey `json:"pub_key"`
	VotingPower int64                 `json:"voting_power"`

	ProposerPriority int64 `json:"proposer_priority"`
}

func (vals *ValidatorSet) CopyIncrementProposerPriority(times int32) *ValidatorSet {
	copy := vals.Copy()
	copy.IncrementProposerPriority(times)
	return copy
}
func (v *Validator) String() string {
	if v == nil {
		return "nil-Validator"
	}
	return fmt.Sprintf("Validator{%v %v VP:%v A:%v}",
		v.Address,
		v.PubKey,
		v.VotingPower,
		v.ProposerPriority)
}
func (v *Validator) ValidateBasic() error {
	if v == nil {
		return errors.New("nil validator")
	}
	if v.PubKey == nil {
		return errors.New("validator does not have a public key")
	}

	if v.VotingPower < 0 {
		return errors.New("validator has negative voting power")
	}

	if len(v.Address) != cryptolibs.AddressSize {
		return fmt.Errorf("validator address is the wrong size: %v", v.Address)
	}

	return nil
}
func (v *Validator) ToProto() (*types3.ValidatorProto, error) {
	if v == nil {
		return nil, errors.New("nil validator")
	}

	vp := types3.ValidatorProto{
		Address:          v.Address,
		PubKey:           v.PubKey.ToBytes(),
		VotingPower:      v.VotingPower,
		ProposerPriority: v.ProposerPriority,
	}

	return &vp, nil
}

func (v *Validator) Bytes() []byte {
	// pk, err := ce.PubKeyToProto(v.PubKey)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// pbv := tmproto.SimpleValidator{
	// 	PubKey:      &pk,
	// 	VotingPower: v.VotingPower,
	// }
	//
	// bz, err := pbv.Marshal()
	// if err != nil {
	// 	panic(err)
	// }
	// return bz
	return nil
}
func (v *Validator) Copy() *Validator {
	vCopy := *v
	return &vCopy
}
func NewValidator(pubKey cryptolibs.IPublicKey, votingPower int64) *Validator {
	return &Validator{
		Address:          pubKey.Address(),
		PubKey:           pubKey,
		VotingPower:      votingPower,
		ProposerPriority: 0,
	}
}

func RandValidator(randPower bool, minPower int64) (*Validator, cryptolibs.ISigner) {
	privVal := NewMockPV()
	votePower := minPower
	if randPower {
		votePower += int64(rand.Uint32())
	}
	pubKey := privVal.GetPublicKey()
	val := NewValidator(pubKey, votePower)
	return val, &privVal
}

// MockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type MockPV struct {
	PrivKey              cryptolibs.ISigner
	breakProposalSigning bool
	breakVoteSigning     bool
}

func (pv *MockPV) SignVote(chainID string, bytes cryptolibs.IToBytes) ([]byte, error) {
	return []byte("123"), nil
}

func (pv *MockPV) SignProposal(chainID string, bytes cryptolibs.IToBytes) ([]byte, error) {
	return []byte("123"), nil
}

func (pv *MockPV) Sign(msg []byte) ([]byte, error) {
	return []byte("123"), nil
}

func (pv *MockPV) GetPublicKey() cryptolibs.IPublicKey {
	return pv.PrivKey.GetPublicKey()
}

func NewMockPV() MockPV {
	// return MockPV{&bccsp.MockSigner{}, false, false}
	v := &cryptolibs.MockSigner{}
	return MockPV{
		PrivKey:              v,
		breakProposalSigning: false,
		breakVoteSigning:     false,
	}
}
