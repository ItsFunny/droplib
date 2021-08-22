/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 6:53 下午
# @File : routine.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"github.com/hyperledger/fabric-droplib/component/routine/worker/gowp/workpoolv0"
	"runtime"
	"sync/atomic"
)

var (
	_ services.IRoutineComponent = (*RoutineComponent)(nil)
)

type RoutineComponent struct {
	*base.BaseComponent
	pool  *workpoolv0.WorkPool
	count int32

	syncLimit int32
}

func (this *RoutineComponent) JobsCount() int32 {
	return atomic.LoadInt32(&this.count)
}

func (this *RoutineComponent) AddJob(enableRoutine bool, job services.Job) {
	c := atomic.AddInt32(&this.count, 1)
	// if enableRoutine || c > this.syncLimit {
	// 	go this.pool.Do(this.wrapJob(job))
	// } else {
	// 	this.pool.Do(this.wrapJob(job))
	// }
	if enableRoutine {
		go this.pool.Do(this.wrapJob(job))
	} else if c > this.syncLimit {
		this.Logger.Warn("任务过多,消费过慢", "当前任务数:", c)
		go this.pool.Do(this.wrapJob(job))
	} else {
		this.pool.Do(this.wrapJob(job))
	}
}
func (this *RoutineComponent) wrapJob(job services.Job) workpoolv0.TaskHandler {
	return func() error {
		defer func() {
			atomic.AddInt32(&this.count, -1)
		}()
		return job.WrapHandler()()
	}
}

func NewRoutineComponent(opts ...Opt) *RoutineComponent {
	r := &RoutineComponent{}
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("ROUTINE", 1), r)
	for _, opt := range opts {
		opt(r)
	}
	if r.pool == nil {
		// small job
		r.pool = workpoolv0.New(runtime.NumCPU() * 2)
	}
	r.syncLimit = 256
	return r
}

func (this *RoutineComponent) OnStop(ctx *models.StopCTX) {
}
