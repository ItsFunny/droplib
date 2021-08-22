/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 4:47 下午
# @File : task_queue.go
# @Description :
# @Attention :
*/
package v2

type TaskWrapper struct {
	Task func()
}

type TaskQueue interface {
	Take() ITask
	Push(task ITask) (int, error)
	PushDisableTask()
}

var (
	_ TaskQueue = (*defaultChannelTaskQueue)(nil)
)

type defaultChannelTaskQueue struct {
	task chan ITask
}

func NewDefaultChannelTaskQueue(f func() int) *defaultChannelTaskQueue {
	r := defaultChannelTaskQueue{}
	r.task = make(chan ITask, f())
	return &r
}

func (d *defaultChannelTaskQueue) Take() ITask {
	return <-d.task
}

func (d *defaultChannelTaskQueue) Push(task ITask) (int, error) {
	d.task <- task
	return len(d.task), nil
}
func (d *defaultChannelTaskQueue) PushDisableTask() {
	d.task <- nil
}
