/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 10:33 下午
# @File : task_queue_linked_blocking.go
# @Description :
# @Attention :
*/
package v2

import (
	"github.com/emirpasic/gods/lists/arraylist"
	"sync"
)

var (
	_ TaskQueue = (*linkedBlockingQueue)(nil)
)

type linkedBlockingQueue struct {
	sync.Mutex
	condition *sync.Cond
	list      *arraylist.List
}

func newLinkedBlockQueue() *linkedBlockingQueue {
	r := &linkedBlockingQueue{}
	r.condition = sync.NewCond(&r.Mutex)
	r.list = arraylist.New()
	return r
}
func (l *linkedBlockingQueue) Take() ITask {
	c := l.condition
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	for l.list.Size() == 0 {
		c.Wait()
	}
	if l.list.Size() > 0 {
		task, _ := l.list.Get(0)
		l.list.Remove(0)
		return task.(ITask)
	}
	return nil
}

func (l *linkedBlockingQueue) Push(task ITask) (int, error) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.list.Add(task)
	l.condition.Signal()
	return l.list.Size(), nil
}

func (l *linkedBlockingQueue) PushDisableTask() {
	l.Push(newDefaultDisableTask())
}
