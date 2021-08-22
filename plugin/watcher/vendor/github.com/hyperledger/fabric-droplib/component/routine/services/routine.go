/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 7:16 下午
# @File : routine.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/routine/worker/gowp/workpoolv0"
	"sync/atomic"
)

type IRoutineComponent interface {
	base.IComponent
	AddJob(enableRoutine bool, job Job)
	JobsCount() int32
}

type Job struct {
	Pre     func()
	Handler workpoolv0.TaskHandler
	Post    func()
}

func (this Job) WrapHandler() workpoolv0.TaskHandler {
	return func() error {
		if nil != this.Pre {
			this.Pre()
		}
		defer func() {
			if nil != this.Post {
				this.Post()
			}
		}()
		return this.Handler()
	}
}

type defaultRoutinePool struct {
	*base.BaseComponent
	size int32
}

func (d *defaultRoutinePool) AddJob(enableRoutine bool, job Job) {
	atomic.AddInt32(&d.size, 1)
	go func() {
		defer atomic.AddInt32(&d.size, -1)
		job.WrapHandler()()
	}()
}

func (d *defaultRoutinePool) JobsCount() int32 {
	return atomic.LoadInt32(&d.size)
}
func NewDefaultGoRoutineNoLimitComponent() IRoutineComponent {
	r := &defaultRoutinePool{
		size: 0,
	}
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("ROUTINE_NOLIMIT", 1), r)
	return r
}
