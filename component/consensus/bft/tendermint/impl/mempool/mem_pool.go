/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 2:35 下午
# @File : mem_pool.go
# @Description :
# @Attention :
*/
package mempool

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	services2 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
	"github.com/hyperledger/fabric-droplib/libs"
	auto "github.com/hyperledger/fabric-droplib/libs/autofiles"
	"github.com/hyperledger/fabric-droplib/libs/clist"
	"github.com/hyperledger/fabric-droplib/libs/queue/goconcurrentqueue"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_          services.IMemPool = (*defaultMemPool)(nil)
	subscriber                   = "MemPoolService"
)

var newline = []byte("\n")

type defaultMemPool struct {
	*impl.BaseServiceImpl

	// Atomic integers
	height   int64 // the last block Update()'d to
	txsBytes int64 // total size of mempool, in bytes

	updateMtx sync.RWMutex
	preCheck  services.PreCheckFunc
	postCheck services.PostCheckFunc

	// Track whether we're rechecking txs.
	// These are not protected by a mutex and are expected to be mutated in
	// serial (ie. by abci responses which are called in serial).
	recheckCursor *clist.CElement // next expected response
	recheckEnd    *clist.CElement // re-checking stops here

	cfg *config.MemPoolConfiguration

	txCache      txCache
	txsMap       sync.Map
	txs          *clist.CList // concurrent linked-list of good txs
	acquireQueue goconcurrentqueue.Queue

	wal *auto.AutoFile

	// 交互
	notifiedTxsAvailable bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty

	eventBus services2.ICommonEventBusComponent
}

func NewDefaultMemPool(height int64, cfg *config.MemPoolConfiguration, eventBus services2.ICommonEventBusComponent) *defaultMemPool {
	d := &defaultMemPool{
		height:       height,
		cfg:          cfg,
		txCache:      newMapTxCache(cfg.Size),
		txsMap:       sync.Map{},
		txs:          clist.New(),
		txsAvailable: nil,
		acquireQueue: goconcurrentqueue.NewFIFO(),
	}
	if cfg.CacheSize > 0 {
		d.txCache = newMapTxCache(cfg.CacheSize)
	} else {
		d.txCache = nopTxCache{}
	}
	d.eventBus = eventBus
	d.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_MEM_POOL, d)

	return d
}

func (mem *defaultMemPool) Update(height int64, blockTxs types.Txs, preCheck services.PreCheckFunc, postCheck services.PostCheckFunc) error {
	mem.Logger.Info("开始更新交易信息")
	// Set height
	mem.height = height
	mem.notifiedTxsAvailable = false
	if preCheck != nil {
		mem.preCheck = preCheck
	}
	if postCheck != nil {
		mem.postCheck = postCheck
	}

	// 认为一切都是ok的,到这一步的时候
	for _, tx := range blockTxs {
		_ = mem.txCache.Push(tx)
		if e, ok := mem.txsMap.Load(TxKey(tx)); ok {
			mem.removeTx(tx, e.(*clist.CElement), false)
		}
	}

	// Either recheck non-committed txs to see if they became invalid
	// or just notify there're some txs left.
	if mem.Size() > 0 {
		if mem.cfg.Recheck {
			mem.Logger.Info("Recheck txs", "numtxs", mem.Size(), "height", height)
			mem.recheckTxs()
			// At this point, mem.txs are being rechecked.
			// mem.recheckCursor re-scans mem.txs and possibly removes some txs.
			// Before mem.Reap(), we should wait for mem.recheckCursor to be nil.
		} else {
			mem.notifyTxsAvailable()
		}
	}

	return nil
}

func (mem *defaultMemPool) OnReset() error {
	return nil
}

func (mem *defaultMemPool) OnStop() {
	mem.eventBus.UnsubscribeAll(context.Background(), subscriber)
}

func (mem *defaultMemPool) EnableTxsAvailable() {
	mem.txsAvailable = make(chan struct{}, 1)
}
func (mem *defaultMemPool) InitWAL() error {
	var (
		walDir  = mem.cfg.WalDir()
		walFile = walDir + "/wal"
	)

	const perm = 0700
	if err := libs.EnsureDir(walDir, perm); err != nil {
		return err
	}

	af, err := auto.OpenAutoFile(walFile)
	if err != nil {
		return fmt.Errorf("can't open autofile %s: %w", walFile, err)
	}

	mem.wal = af
	return nil
}

func (mem *defaultMemPool) CloseWAL() {
	if err := mem.wal.Close(); err != nil {
		mem.Logger.Error("Error closing WAL", "err", err)
	}
	mem.wal = nil
}

