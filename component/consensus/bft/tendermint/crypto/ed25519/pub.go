/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/20 10:32 下午
# @File : pub.go
# @Description :
# @Attention :
*/
package ed25519

// import (
// 	"bytes"
// 	"crypto/ed25519"
// 	"fmt"
// 	"github.com/ChainSafe/go-schnorrkel"
// 	"github.com/hdevalence/ed25519consensus"
// 	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto"
// 	"github.com/hyperledger/fabric-droplib/libs"
// 	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
// )
//
// // PubKeySr25519 implements crypto.PubKey for the Sr25519 signature scheme.
// type PubKey []byte
//
// // Address is the SHA256-20 of the raw pubkey bytes.
// func (pubKey PubKey) Address() cryptolibs.Address {
// 	return cryptolibs.Address(libs.SumTruncated(pubKey[:]))
// }
//
// // Bytes returns the byte representation of the PubKey.
// func (pubKey PubKey) Bytes() []byte {
// 	return []byte(pubKey)
// }
//
// func (pubKey PubKey) VerifySignature(msg []byte, sig []byte) bool {
// 	// make sure we use the same algorithm to sign
// 	if len(sig) != SignatureSize {
// 		return false
// 	}
// 	var sig64 [SignatureSize]byte
// 	copy(sig64[:], sig)
//
// 	publicKey := &(schnorrkel.PublicKey{})
// 	var p [PubKeySize]byte
// 	copy(p[:], pubKey)
// 	err := publicKey.Decode(p)
// 	if err != nil {
// 		return false
// 	}
//
// 	signingContext := schnorrkel.NewSigningContext([]byte{}, msg)
//
// 	signature := &(schnorrkel.Signature{})
// 	err = signature.Decode(sig64)
// 	if err != nil {
// 		return false
// 	}
//
// 	return publicKey.Verify(signature, signingContext)
// }
//
// func (pubKey PubKey) String() string {
// 	return fmt.Sprintf("PubKeySr25519{%X}", []byte(pubKey))
// }
//
// // Equals - checks that two public keys are the same time
// // Runs in constant time based on length of the keys.
// func (pubKey PubKey) Equals(other PubKey) bool {
// 	if otherEd, ok := other.(PubKey); ok {
// 		return bytes.Equal(pubKey[:], otherEd[:])
// 	}
// 	return false
// }
//
// func (pubKey PubKey) Type() string {
// 	return KeyType
//
// }
//
// // BatchVerifier implements batch verification for ed25519.
// // https://github.com/hdevalence/ed25519consensus is used for batch verification
// type BatchVerifier struct {
// 	ed25519consensus.BatchVerifier
// }
//
// func NewBatchVerifier() crypto.BatchVerifier {
// 	return &BatchVerifier{ed25519consensus.NewBatchVerifier()}
// }
//
//
// func (b *BatchVerifier) Add(key cryptolibs.IPublicKey, msg, signature []byte) error {
// 	if l := len(key.ToBytes()); l != PubKeySize {
// 		return fmt.Errorf("pubkey size is incorrect; expected: %d, got %d", PubKeySize, l)
// 	}
//
// 	// check that the signature is the correct length & the last byte is set correctly
// 	if len(signature) != SignatureSize || signature[63]&224 != 0 {
// 		return errors.New("invalid signature")
// 	}
//
// 	b.BatchVerifier.Add(ed25519.PublicKey(key.ToBytes()), msg, signature)
//
// 	return nil
// }
//
// func (b *BatchVerifier) Verify() bool {
// 	return b.BatchVerifier.Verify()
// }

