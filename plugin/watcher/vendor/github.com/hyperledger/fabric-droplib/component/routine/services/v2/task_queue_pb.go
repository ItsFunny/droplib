/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 5:09 下午
# @File : task_queue_pb.go
# @Description :
# @Attention :
*/
package v2

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/libs/queue/concurrentpriority"
	"sync"
)

// ////////////////
var (
	_             IPriorityTaskQueue = (*PriorityTaskQueue)(nil)
	WrongTaskType                    = errors.New("wrong task type")
	emptyFunc     func()
	max_priority  = 100
)

type IPriorityTask interface {
	ITask
	GetPriority() int
	ResetProperty(p int) IPriorityTask
}
type defaultPriorityTask struct {
	defaultTask
	priority int
	disable  bool
}

func (d defaultPriorityTask) GetPriority() int {
	return d.priority
}
func (d defaultPriorityTask) ResetProperty(p int) IPriorityTask {
	d.priority = p
	return d
}
func (d defaultPriorityTask) Available() bool {
	return !d.disable
}

func NewDefaultPriorityTask(f func(), pri int) IPriorityTask {
	// FIXME ,添加cache
	return defaultPriorityTask{
		defaultTask: NewDefaultTask(f),
		priority:    pri,
	}
}
func newDefaultDisableTask() IPriorityTask {
	r := defaultPriorityTask{
		defaultTask: newDefaultTask(emptyFunc, DISABLE),
		priority:    max_priority,
		disable:     true,
	}
	return r
}

type IPriorityTaskQueue interface {
	TaskQueue
}
type PriorityTaskQueue struct {
	sync.Mutex
	condition *sync.Cond
	queue     *concurrentpriority.PriorityQueue
}

func NewPriorityTaskQueue() *PriorityTaskQueue {
	r := &PriorityTaskQueue{}
	r.condition = sync.NewCond(&r.Mutex)
	r.queue = concurrentpriority.NewPriorityQueue(func(a, b interface{}) int {
		a1 := a.(IPriorityTask)
		a2 := a.(IPriorityTask)
		return a1.GetPriority() - a2.GetPriority()
	})
	return r
}

func (p *PriorityTaskQueue) PushDisableTask() {
	p.Push(newDefaultDisableTask())
}

func (p *PriorityTaskQueue) Push(task ITask) (int,error) {
	pt, ok := task.(IPriorityTask)
	if !ok {
		return 0,WrongTaskType
	}
	if pt.GetPriority() > max_priority {
		pt = pt.ResetProperty(max_priority)
	}
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.queue.Push(pt)
	p.condition.Signal()
	return p.queue.Len(),nil
}

func (p *PriorityTaskQueue) Take() ITask {
	c := p.condition
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	for p.queue.Len() == 0 {
		c.Wait()
	}
	if p.queue.Len() > 0 {
		task := p.queue.Pop()
		return task.(ITask)
	}
	return nil
}
