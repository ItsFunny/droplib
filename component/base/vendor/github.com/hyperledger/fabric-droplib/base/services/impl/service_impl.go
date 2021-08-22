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
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/constants"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	_ services.IBaseService = (*BaseServiceImpl)(nil)
)

var (
	// ErrAlreadyStarted is returned when somebody tries to start an already
	// running service.
	ErrAlreadyStarted = errors.New("already started")
	// ErrAlreadyStopped is returned when somebody tries to stop an already
	// stopped service (without resetting it).
	ErrAlreadyStopped = errors.New("already stopped or flushed")
	// ErrNotStarted is returned when somebody tries to stop a not running
	// service.
	ErrNotStarted = errors.New("not started")

	ErrNotRightStatus = errors.New("wrong status")

	ErrAlreadyReady = errors.New("already ready")
)

type BaseServiceImpl struct {
	Logger  v2.Logger
	name    string
	started uint32 // atomic
	stopped uint32 // atomic
	quit    chan struct{}
	impl    services.IBaseService
}

func (b *BaseServiceImpl) waitUntilReady() {
	c := func() bool {
		return b.started == constants.READY
	}
	for !c() {
		b.Logger.Info("该 [" + b.impl.String() + "],未就绪,阻塞中~~~")
		time.Sleep(time.Second * 1)
	}
}
func (b *BaseServiceImpl) waitUntilStart() {
	c := func() bool {
		return b.started == constants.STARTED
	}
	for !c() {
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
	status := atomic.LoadUint32(&bs.started)
	ctx := &models.ReadyCTX{}
	for _, op := range ops {
		op(ctx)
	}

	if status == constants.NONE {
		if ctx.ReadyFlag&constants.READY_UNTIL_START > 0 {
			bs.waitUntilReady()
		} else {
			bs.Logger.Error("服务状态错误,ready之前必须先start,当前状态为:", status)
			debug.PrintStack()
			return ErrNotRightStatus
		}
	}
	if !atomic.CompareAndSwapUint32(&bs.started, constants.STARTED, constants.READY) {
		st := atomic.LoadUint32(&bs.stopped)
		if st&constants.STOP > 0 {
			bs.Logger.Error("当前处于停止或者flush状态,不处于started状态,无法ready", "name:", bs.name, "状态为:", st, "impl", bs.impl)
			atomic.StoreUint32(&bs.started, constants.NONE)
			return ErrAlreadyStopped
		} else {
			bs.Logger.Info("already ready, impl:", bs.impl)
			return nil
		}
	}

	bs.Logger.Info("notify as ready service ,impl:", bs.impl)
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
	return nil
}

func (bs *BaseServiceImpl) OnReady(ctx *models.ReadyCTX) error {
	return nil
}

// 默认start 为panic
// func (bs *BaseServiceImpl) Start() error {
// 	return bs.start()
// }

func (bs *BaseServiceImpl) start(ctx *models.StartCTX) error {
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
		}
		bs.Logger.Info(fmt.Sprintf("Starting %v service,impl:%v", bs.name, bs.impl))
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
		return nil
	}
	bs.Logger.Debug(fmt.Sprintf("Not starting %v service -- already started", bs.name), "impl", bs.impl)
	return ErrAlreadyStarted
}

func (bs *BaseServiceImpl) BStart(opts ...models.StartOption) error {
	c := make(chan error)
	ctx := &models.StartCTX{}
	for _, op := range opts {
		op(ctx)
	}

	if ctx.Flag&constants.SYNC_START > 0 {
		go func() {
			err := bs.start(ctx)
			c <- err
			close(c)
		}()
		return <-c
	} else {
		go func() {
			if err := bs.start(ctx); nil != err {
				bs.Logger.Error("启动失败,impl:", bs.impl, " error:", err.Error())
			}
		}()
		close(c)
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
		quit:   make(chan struct{}),
		impl:   concreteImpl,
	}
	return res
}

func (bs *BaseServiceImpl) OnStart(ctx *models.StartCTX) error {
	return nil
}

func (bs *BaseServiceImpl) BStop(ops ...models.StopOption) error {
	ctx := &models.StopCTX{}
	for _, op := range ops {
		op(ctx)
	}
	if atomic.CompareAndSwapUint32(&bs.stopped, 0, 1) {
		if atomic.LoadUint32(&bs.started) == 0 {
			bs.Logger.Error(fmt.Sprintf("状态非处于store状态 %v service ", bs.name),
				"impl", bs.impl)
			// revert flag
			atomic.StoreUint32(&bs.stopped, 0)
			return ErrNotStarted
		}
		bs.Logger.Info(fmt.Sprintf("Stopping %v service", bs.name), "impl", bs.impl)
		bs.impl.OnStop(ctx)
		close(bs.quit)
		return nil
	}
	bs.Logger.Debug(fmt.Sprintf("停止 %v service (already stopped)", bs.name), "impl", bs.impl)
	return ErrAlreadyStopped
}

func (bs *BaseServiceImpl) OnStop(ctx *models.StopCTX) {
	bs.impl.OnStop(ctx)
}

func (bs *BaseServiceImpl) Reset() error {
	bs.Logger.Info("reset 开始重新初始化service")
	if !atomic.CompareAndSwapUint32(&bs.stopped, 1, 0) {
		bs.Logger.Debug(fmt.Sprintf("reset 状态设置失败%v service. Not stopped", bs.name), "impl", bs.impl)
		return fmt.Errorf("can't reset running %s", bs.name)
	}

	// whether or not we've started, we can reset
	atomic.CompareAndSwapUint32(&bs.started, 1, 0)

	bs.quit = make(chan struct{})
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
	return bs.quit
}

func (bs *BaseServiceImpl) String() string {
	return bs.name
}

func (bs *BaseServiceImpl) SetLogger(logger v2.Logger) {
	bs.Logger = logger
}
