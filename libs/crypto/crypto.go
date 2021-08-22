/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:05 上午
# @File : crypto.go
# @Description :
# @Attention :
*/
package cryptolibs

import (
	"bytes"
	"encoding/hex"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
)

func ConvtByes2PublicKey(bytes []byte) (IPublicKey, error) {
	return nil, nil
}

type Address []byte
type MockPublicKey struct {
	addr Address
}

func NewMockPublicKey(pubk string) *MockPublicKey {
	decodeString, _ := hex.DecodeString(pubk)
	r := &MockPublicKey{
		addr: decodeString,
	}
	return r
}

func (m *MockPublicKey) Type() string {
	panic("implement me")
}

func (m *MockPublicKey) Address() Address {
	if m.addr == nil {
		m.addr = []byte(rand.Str(20))
	}
	return m.addr
}

func (m *MockPublicKey) VerifySignature(msg []byte, sig []byte) bool {
	return string(sig) == "123"
}

func (m *MockPublicKey) ToBytes() []byte {
	return m.Address()
}

type MockSigner struct {
	pubK IPublicKey
}

func (m *MockSigner) Sign(msg []byte) ([]byte, error) {
	return []byte("123"), nil
}

func (m *MockSigner) GetPublicKey() IPublicKey {
	if m.pubK == nil {
		m.pubK = &MockPublicKey{}
	}
	return m.pubK
}

func (m *MockSigner) SignVote(chainID string, bytes IToBytes) ([]byte, error) {
	return []byte("123"), nil
}

func (m *MockSigner) SignProposal(chainID string, bytes IToBytes) ([]byte, error) {
	return []byte("123"), nil
}

type PrivValidatorsByAddress []ISigner

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	pvi := pvs[i].GetPublicKey()
	pvj := pvs[j].GetPublicKey()
	return bytes.Compare(pvi.Address(), pvj.Address()) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	pvs[i], pvs[j] = pvs[j], pvs[i]
}
