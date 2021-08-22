/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/10 8:50 上午
# @File : base.go
# @Description :
# @Attention :
*/
package watcher

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/listener"
	services2 "github.com/hyperledger/fabric-droplib/component/routine/services"
	v2 "github.com/hyperledger/fabric-droplib/component/routine/services/v2"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"github.com/hyperledger/fabric-droplib/structure/lists/singlylinkedlist"
	"github.com/sasha-s/go-deadlock"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	logcomponent.RegisterBlackList("log/g_log")
}

type ChannelWatcher interface {
	services.IBaseService
	Started() bool

	WatchMemberChanged(c ChannelMember) bool
	OnMemberChanged(c2 ChannelMember)

	AutoChange(ops ...Option) ChannelWatcher

	RollBack(ops ...Option) ChannelWatcher
	OnRollBack(opt Opt) (ChannelWatcher, []models.StartOption)

	Upgrade(ops ...Option) ChannelWatcher
	OnUpgrade(opt Opt) (ChannelWatcher, []models.StartOption)

	GetChannelShims(cap int) (map[channel.ChannelID]*ChannelWp, int)
	Size() int
}
type ChannelMember struct {
	name     string
	c        <-chan channel.IData
	consumer DataConsumer
}

func (this ChannelMember) ID() interface{} {
	return this.name
}

func (this ChannelMember) String() string {
	return this.name
}

type baseStatus struct {
	status                        uint32
	waitStatusIntervalMillSeconds int
}
type ChannelWp struct {
	ch              *channel.Channel
	inFlightRoutine int32
	flush           func(v channel.IData)
}

type baseChannelWatcher struct {
	*impl.BaseServiceImpl

	mtx deadlock.RWMutex

	impl ChannelWatcher

	facadedC    chan channel.Envelope
	inFlight    int32
	internalChs map[channel.ChannelID]*ChannelWp
	deleta      *singlylinkedlist.List

	listener    listener.IListenerComponent
	routinePool services2.IRoutineComponent

	operation chan operation

	operationCache map[interface{}]channel.IData
	debugC         chan struct{}

	wg sync.WaitGroup
	// in case of  closing the  logic channel and thus system panic
	internalFastQuitC chan struct{}

	// pro
	mode          byte
	spinTimeMills time.Duration
	rollBackLimit int32
	upgradeLimit  int32
	*baseStatus
}

// rollBack 0 / <0 means: never roll back
// upgradeLimit 0  / <0 means: never upgrade
func newBaseChannelWatcher(name string, cImpl ChannelWatcher, upgradeLimit, rollbackLimit int32, spinTimeMills int32, f func() services2.IRoutineComponent, mode byte) *baseChannelWatcher {
	if f == nil || f() == nil {
		f = func() services2.IRoutineComponent {
			return v2.NewV2RoutinePoolExecutorComponent(v2.WithSize(default_routine_pool_size))
		}
	}
	r := &baseChannelWatcher{
		impl:              cImpl,
		facadedC:          make(chan channel.Envelope, 10),
		internalChs:       nil,
		internalFastQuitC: make(chan struct{}),
		rollBackLimit:     rollbackLimit,
		upgradeLimit:      upgradeLimit,
		baseStatus:        nil,
		deleta:            singlylinkedlist.New(),
		operation:         make(chan operation, 20),
		listener:          listener.DefaultNewListenerComponent(),
		spinTimeMills:     time.Duration(spinTimeMills),
		mode:              mode,
	}
	r.routinePool = f()
	if upgradeLimit <= 0 {
		upgradeLimit = math.MaxInt32
	}
	if rollbackLimit < 0 {
		rollbackLimit = 0
	}
	r.rollBackLimit = rollbackLimit
	r.upgradeLimit = upgradeLimit
	r.BaseServiceImpl = impl.NewBaseService(nil, modules.NewModule(name, 1), r)
	r.baseStatus = &baseStatus{
		status: status_ok,
	}
	r.operationCache = make(map[interface{}]channel.IData)
	r.debugC = make(chan struct{})
	return r
}

