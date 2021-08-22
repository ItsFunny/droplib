/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 3:52 下午
# @File : identity.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/p2p/utils"
	"github.com/libp2p/go-libp2p-core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
)

var (
	_ crypto.PrivKey = (*EcdsaPrvK)(nil)
	_ crypto.PubKey  = (*EcdsaPubK)(nil)
)

// FIXME  unexport
type EcdsaPrvK struct {
	PrivKey *ecdsa.PrivateKey
}

func (k *EcdsaPrvK) Bytes() ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(k.PrivKey)
}

func (e *EcdsaPrvK) Equals(key crypto.Key) bool {
	b, _ := e.Bytes()
	o, _ := key.Bytes()
	return bytes.Equal(b, o)
}

func (e *EcdsaPrvK) Raw() ([]byte, error) {
	return e.Bytes()
}

func (e *EcdsaPrvK) Type() crypto_pb.KeyType {
	return crypto_pb.KeyType_ECDSA
}

func (e *EcdsaPrvK) Sign(bytes []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, e.PrivKey, bytes)
	if err != nil {
		return nil, err
	}

	s, err = utils.ToLowS(&e.PrivKey.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return utils.MarshalECDSASignature(r, s)
}
func (e *EcdsaPrvK) GetPublic() crypto.PubKey {
	p := &EcdsaPubK{
		pubKey: &e.PrivKey.PublicKey,
	}
	return p
}

type EcdsaPubK struct {
	pubKey *ecdsa.PublicKey
}

func (e *EcdsaPubK) Bytes() ([]byte, error) {
	return x509.MarshalPKIXPublicKey(e.pubKey)
}

func (e *EcdsaPubK) Equals(key crypto.Key) bool {
	b, _ := e.Bytes()
	o, _ := key.Bytes()
	return bytes.Equal(b, o)
}

func (e *EcdsaPubK) Raw() ([]byte, error) {
	return e.Bytes()
}

func (e *EcdsaPubK) Type() crypto_pb.KeyType {
	return crypto_pb.KeyType_ECDSA
}

func (e *EcdsaPubK) Verify(data []byte, sig []byte) (bool, error) {
	r, s, err := utils.UnmarshalECDSASignature(sig)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling signature [%s]", err)
	}

	lowS, err := utils.IsLowS(e.pubKey, s)
	if err != nil {
		return false, err
	}

	if !lowS {
		return false, fmt.Errorf("Invalid S. Must be smaller than half the order [%s][%s].", s, utils.GetCurveHalfOrdersAt(e.pubKey.Curve))
	}

	return ecdsa.Verify(e.pubKey, data, r, s), nil
}
