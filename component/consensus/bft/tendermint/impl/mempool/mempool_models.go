/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 2:36 下午
# @File : mempool_models.go
# @Description :
# @Attention :
*/
package mempool

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"sync"
)

type mempoolTxWrapper struct {
	height int64 // height that this tx had been validated in
	// gasWanted int64    // amount of gas this tx states it will require
	tx types.Tx //
	// ids of peers who've sent us this tx (as a map for quick lookups).
	// senders: PeerID -> bool
	senders sync.Map
}


// 预留结构体,用于实现跨链
type ResponseCheckTx struct {
}
