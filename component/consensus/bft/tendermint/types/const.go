/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 1:27 下午
# @File : const.go
# @Description :
# @Attention :
*/
package types

import (
	"crypto/ed25519"
	"github.com/hyperledger/fabric-droplib/libs"
)

var (
	// MaxSignatureSize is a maximum allowed signature size for the Proposal
	// and Vote.
	// XXX: secp256k1 does not have Size nor MaxSize defined.
	MaxSignatureSize = libs.MaxInt(ed25519.SignatureSize, 64)
)
const (
	// MaxVotesCount is the maximum number of votes in a set. Used in ValidateBasic funcs for
	// protection against DOS attacks. Note this implies a corresponding equal limit to
	// the number of validators.
	MaxVotesCount = 10000
)





const (
	MaxMsgSize = 1048576 // 1MB; NOTE: keep in sync with types.PartSet sizes.
)
