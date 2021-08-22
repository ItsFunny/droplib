/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 10:36 下午
# @File : toproto.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/protos/libs/bits"
)

func BitArrayToProto(bA *libs.BitArray) *bits.BitArrayProto {
	if bA == nil ||
		(len(bA.Elems) == 0 && bA.Bits == 0) { // empty
		return nil
	}

	return &bits.BitArrayProto{Bits: int64(bA.Bits), Elems: bA.Elems}
}
