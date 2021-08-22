/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 4:48 下午
# @File : task.go
# @Description :
# @Attention :
*/
package v2

type TaskStatus byte

var (
	_ ITask = defaultTask{}
)

var (
	DISABLE   = TaskStatus(1 << 0)
	AVAILABLE = TaskStatus(1 << 1)
	REINQUEUE = 1<<2 | AVAILABLE
)

type ITask interface {
	Execute()
	Status() TaskStatus
	// status
}

func (t TaskStatus) available() bool {
	return t&AVAILABLE >= AVAILABLE
}
func (t TaskStatus) mustConsume() bool {
	return t&REINQUEUE >= REINQUEUE
}

type defaultTask struct {
	f      func()
	status TaskStatus
}

func (d defaultTask) Execute() {
	d.f()
}
func (d defaultTask) Available() bool {
	return true
}
func (d defaultTask) Status() TaskStatus {
	return d.status
}

func NewDefaultTask(f func()) defaultTask {
	return newDefaultTask(f, AVAILABLE)
}

func newDefaultTask(f func(), status TaskStatus) defaultTask {
	r := defaultTask{f: f}
	r.status = status
	return r
}
