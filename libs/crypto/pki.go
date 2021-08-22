/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:25 上午
# @File : pki.go
# @Description :
# @Attention :
*/
package cryptolibs

const (
	AddressSize = 20
)
type IToBytes interface {
	ToBytes() []byte
}
type ISigner interface {
	GetPublicKey() IPublicKey
	SignVote(chainID string, bytes IToBytes) ([]byte,error)
	SignProposal(chainID string, bytes IToBytes) ([]byte,error)
	Sign(msg []byte) ([]byte, error)
}

type IPublicKey interface {
	Address() Address
	VerifySignature(msg []byte, sig []byte) bool
	ToBytes() []byte
	Type() string
}



type DefaultEcdsaPublicKey struct {
}

func (d *DefaultEcdsaPublicKey) Address() Address {
	panic("implement me")
}

func (d *DefaultEcdsaPublicKey) VerifySignature(msg []byte, sig []byte) bool {
	panic("implement me")
}

func (d *DefaultEcdsaPublicKey) ToBytes() []byte {
	panic("implement me")
}


