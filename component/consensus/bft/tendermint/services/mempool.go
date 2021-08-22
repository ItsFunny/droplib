/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:14 上午
# @File : mempool.go
# @Description :
# @Attention :
*/
package services

import (
	"crypto/sha256"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
)

const TxKeySize = sha256.Size

// --------------------------------------------------------------------------------

// 预留结构体,用于实现跨链
type ResponseCheckTx struct {
}

// PreCheckFunc is an optional filter executed before CheckTx and rejects
// transaction if false is returned. An example would be to ensure that a
// transaction doesn't exceeded the block size.
type PreCheckFunc func(types.Tx) error
type PostCheckFunc func(types.Tx, *ResponseCheckTx) error

type IMemPool interface {
	services2.IBaseService
	CheckAndAppendTx(tx types.Txs, txInfo models.TxInfo) error

	// AcquireTxWithLimit(max int) types.Txs
	// 应该是每次只获取一个,无论多大
	AcquireOnePatch() types.Txs
	Size() int
	Update(blockHeight int64, blockTxs types.Txs, newPreFn PreCheckFunc, newPostFn PostCheckFunc, ) error

	TxsAvailable() <-chan struct{}
	// EnableTxsAvailable initializes the TxsAvailable channel, ensuring it will
	// trigger once every height when transactions are available.
	EnableTxsAvailable()

	// Flush removes all transactions from the mempool and cache
	Flush()

	// TxsBytes returns the total size of all txs in the mempool.
	TxsBytes() int64

	// InitWAL creates a directory for the WAL file and opens a file itself. If
	// there is an error, it will be of type *PathError.
	InitWAL() error

	// CloseWAL closes and discards the underlying WAL file.
	// Any further writes will not be relayed to disk.
	CloseWAL()

	EnsureNoFire()

	// Lock locks the mempool. The consensus must be able to hold lock to safely update.
	Lock()

	// Unlock unlocks the mempool.
	Unlock()
}

type IMemPoolReactor interface {
	services.IReactor
}


type IMemPoolLogicService interface {
	base.ILogicService
	MemPool()
}