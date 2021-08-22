/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:44 上午
# @File : tx.go
# @Description :
# @Attention :
*/
package types

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	"github.com/hyperledger/fabric-droplib/libs"
)

type Tx []byte

func (tx Tx) Hash() []byte {
	return libs.Sum(tx)
}

type Txs []Tx

func (txs Txs) Hash() []byte {
	// These allocations will be removed once Txs is switched to [][]byte,
	// ref #2603. This is because golang does not allow type casting slices without unsafe
	txBzs := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		txBzs[i] = txs[i].Hash()
	}
	return merkle.HashFromByteSlices(txBzs)
}
func (this Txs) Size() int {
	l := 0
	for i := 0; i < len(this); i++ {
		l += len(this[i])
	}
	return l
}
