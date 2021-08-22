/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:08 上午
# @File : event_manager_component.go
# @Description :
# @Attention :
*/
package v2

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"sync"
)

var (
	_ IListenerComponentV2 = (*eventComponent)(nil)
)

type eventComponent struct {
	*impl.BaseServiceImpl
	mutx       sync.RWMutex
	eventCells map[EventNameSpace]*eventCell
	listeners  map[string]*eventListener
}

func NewDefaultEventComponent() *eventComponent {
	r := &eventComponent{
		eventCells: make(map[EventNameSpace]*eventCell),
		listeners:  make(map[string]*eventListener),
	}
	m := modules.NewModule("EVENT_MANAGER", 7)
	r.BaseServiceImpl = impl.NewBaseService(nil, m, r)

	return r
}

func (e *eventComponent) Fire(wrapper DefaultEventWrapper) {
	e.mutx.RLock()
	eventCell := e.eventCells[wrapper.NameSpace]
	e.mutx.RUnlock()

	if eventCell == nil {
		return
	}
	e.Logger.Info("开始处理event:" + wrapper.NameSpace.String())
	eventCell.FireEvent(wrapper.EventData)
}

func (e *eventComponent) RegisterListenerForEvent(listenerName string, eventNameSpace EventNameSpace, cb EventCallback) error {
	e.mutx.Lock()
	eventCell := e.eventCells[eventNameSpace]
	if eventCell == nil {
		eventCell = newEventCell()
		e.eventCells[eventNameSpace] = eventCell
	}
	listener := e.listeners[listenerName]
	if listener == nil {
		listener = newEventListener(listenerName)
		e.listeners[listenerName] = listener
	}
	e.mutx.Unlock()
	if err := listener.AddEvent(eventNameSpace); err != nil {
		return err
	}
	eventCell.AddListener(listenerName, cb)

	return nil
}

func (e *eventComponent) RemoveListenerForEvent(listenerID string, event EventNameSpace, ) {
	e.mutx.Lock()
	eventCell := e.eventCells[event]
	e.mutx.Unlock()

	if eventCell == nil {
		return
	}

	numListeners := eventCell.RemoveListener(listenerID)
	if numListeners == 0 {
		e.mutx.Lock()
		eventCell.mtx.Lock()
		if len(eventCell.listeners) == 0 {
			delete(e.eventCells, event)
		}
		eventCell.mtx.Unlock()
		e.mutx.Unlock()
	}
}

func (evsw *eventComponent) RemoveListener(listenerID string) {
	evsw.mutx.RLock()
	listener := evsw.listeners[listenerID]
	evsw.mutx.RUnlock()
	if listener == nil {
		return
	}

	evsw.mutx.Lock()
	delete(evsw.listeners, listenerID)
	evsw.mutx.Unlock()

	listener.SetRemoved()
	for _, event := range listener.GetEvents() {
		evsw.RemoveListenerForEvent(listenerID, event)
	}
}