type upgrade struct {
	opt Opt
	id  int
}

func newUpgrade(opt Opt) upgrade {
	return upgrade{
		opt: opt,
		id:  acquireId(),
	}
}
func (u upgrade) ID() interface{} {
	return u.id
}

type rollback struct {
	opt Opt
}

func (r rollback) ID() interface{} {
	return "rollback"
}

func (this *baseChannelWatcher) inc(d channel.IData) {
	size := atomic.AddInt32(&this.inFlight, 1)
	if this.mode == mode_debug {
		this.mtx.Lock()
		this.operationCache[d.ID()] = d
		this.mtx.Unlock()
		this.Logger.Debug("sendMsg", "size", size, "data", d, "dataType", reflect.TypeOf(d))
	}
}
func (this *baseChannelWatcher) dec(data channel.IData) {
	size := atomic.AddInt32(&this.inFlight, -1)
	if this.mode == mode_debug {
		this.mtx.Lock()
		delete(this.operationCache, data.ID())
		this.mtx.Unlock()
		this.Logger.Debug("consumeMsg", "size", size, "data", data, "dataType", reflect.TypeOf(data))
	}
}
func (this *baseChannelWatcher) sendOperation(opType byte, data IOperationData) {
	op := acquireOperation(opType, data)
	this.operation <- op
}
func (this *baseChannelWatcher) addDelta(c ChannelMember) {
	size := this.deleta.AddReturnSize(c)
	this.Logger.Debug("deltaAddMember", "name", c.name, "size", size)
}
func (b *baseChannelWatcher) AutoChange(ops ...Option) ChannelWatcher {
	return b.action(func(opt Opt) (ChannelWatcher, []models.StartOption) {
		var r ChannelWatcher
		var opts []models.StartOption
		if b.impl.Size() < int(b.rollBackLimit) {
			r, opts = b.OnRollBack(opt)
		} else {
			r, opts = b.impl.OnUpgrade(opt)
		}
		return r, opts
	}, ops...)
}
func (b *baseChannelWatcher) Upgrade(ops ...Option) ChannelWatcher {
	return b.action(func(opt Opt) (ChannelWatcher, []models.StartOption) {
		return b.OnUpgrade(opt)
	}, ops...)
}
func (this *baseChannelWatcher) RollBack(ops ...Option) ChannelWatcher {
	return this.action(func(opt Opt) (ChannelWatcher, []models.StartOption) {
		return this.OnRollBack(opt)
	}, ops...)
}
func (this *baseChannelWatcher) action(f func(opt Opt) (ChannelWatcher, []models.StartOption), ops ...Option) ChannelWatcher {
	if !this.Started() {
		this.Logger.Debug("该服务还未启动", "impl", this.impl)
		return nil
	}
	if !atomic.CompareAndSwapUint32(&this.status, status_ok, status_changing) {
		this.Logger.Infof("该watcher处于更新状态中,本次操作取消")
		return nil
	}
	close(this.internalFastQuitC)
	this.wg.Wait()
	notify := this.handleDeltaMember()

	opt := GetOption(ops...)
	var r ChannelWatcher
	var opts []models.StartOption
	r, opts = f(opt)
	if nil != r {
		this.BStop(models.StopCTXWithKV("notify", notify))
		r.BStart(opts...)
		delIt := this.deleta.Iterator()
		for delIt.Next() {
			m := delIt.Value().(ChannelMember)
			r.OnMemberChanged(m)
		}
	} else {
		// FIXME keep going
		panic(PROGRAMA_ERROR)
	}
	return r
}