// Safe for concurrent use by multiple goroutines.
func (mem *defaultMemPool) Lock() {
	mem.updateMtx.Lock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *defaultMemPool) Unlock() {
	mem.updateMtx.Unlock()
}

func (mem *defaultMemPool) Flush() {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	_ = atomic.SwapInt64(&mem.txsBytes, 0)
	mem.txCache.Reset()

	for e := mem.txs.Front(); e != nil; e = e.Next() {
		mem.txs.Remove(e)
		e.DetachPrev()
	}

	mem.txsMap.Range(func(key, _ interface{}) bool {
		mem.txsMap.Delete(key)
		return true
	})
}

func (mem *defaultMemPool) TxsAvailable() <-chan struct{} {
	return mem.txsAvailable
}

func (mem *defaultMemPool) recheckTxs() {
	if mem.Size() == 0 {
		panic("recheckTxs is called, but the mempool is empty")
	}

	mem.recheckCursor = mem.txs.Front()
	mem.recheckEnd = mem.txs.Back()

	// Push txs to proxyAppConn
	// NOTE: globalCb may be called concurrently.
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTxWrapper)
		mem.resCbRecheck(memTx.tx)
		if e.Next() != nil && e.Next().Next() == nil {
			fmt.Println(1)
		}
	}
}

func (mem *defaultMemPool) resCbRecheck(txOrigin types.Tx) {
	memTx := mem.recheckCursor.Value.(*mempoolTxWrapper)
	// txOrigin := req.GetCheckTx().Tx
	// memTx := mem.recheckCursor.Value.(*mempoolTxWrapper)
	if !bytes.Equal(txOrigin, memTx.tx) {
		panic(fmt.Sprintf(
			"Unexpected tx response from proxy during recheck\nExpected %X, got %X",
			memTx.tx,
			txOrigin))
	}
	// var postCheckErr error
	// if mem.postCheck != nil {
	// 	postCheckErr = mem.postCheck(txOrigin, nil)
	// }
	if mem.recheckCursor == mem.recheckEnd {
		mem.recheckCursor = nil
	} else {
		mem.recheckCursor = mem.recheckCursor.Next()
	}
	if mem.recheckCursor == nil {
		// Done!
		mem.Logger.Info("Done rechecking txs")
		// incase the recheck removed all txs
		if mem.Size() > 0 {
			mem.notifyTxsAvailable()
		}
	}
}

// Called from:
//  - Update (lock held) if tx was committed
// 	- resCbRecheck (lock not held) if tx was invalidated
func (mem *defaultMemPool) removeTx(tx types.Tx, elem *clist.CElement, removeFromCache bool) {
	mem.txs.Remove(elem)
	elem.DetachPrev()
	mem.txsMap.Delete(TxKey(tx))
	atomic.AddInt64(&mem.txsBytes, int64(-len(tx)))

	if removeFromCache {
		mem.txCache.Remove(tx)
	}
}
func (d *defaultMemPool) Size() int {
	return d.txs.Len()
}

func (d *defaultMemPool) CheckAndAppendTx(txs types.Txs, txInfo models.TxInfo) error {
	d.updateMtx.Lock()
	defer d.updateMtx.Unlock()

	if len(txs) == 0 {
		return nil
	}
	if d.wal != nil {
		for _, tx := range txs {
			// TODO: Notify administrators when WAL fails
			// 缓存交易信息
			_, err := d.wal.Write(append([]byte(tx), newline...))
			if err != nil {
				return fmt.Errorf("wal.Write: %w", err)
			}
		}
	}
	size := 0
	for _, tx := range txs {
		if !d.txValid(tx) || !d.appendTx(tx) {
			// Record a new sender for a tx we've already seen.
			// Note it's possible a tx is still in the cache but no longer in the mempool
			// (eg. after committing a block, txs are removed from mempool but not cache),
			// so we only record the sender for txs still in the mempool.
			if e, ok := d.txsMap.Load(TxKey(tx)); ok {
				memTx := e.(*clist.CElement).Value.(*mempoolTxWrapper)
				memTx.senders.LoadOrStore(txInfo.SenderID, true)
				// TODO: consider punishing peer for dups,
				// its non-trivial since invalid txs can become valid,
				// but they can spam the same tx with little cost to them atm.
			}
			return commonerrors.ErrTxInCache
		}
		if err := d.addBlockTx(tx, txInfo); nil != err {
			return err
		}
		size++
	}
	d.Logger.Info("添加交易,交易数量为:" + strconv.Itoa(size))
	if err := d.acquireQueue.Enqueue(size); nil != err {
		panic("should not happen:" + err.Error())
	}

	d.notifyTxsAvailable()

	return nil
}
func (mem *defaultMemPool) addBlockTx(tx types.Tx, txInfo models.TxInfo) error {
	if err := mem.isFull(len(tx)); err != nil {
		// remove from cache (mempool might have a space later)
		mem.txCache.Remove(tx)
		mem.Logger.Error(err.Error())
		return err
	}
	memTx := &mempoolTxWrapper{
		height: mem.height,
		// gasWanted: r.CheckTx.GasWanted,
		tx: tx,
	}
	memTx.senders.Store(txInfo.SenderID, true)
	mem.addTx(memTx)
	mem.Logger.Info("添加交易,", "tx:", txID(tx), "高度:", memTx.height, "total:", mem.Size())

	return nil
}
func (mem *defaultMemPool) notifyTxsAvailable() {
	if mem.Size() == 0 {
		panic("notified txs available but mempool is empty!")
	}
	if mem.txsAvailable != nil && !mem.notifiedTxsAvailable {
		// channel cap is 1, so this will send once
		mem.notifiedTxsAvailable = true
		select {
		case mem.txsAvailable <- struct{}{}:
		default:
		}
	}
}
func (mem *defaultMemPool) addTx(memTx *mempoolTxWrapper) {
	e := mem.txs.PushBack(memTx)
	mem.txsMap.Store(TxKey(memTx.tx), e)
	atomic.AddInt64(&mem.txsBytes, int64(len(memTx.tx)))
	// mem.metrics.TxSizeBytes.Observe(float64(len(memTx.tx)))
}

