/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/12 12:39 下午
# @File : c.go
# @Description :
# @Attention :
*/
package watcher

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/listener"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"sync/atomic"
	"time"
)

var (
	id int32
)

func acquireId() int {
	addInt32 := atomic.AddInt32(&id, 1)
	return int(addInt32)
}

type c struct {
	id int
	// halt     <-chan byte
	notifyC  chan struct{}
	c        <-chan channel.IData
	listener <-chan interface{}
	consumer DataConsumer
	priority uint16
	name     string
	status   uint32

	startt time.Time
}

func releaseC(cc *c) {
	// FIXME
	atomic.StoreUint32(&cc.status, routinec_status_release)
	cc.c = nil
	cc.consumer = nil
}

func newC(name string, cc <-chan channel.IData, f DataConsumer, ff func() <-chan interface{}) *c {
	r := &c{
		notifyC:  make(chan struct{}),
		c:        cc,
		consumer: f,
		name:     name,
		status:   status_ok,
		id:       acquireId(),
		startt:   time.Now(),
	}
	if nil != ff {
		l := ff()
		r.listener = l
	}
	return r
}
func (this *c) close(l listener.IListenerComponent) <-chan interface{} {
	//  spin is better
	if atomic.CompareAndSwapUint32(&this.status, routinec_status_running, routinec_status_wait_listener) {
		close(this.notifyC)
		return this.listener
	} else if this.released() {
		return this.listener
	} else if atomic.LoadUint32(&this.status) == routinec_status_ok {
		this.status=routinec_status_upgrade
		l.NotifyListener(nil, this.name)
		return this.listener
	} else {
		return this.listener
	}
}

func (this *c) stopped() bool {
	return atomic.LoadUint32(&this.status)&routinec_status_upgrade >= routinec_status_upgrade
}
func (this *c) closed() bool {
	return atomic.LoadUint32(&this.status)&routinec_status_close >= routinec_status_close
}
func (this *c) released() bool {
	return atomic.LoadUint32(&this.status) == routinec_status_release
}
func (this *c) panicCAS(exce, newV uint32) {
	if !atomic.CompareAndSwapUint32(&this.status, exce, newV) {
		panic(fmt.Sprintf("%d,%d,%d", this.status, exce, newV))
	}
}
func (this *c) cas(exce, newV uint32, f func()) {
	if atomic.CompareAndSwapUint32(&this.status, exce, newV) {
		f()
	}
}

func (this *c) fromOther(cc *c) {
	this.c = cc.c
	this.name = cc.name
	this.consumer = cc.consumer
}
func (this *c) fromMember(m ChannelMember, l listener.IListenerComponent) {
	this.name = m.name
	this.consumer = m.consumer
	this.c = m.c
	this.listener = l.RegisterListener(m.name)
	this.notifyC = make(chan struct{})
	atomic.StoreUint32(&this.status, routinec_status_ok)
}