func (this *baseChannelWatcher) OnStop(ctx *models.StopCTX) {
	size := atomic.LoadInt32(&this.inFlight)
	// spin is better
	for ; size != 0; {
		this.Logger.Info("not done yet", "size", size)
		time.Sleep(time.Millisecond * this.spinTimeMills)
		size = atomic.LoadInt32(&this.inFlight)
	}
	for _, v := range this.internalChs {
		v.ch.Close()
	}
	this.wg.Wait()
	this.impl.OnStop(ctx)
	this.listener.BStop(models.StopCTXAsChild(ctx))
	this.routinePool.BStop(models.StopCTXAsChild(ctx))
	close(this.operation)
	close(this.facadedC)
	close(this.debugC)
}
func (this *baseChannelWatcher) OnRollBack(opt Opt) (ChannelWatcher, []models.StartOption) {
	begin := time.Now()
	this.Logger.Info("开始回滚")
	r, opts := this.impl.OnRollBack(opt)
	cost := time.Now().Sub(begin)
	this.Logger.Info("回滚结束", fmt.Sprintf("耗时:%f秒,%d毫秒", cost.Seconds(), cost.Milliseconds()))
	return r, opts
}
func (this *baseStatus) Status() uint32 {
	return atomic.LoadUint32(&this.status)
}
func (this *baseStatus) And(status, check uint32) bool {
	return atomic.LoadUint32(&this.status)&status >= check
}
func (this *baseStatus) CasStatus(exce int, newV int) bool {
	return atomic.CompareAndSwapUint32(&this.status, uint32(exce), uint32(newV))
}
func (this *baseStatus) panicCAS(exce, newV uint32, f func()) {
	if !atomic.CompareAndSwapUint32(&this.status, exce, newV) {
		panic(fmt.Sprintf("%d,%d,%d", this.status, exce, newV))
	} else {
		if nil != f {
			f()
		}
	}
}
func (b *baseChannelWatcher) OnMemberChanged(c2 ChannelMember) {
	b.sendMsg(memberNotifyC, c2)
}

func (b *baseChannelWatcher) WatchMemberChanged(c ChannelMember) bool {
	if atomic.LoadUint32(&b.status)&status_deny_memchanged >= status_deny_memchanged {
		b.Logger.Info("该watch 处于 memchanged 中", "impl", b.impl)
		return true
	}
	select {
	case <-b.Quit():
		panic("该routine已经退出")
	default:
	}
	size := b.impl.Size()
	if size > int(b.upgradeLimit) || (size > 0 && size < int(b.rollBackLimit)) {
		if size < int(b.rollBackLimit) {
			b.Logger.Info("该watcher需要回滚", "impl", b.impl, "size", size, "upgradeLimit", b.upgradeLimit, "rollbackLimit", b.rollBackLimit)
		} else {
			b.Logger.Info("该watcher需要升级", "impl", b.impl, "size", size, "upgradeLimit", b.upgradeLimit, "rollbackLimit", b.rollBackLimit)
		}
		return true
	}
	b.OnMemberChanged(c)
	return false
}

