/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/1 4:11 下午
# @File : service_impl.go
# @Description :    基础services 类
# @Attention :
*/
package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/constants"
	"github.com/hyperledger/fabric-droplib/base/services/e"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"sync/atomic"
	"time"
)

var (
	_ services.IBaseService = (*BaseServiceImpl)(nil)
)

var (
	ErrAlreadyStarted = errors.New("already started")
	ErrAlreadyStopped = errors.New("already stopped or flushed")
	ErrNotStarted     = errors.New("not started")
	ErrNotRightStatus = errors.New("wrong status")
	ErrAlreadyReady   = errors.New("already ready")
	ErrCanceled       = errors.New("取消启动")
)

type BaseServiceImpl struct {
	Logger  v2.Logger
	name    string
	started uint32 // atomic
	stopped uint32 // atomic
	ctx     context.Context
	cancel  func()
	impl    services.IBaseService
}

// false: 需要关闭整个service
func (b *BaseServiceImpl) waitUntilReady() {
	c := func() bool {
		return atomic.LoadUint32(&b.started)&constants.READY >= constants.READY
	}
	for !c() {
		// FIXME ,添加 task
		select {
		case <-b.ctx.Done():
			b.Logger.Info("该 [" + b.impl.String() + "],收到cancel通知,结束等待")
			return
		default:
		}
		b.Logger.Info("该 [" + b.impl.String() + "],未就绪,阻塞中~~~")
		time.Sleep(time.Second * 1)
	}
}

func (b *BaseServiceImpl) waitUntilStart() {
	c := func() bool {
		return atomic.LoadUint32(&b.started) == constants.STARTED
	}
	for !c() {
		// FIXME ,添加 task
		select {
		case <-b.ctx.Done():
			b.Logger.Info("该 [" + b.impl.String() + "],收到cancel通知,结束等待")
			return
		default:
		}
		b.Logger.Info("该 [" + b.impl.String() + "],未 start,阻塞中~~~")
		time.Sleep(time.Second * 1)
	}
}

func (b *BaseServiceImpl) ReadyOrNot() bool {
	return atomic.LoadUint32(&b.started) == constants.READY
}

// func (bs *BaseServiceImpl) ReadyIfPanic(flag services.READY_FALG) {
// 	if err := bs.Ready(flag); nil != err {
// 		debug.PrintStack()
// 		panic(err)
// 	}
// }

func (bs *BaseServiceImpl) BReady(ops ...models.ReadyOption) error {
	ctx := &models.ReadyCTX{}
	for _, op := range ops {
		op(ctx)
	}

	if ctx.ReadyFlag.Sync() {
		return bs.ready(ctx)
	} else {
		go func() {
			if err := bs.ready(ctx); nil != err {
				bs.Logger.Error("ready失败,impl:", bs.impl, " error:", err.Error())
				if err == e.FORCELOSE_ERROR {
					bs.BStop(models.StopCTXWithForce)
				}
			}
		}()
	}
	return nil
}

func (bs *BaseServiceImpl) OnReady(ctx *models.ReadyCTX) error {
	return nil
}

