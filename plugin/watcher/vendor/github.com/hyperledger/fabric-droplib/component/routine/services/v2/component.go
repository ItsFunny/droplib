/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 9:41 下午
# @File : component.go
# @Description :
# @Attention :
*/
package v2

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"sync/atomic"
)

var (
	_ services.IRoutineComponent = (*routineComponentV2)(nil)
)

type routineComponentV2 struct {
	*base.BaseComponent
	pool *Pool
	size int32
}

var (
	PriorityTask Option = func(opts *Options) {
		opts.TaskQueue = func() TaskQueue {
			return NewPriorityTaskQueue()
		}
	}
)

func NewSingleRoutingPoolExecutor(opts ...Option) *routineComponentV2 {
	r := &routineComponentV2{}
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("ROUTINE_V2", 1), r)
	pool, _ := NewPool(append(opts, WithSize(1))...)
	r.pool = pool
	return r
}

func NewV2RoutinePoolExecutorComponent(opts ...Option) *routineComponentV2 {
	r := &routineComponentV2{}
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("ROUTINE_V2", 1), r)
	pool, _ := NewPool(opts...)
	r.pool = pool
	return r
}

func (r *routineComponentV2) AddJob(enableRoutine bool, job services.Job) {
	if err := r.pool.Submit(func() {
		job.WrapHandler()()
		atomic.AddInt32(&r.size, -1)
	}); nil != err {
		r.Logger.Warn("添加job失败", "err", err.Error())
	} else {
		atomic.AddInt32(&r.size, 1)
	}
}

func (r *routineComponentV2) JobsCount() int32 {
	return atomic.LoadInt32(&r.size)
}

func (r *routineComponentV2) OnStop(c *models.StopCTX) {
	r.pool.Release()
}
