/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 1:08 下午
# @File : crypto.go
# @Description :
# @Attention :
*/
package crypto

import cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"

// If a new key type implements batch verification,
// the key type must be registered in github.com/tendermint/tendermint/crypto/batch
type BatchVerifier interface {
	// Add appends an entry into the BatchVerifier.
	Add(key cryptolibs.IPublicKey, message, signature []byte) error
	// Verify verifies all the entries in the BatchVerifier.
	// If the verification fails it is unknown which entry failed and each entry will need to be verified individually.
	Verify() bool
}
