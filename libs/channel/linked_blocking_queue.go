/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/14 6:23 下午
# @File : linked_blocking_queue.go
# @Description :
# @Attention :
*/
package channel

import (
	"github.com/hyperledger/fabric-droplib/structure/lists/arraylist"
	"sync"
)

var (
	_ IChan = (*linkedBlockingChan)(nil)
)

type linkedBlockingChan struct {
	sync.Mutex
	condition *sync.Cond
	list      *arraylist.List
}

func NewLinkedBlockQueue() *linkedBlockingChan {
	r := &linkedBlockingChan{}
	r.condition = sync.NewCond(&r.Mutex)
	r.list = arraylist.New()
	return r
}
func (l *linkedBlockingChan) Take() IData {
	c := l.condition
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	for l.list.Size() == 0 {
		c.Wait()
	}
	if l.list.Size() > 0 {
		task, _ := l.list.Get(0)
		l.list.Remove(0)
		return task.(IData)
	}
	return nil
}

func (l *linkedBlockingChan) Push(task IData) (int, error) {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	l.list.Add(task)
	l.condition.Signal()
	return l.list.Size(), nil
}
