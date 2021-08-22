/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 5:04 下午
# @File : mem_pool_test.go.go
# @Description :
# @Attention :
*/
package mempool

import (
	"fmt"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/event"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

func makeTxs(count int) types.Txs {
	r := make([]types.Tx, count)
	for i := 0; i < count; i++ {
		r[i] = makeTx(100)
	}
	return r
}
func makeTx(size int) types.Tx {
	return rand.Bytes(size)
}
func Test_AppendWithAcquire(t *testing.T) {
	bus := event.NewDefaultEventBus()
	cfg := config.NewDefaultMemPoolConfiguration()
	pool := NewDefaultMemPool(1, cfg, bus)
	all := 0
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 1; i <= 10; i++ {
		all += i
		go func(count int) {
			defer wg.Done()
			txs := makeTxs(count)
			info := models.TxInfo{
				SenderID:    uint16(count),
				SenderP2PID: types2.NodeID(rand.CRandHex(3)),
			}
			err := pool.CheckAndAppendTx(txs, info)
			if nil != err {
				fmt.Println(err.Error())
				os.Exit(-1)
			}
		}(i)
	}
	wg.Wait()

	res := 0
	go func() {
		for {
			txs := pool.acquireTillEnd()
			res += len(txs)
		}
	}()
	time.Sleep(time.Second * 3)
	require.Equal(t, all, res)
}