func (this *baseChannelWatcher) OnUpgrade(opt Opt) (ChannelWatcher, []models.StartOption) {
	begin := time.Now()
	this.Logger.Info("开始upgrade")
	r, ops := this.impl.OnUpgrade(opt)
	cost := time.Now().Sub(begin)
	this.Logger.Info("upgrade结束", fmt.Sprintf("耗时:%f秒,%d毫秒", cost.Seconds(), cost.Milliseconds()))
	return r, ops
}
func (b *baseChannelWatcher) getStartChannelFromCtx(ctx *models.StartCTX) *[]ChannelMember {
	if ctx == nil {
		return nil
	}
	if ctx.Ctx == nil {
		return nil
	}
	v := ctx.Ctx.Value("channels")
	if nil != v {
		chmems := v.(*[]ChannelMember)
		return chmems
	}
	return nil
}
func (b *baseChannelWatcher) wrapNewCtxWithMember(mems []ChannelMember) context.Context {
	return context.WithValue(context.Background(), "channels", &mems)
}
func (this *baseChannelWatcher) protector() {
	tt := time.NewTicker(time.Second * 3)
	if this.mode != mode_debug {
		return
	}
	for {
		select {
		case <-this.debugC:
			return
		case <-tt.C:
			sb := strings.Builder{}
			this.mtx.Lock()
			for k, v := range this.operationCache {
				sb.WriteString(fmt.Sprintf("key:%v,data:%+v,dataType:%v", k, v, reflect.TypeOf(v)))
			}
			this.mtx.Unlock()
			this.Logger.Error("protector", "信息为:", sb.String())
		}
	}
}
func (this *baseChannelWatcher) OnStart(ctx *models.StartCTX) error {
	if err := this.listener.BStart(models.CtxStartOpt(ctx.Ctx)); nil != err {
		return err
	}
	if nil != this.routinePool {
		if err := this.routinePool.BStart(models.CtxStartOpt(ctx.Ctx)); nil != err {
			return err
		}
	}
	go this.protector()

	value := ctx.GetValue(config_channel_cap)
	if value == nil {
		value = default_channel_cap
	}
	chs, wg := this.GetChannelShims(value.(int))
	this.internalChs = chs
	this.wg.Add(wg)
	go this.dispath()
	return this.impl.OnStart(ctx)
}
func (this *baseChannelWatcher) GetChannelShims(cap int) (map[channel.ChannelID]*ChannelWp, int) {
	r, wg := this.impl.GetChannelShims(cap)
	if r == nil {
		r = make(map[channel.ChannelID]*ChannelWp)
	} else {
		for k, v := range r {
			if k == memberNotifyC {
				continue
			}
			if v.flush != nil {
				v.flush = this.wrapFlush(v.flush)
			}
		}
	}
	return r, wg
}
func (this *baseChannelWatcher) wrapFlush(f func(v channel.IData)) func(v channel.IData) {
	return func(v channel.IData) {
		f(v)
	}
}

func (this *baseChannelWatcher) handleDeltaMember() chan struct{} {
	chs := make([]*ChannelWp, 0)
	for _, v := range this.internalChs {
		chs = append(chs, v)
	}
	if atomic.LoadUint32(&this.status) != status_changing {
		panic(PROGRAMA_ERROR)
	}

	notify := make(chan struct{})

	this.wg.Add(len(chs))
	for index := range chs {
		go func(index int) {
			nodeCh := chs[index]
			defer this.wg.Done()
			for {
				select {
				case v, ok := <-nodeCh.ch.Ch:
					if !ok {
						return
					}
					nodeCh.flush(v)
				}
			}
		}(index)
	}
	return notify
}
func (this *baseChannelWatcher) dispath() {
	for {
		select {
		case env, ok := <-this.facadedC:
			if !ok {
				return
			}
			v, exist := this.internalChs[env.ChannelId]
			if !exist {
				panic(PROGRAMA_ERROR)
			}
			go func() {
				v.ch.Ch <- env.Data
			}()
		}
	}
}
func (this *baseChannelWatcher) sendMsg(chid channel.ChannelID, data channel.IData) {
	this.inc(data)
	select {
	case this.facadedC <- channel.Envelope{
		ChannelId: chid,
		Data:      data,
	}:
	default:
		this.Logger.Debug("消息阻塞", "channelId", chid, "data", data, "dataType", reflect.TypeOf(data))
		this.routinePool.AddJob(false, services2.Job{
			Handler: func() error {
				this.facadedC <- channel.Envelope{
					ChannelId: chid,
					Data:      data,
				}
				return nil
			},
		})
	}
}

type upgradeResp struct {
	r    ChannelWatcher
	opts []models.StartOption
}

func (this *baseChannelWatcher) execute(async bool, f func()) {
	if async {
		this.routinePool.AddJob(false, services2.Job{
			Handler: func() error {
				f()
				return nil
			},
		})
	} else {
		f()
	}
}

func (this *baseChannelWatcher) printMembers(m []ChannelMember) string {
	names := strings.Builder{}
	for _, v := range m {
		names.WriteString(v.name + ",")
	}
	return names.String()
}