func (bs *BaseServiceImpl) ready(ctx *models.ReadyCTX) error {
	status := atomic.LoadUint32(&bs.started)
	if status == constants.ON_READY || status == constants.READY {
		bs.Logger.Error("服务状态错误,已经处于on_ready|ready状态")
		return ErrAlreadyReady
	}
	if ctx.ReadyFlag&constants.READY_UNTIL_START >= constants.READY_UNTIL_START {
		bs.waitUntilStart()
		select {
		case <-bs.Quit():
			return ErrCanceled
		default:
		}
	}
	if atomic.CompareAndSwapUint32(&bs.started, constants.STARTED, constants.ON_READY) {
		if atomic.LoadUint32(&bs.stopped) == constants.STOP {
			bs.Logger.Error(fmt.Sprintf("不处于start 状态 %v service -- 处于宕机状态", bs.name),
				"impl", bs.impl)
			// revert flag
			atomic.StoreUint32(&bs.started, constants.NONE)
			return ErrAlreadyStopped
		}
	} else {
		if atomic.LoadUint32(&bs.stopped) == constants.STOP {
			bs.Logger.Error(fmt.Sprintf("不处于start 状态 %v service -- 处于宕机状态", bs.name),
				"impl", bs.impl)
			// revert flag
			atomic.StoreUint32(&bs.started, constants.NONE)
			return ErrAlreadyStopped
		}
	}
	bs.Logger.Info("进入[PRE-READY]状态,impl:", bs.impl)
	if nil != ctx.PreReady {
		ctx.PreReady()
	}
	err := bs.impl.OnReady(ctx)
	if nil != err {
		bs.Logger.Error("ready 失败:", err.Error())
		atomic.StoreUint32(&bs.started, constants.NONE)
		return err
	}
	if nil != ctx.PostReady {
		ctx.PostReady()
	}
	if !atomic.CompareAndSwapUint32(&bs.started, constants.ON_READY, constants.READY) {
		st := atomic.LoadUint32(&bs.stopped)
		if st&constants.STOP > 0 {
			bs.Logger.Error("当前处于停止或者flush状态,不处于ON_READY状态,无法ready", "name:", bs.name, "状态为:", st, "impl", bs.impl)
			atomic.StoreUint32(&bs.started, constants.NONE)
			return ErrAlreadyStopped
		} else {
			bs.Logger.Info("already ready, impl:", bs.impl)
			return nil
		}
	}
	bs.Logger.Info("服务进入[READY]状态,impl:", bs.impl)
	return nil
}

func (bs *BaseServiceImpl) start(ctx *models.StartCTX) error {
	now := time.Now()
	status := atomic.LoadUint32(&bs.started)
	if status == constants.READY {
		bs.Logger.Error("服务状态错误,已经处于ready状态")
		return ErrAlreadyReady
	}
	if atomic.CompareAndSwapUint32(&bs.started, constants.NONE, constants.STARTED) {
		if atomic.LoadUint32(&bs.stopped) == constants.STOP {
			bs.Logger.Error(fmt.Sprintf("不处于start 状态 %v service -- 处于宕机状态", bs.name),
				"impl", bs.impl)
			// revert flag
			atomic.StoreUint32(&bs.started, constants.NONE)
			return ErrAlreadyStopped
		}
		if ctx.Flag&constants.WAIT_READY > 0 {
			bs.waitUntilReady()
			// defer
			// 或许是因为cancel 而退出
			select {
			case <-bs.Quit():
				return ErrCanceled
			default:
			}
		}

		bs.Logger.Info(fmt.Sprintf("准备启动 Pre Starting %v service,impl:%v", bs.name, bs.impl))
		if nil != ctx.PreStart {
			ctx.PreStart()
		}
		err := bs.impl.OnStart(ctx)
		if err != nil {
			// revert flag
			atomic.StoreUint32(&bs.started, constants.NONE)
			return err
		}
		if nil != ctx.PostStart {
			ctx.PostStart()
		}

		cost := time.Now().Sub(now)
		bs.Logger.Info(fmt.Sprintf("成功启动服务:%v ,impl:%v,耗时[%v]毫秒(%v秒)", bs.name, bs.impl, cost.Milliseconds(), cost.Seconds()))
		return nil
	}
	bs.Logger.Debug(fmt.Sprintf("Not starting %v service -- already started", bs.name), "impl", bs.impl)
	return ErrAlreadyStarted
}

