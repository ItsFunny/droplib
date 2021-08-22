/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/5 8:44 上午
# @File : internal_pubsub.go
# @Description :
# @Attention :
*/
package models


import (
	"errors"
	"sync"
)

var (
	ErrUnsubscribed = errors.New("client unsubscribed")
	ErrOutOfCapacity = errors.New("client is not pulling messages fast enough")
)

type SubscriptionImpl struct {
	out chan PubSubMessage

	canceled chan struct{}
	mtx      sync.RWMutex
	err      error
}


func NewSubscription(outCapacity int) *SubscriptionImpl {
	return &SubscriptionImpl{
		out:      make(chan PubSubMessage, outCapacity),
		canceled: make(chan struct{}),
	}
}
func(this *SubscriptionImpl)GetOut()chan PubSubMessage{
	return this.out
}

func (s *SubscriptionImpl) Out() <-chan PubSubMessage {
	return s.out
}

func (s *SubscriptionImpl) Canceled() <-chan struct{} {
	return s.canceled
}

func (s *SubscriptionImpl) Err() error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.err
}

func (s *SubscriptionImpl) Cancel(err error) {
	s.mtx.Lock()
	s.err = err
	s.mtx.Unlock()
	close(s.canceled)
}

type PubSubMessage struct {
	data   interface{}
	events map[string][]string
}



func NewPubSubMessage(data interface{}, events map[string][]string) PubSubMessage {
	return PubSubMessage{data, events}
}

func (msg PubSubMessage) Data() interface{} {
	return msg.data
}

func (msg PubSubMessage) Events() map[string][]string {
	return msg.events
}