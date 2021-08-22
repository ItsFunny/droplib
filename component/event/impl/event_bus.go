/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:41 上午
# @File : event_bus.go
# @Description :
# @Attention :
*/
package impl

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/event/services"
	"github.com/hyperledger/fabric-droplib/component/event/types"
	"github.com/libp2p/go-libp2p-core/event"
	"reflect"
	"sync"
)

var (
	_ services.IP2PEventBusComponent = (*P2PEventBusComponent)(nil)
)

type P2PEventBusComponent struct {
	sync.RWMutex
	*base.BaseComponent

	bus event.Bus

	events   map[reflect.Type]*SubscriptionWrapper
	emitters map[reflect.Type]event.Emitter

	eventNameSpaceMap map[types.EventNameSpace]reflect.Type

}

func NewLibP2PEventBusComponent(bus event.Bus) *P2PEventBusComponent {
	r := &P2PEventBusComponent{
		bus:               bus,
		events:            make(map[reflect.Type]*SubscriptionWrapper),
		emitters:          make(map[reflect.Type]event.Emitter),
		eventNameSpaceMap: make(map[types.EventNameSpace]reflect.Type),
	}
	m := modules.NewModule("EVENT_BUS", 1)
	r.BaseComponent = base.NewBaseComponent(m, r)
	return r
}

func (e *P2PEventBusComponent) OnStop(ctx *models.StopCTX) {
	for _, v := range e.events {
		v.subscription.Close()
	}
}

func (e *P2PEventBusComponent) Setup(m map[types.EventNameSpace]interface{}, subOpts []event.SubscriptionOpt, emiOps ...event.EmitterOpt) error {
	// for k, v := range m {
	// 	_, err := e.SubscribeWithEmit(k, v, subOpts, emiOps...)
	// 	if nil != err {
	// 		panic(err)
	// 	}
	// }
	return nil
}

func (e *P2PEventBusComponent) SubscribeWithEmit(eventNameSpace types.EventNameSpace, eventType interface{}, subOpts []event.SubscriptionOpt, emiOps ...event.EmitterOpt) error {
	e.Lock()
	defer e.Unlock()
	t := reflect.TypeOf(eventType)
	if _, exist := e.eventNameSpaceMap[eventNameSpace]; exist {
		e.Logger.Warn("重复注册event:" + t.String())
		return nil
	}
	e.eventNameSpaceMap[eventNameSpace] = t

	e.Logger.Info(fmt.Sprintf("subscribe  event,namespace:%s,type:%T", eventNameSpace, eventType))

	subscribe, err := e.bus.Subscribe(eventType, subOpts...)
	if nil != err {
		return err
	}
	e.events[t] = &SubscriptionWrapper{
		subscription: subscribe,
	}
	emitter, err := e.bus.Emitter(eventType, emiOps...)
	if nil != err {
		return err
	}
	e.emitters[t] = emitter

	return nil
}
func (e *P2PEventBusComponent) GetBus() event.Bus {
	return e.bus
}
func (e *P2PEventBusComponent) Publish(space types.EventNameSpace, data interface{}) error {
	return e.publish(e.eventNameSpaceMap[space], data)
}

func (e *P2PEventBusComponent) publish(t reflect.Type, data interface{}) error {
	return e.emitters[t].Emit(data)
}

func (e *P2PEventBusComponent) GetRegisteredEvents() map[reflect.Type]*SubscriptionWrapper {
	// e.WaitUntilReady()

	res := make(map[reflect.Type]*SubscriptionWrapper)
	for k, v := range e.events {
		res[k] = v
	}
	return res
}

func (e *P2PEventBusComponent) AcquireEventChanByType(t reflect.Type) <-chan interface{} {
	// e.WaitUntilReady()
	e.RLock()
	defer e.RUnlock()

	v, exist := e.events[t]
	if !exist {
		panic(fmt.Sprintf("该类型:%T 未注册event:", t))
	}

	if v.status == ACQUIRED {
		panic(fmt.Sprintf("重复获取,该类型已经被获取,类型为%v", t))
	}
	v.status = ACQUIRED

	return v.subscription.Out()
}

func (e *P2PEventBusComponent) AcquireEventChanByNameSpace(nameSpace types.EventNameSpace) <-chan interface{} {
	// e.WaitUntilReady()

	e.RLock()
	defer e.RUnlock()

	t, exist := e.eventNameSpaceMap[nameSpace]
	if !exist {
		panic("nameSpace为:" + nameSpace + " 不存在")
	}
	v, exist := e.events[t]
	if !exist {
		panic(fmt.Sprintf("该类型:%T 未注册event:", t))
	}
	if v.status == ACQUIRED {
		panic(fmt.Sprintf("重复获取,该类型已经被获取,类型为%v", t))
	}
	v.status = ACQUIRED

	return v.subscription.Out()
}