func (bs *BaseServiceImpl) BStart(opts ...models.StartOption) error {
	ctx := &models.StartCTX{}
	for _, op := range opts {
		op(ctx)
	}
	if ctx.Ctx == nil {
		ctx.Ctx = context.Background()
	}
	bs.ctx, bs.cancel = context.WithCancel(ctx.Ctx)

	if ctx.Flag == 0 {
		ctx.Flag = constants.SYNC_START
	}

	if ctx.Flag.Sync() {
		return bs.start(ctx)
	} else {
		go func() {
			if err := bs.start(ctx); nil != err {
				bs.Logger.Error("启动失败,impl:", bs.impl, " error:", err.Error())
				if err == e.FORCELOSE_ERROR {
					bs.BStop(models.StopCTXWithForce)
				}
			}
		}()
	}
	return nil
}

func NewBaseService(logger v2.Logger, m modules.Module, concreteImpl services.IBaseService) *BaseServiceImpl {
	if logger == nil {
		logger = logrusplugin.NewLogrusLogger(m)
	}
	res := &BaseServiceImpl{
		Logger: logger,
		name:   m.String(),
		impl:   concreteImpl,
		ctx:    context.Background(),
	}
	return res
}

func (bs *BaseServiceImpl) OnStart(ctx *models.StartCTX) error {
	return nil
}

func (bs *BaseServiceImpl) BStop(ops ...models.StopOption) error {
	ctx := &models.StopCTX{
		Value: make(map[string]interface{}, 1),
	}
	for _, op := range ops {
		op(ctx)
	}
	value := uint32(constants.STOP)
	if ctx.Force {
		bs.cancel()
		atomic.StoreUint32(&bs.stopped, value)
		atomic.StoreUint32(&bs.started, 0)
		return nil
	}

	if atomic.CompareAndSwapUint32(&bs.stopped, 0, value) {
		if atomic.LoadUint32(&bs.started) == 0 {
			bs.Logger.Error(fmt.Sprintf("状态非处于start状态 %v service ", bs.name),
				"impl", bs.impl)
			// revert flag
			atomic.StoreUint32(&bs.stopped, 0)
			return ErrNotStarted
		}
		bs.Logger.Info(fmt.Sprintf("Stopping %v service", bs.name), "impl", bs.impl)
		bs.cancel()
		bs.impl.OnStop(ctx)
		return nil
	}
	bs.Logger.Debug(fmt.Sprintf("停止 %v service (already stopped)", bs.name), "impl", bs.impl)
	return ErrAlreadyStopped
}

func (bs *BaseServiceImpl) OnStop(ctx *models.StopCTX) {
}

func (bs *BaseServiceImpl) Reset() error {
	bs.Logger.Info("reset 开始重新初始化service")
	if !atomic.CompareAndSwapUint32(&bs.stopped, 1, 0) {
		bs.Logger.Debug(fmt.Sprintf("reset 状态设置失败%v service. Not stopped", bs.name), "impl", bs.impl)
		return fmt.Errorf("can't reset running %s", bs.name)
	}

	// whether or not we've started, we can reset
	atomic.CompareAndSwapUint32(&bs.started, 1, 0)

	return bs.impl.OnReset()
}

func (bs *BaseServiceImpl) OnReset() error {
	panic("The service cannot be reset")
}

func (bs *BaseServiceImpl) IsRunning() bool {
	r := atomic.LoadUint32(&bs.started)
	return (r == constants.STARTED || r == constants.READY) && atomic.LoadUint32(&bs.stopped) == 0
}

func (bs *BaseServiceImpl) Quit() <-chan struct{} {
	return bs.ctx.Done()
}

func (bs *BaseServiceImpl) String() string {
	return bs.name
}

func (bs *BaseServiceImpl) SetLogger(logger v2.Logger) {
	bs.Logger = logger
}

func (bs *BaseServiceImpl) CtxWithValue(key interface{}, value interface{}) context.Context {
	return context.WithValue(bs.ctx, key, value)
}

func (bs *BaseServiceImpl) Started() bool {
	return atomic.LoadUint32(&bs.started) == constants.STARTED
}