func (d *defaultMemPool) appendTx(tx types.Tx) bool {
	return d.txCache.Push(tx)
}
func (mem *defaultMemPool) AcquireOnePatch() types.Txs {
	l, _ := mem.acquireQueue.Dequeue()
	if l == nil {
		return nil
	}
	size := l.(int)

	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	min := libs.MinInt(mem.txs.Len(), size)
	txs := make([]types.Tx, 0, min)
	index := 0
	for e := mem.txs.Front(); e != nil && index < min; e = e.Next() {
		memTx := e.Value.(*mempoolTxWrapper)
		txs = append(txs, memTx.tx)
		index++
	}

	return txs
}

// func (mem *defaultMemPool) AcquireTxWithLimit(max int) types.Txs {
// 	mem.updateMtx.RLock()
// 	defer mem.updateMtx.RUnlock()
//
// 	if max < 0 {
// 		max = mem.txs.Len()
// 	}
//
// 	txs := make([]types.Tx, 0, libs.MinInt(mem.txs.Len(), max))
//
// 	for e := mem.txs.Front(); e != nil && len(txs) <= max; e = e.Next() {
// 		memTx := e.Value.(*mempoolTxWrapper)
// 		txs = append(txs, memTx.tx)
// 	}
//
// 	return txs
// }

func (d *defaultMemPool) txValid(tx types.Tx) bool {
	// 默认都认为是有效的
	return true
}

func (mem *defaultMemPool) isFull(txSize int) error {
	var (
		memSize  = mem.Size()
		txsBytes = mem.TxsBytes()
	)

	if memSize >= mem.cfg.Size || int64(txSize)+txsBytes > mem.cfg.MaxTxsBytes {
		return ErrMempoolIsFull{
			memSize, mem.cfg.Size,
			txsBytes, mem.cfg.MaxTxsBytes,
		}
	}

	return nil
}
func (mem *defaultMemPool) TxsBytes() int64 {
	return atomic.LoadInt64(&mem.txsBytes)
}

func (mem *defaultMemPool) EnsureNoFire() {
	timer := time.NewTimer(time.Duration(500) * time.Millisecond)
	select {
	case <-mem.txsAvailable:
		log.Fatalln("Expected not to fire")
	case <-timer.C:
	}
}

// FOR TEST 
func (mem *defaultMemPool) ensureFire(ch <-chan struct{}, timeoutMS int) {
	timer := time.NewTimer(time.Duration(timeoutMS) * time.Millisecond)
	select {
	case <-ch:
	case <-timer.C:
		log.Fatalln("Expected to fire")
	}
}

func (mem *defaultMemPool) acquireTillEnd() types.Txs {
	l, _ := mem.acquireQueue.DequeueOrWaitForNextElement()
	size := l.(int)

	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	min := libs.MinInt(mem.txs.Len(), size)
	txs := make([]types.Tx, 0, min)
	index := 0
	for e := mem.txs.Front(); e != nil && index < min; e = e.Next() {
		memTx := e.Value.(*mempoolTxWrapper)
		txs = append(txs, memTx.tx)
		index++
	}
	return txs
}
