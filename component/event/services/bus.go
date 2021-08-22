/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/12 2:47 下午
# @File : bus.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/event/types"
	"github.com/libp2p/go-libp2p-core/event"
	"reflect"
)

// type IEventBus interface {
// 	Subscribe(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error)
// 	Emitter(eventType interface{}, opts ...event.EmitterOpt) (event.Emitter, error)
// 	GetAllEventTypes() []reflect.Type
// }

type IP2PEventBusComponent interface {
	base.IComponent

	SubscribeWithEmit(eventNameSpace types.EventNameSpace, eventType interface{}, subOpts []event.SubscriptionOpt, emiOps ...event.EmitterOpt) error
	Setup(m map[types.EventNameSpace]interface{}, subOpts []event.SubscriptionOpt, emiOps ...event.EmitterOpt) error
	Publish(eventNameSpace types.EventNameSpace, data interface{}) error

	GetBus()event.Bus
	AcquireEventChanByType(t reflect.Type) <-chan interface{}
	AcquireEventChanByNameSpace(nameSpace types.EventNameSpace) <-chan interface{}
}