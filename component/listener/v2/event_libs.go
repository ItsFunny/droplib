/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:09 上午
# @File : event_libs.go
# @Description :
# @Attention :
*/
package v2

import (
	"errors"
	"sync"
)

const (
	AVAILABLE = 0
	ACQUIRED  = 1
)



type eventCell struct {
	mtx       sync.RWMutex
	listeners map[string]EventCallback
}
type eventListener struct {
	id string

	mtx     sync.RWMutex
	removed bool
	events  []EventNameSpace
}

func newEventCell() *eventCell {
	return &eventCell{
		listeners: make(map[string]EventCallback),
	}
}

func newEventListener(id string) *eventListener {
	return &eventListener{
		id:      id,
		removed: false,
		events:  nil,
	}
}

func (evl *eventListener) AddEvent(event EventNameSpace) error {
	evl.mtx.Lock()

	if evl.removed {
		evl.mtx.Unlock()
		return errors.New("该listenerId已经remove" + evl.id)
	}

	evl.events = append(evl.events, event)
	evl.mtx.Unlock()
	return nil
}

func (cell *eventCell) AddListener(listenerID string, cb EventCallback) {
	cell.mtx.Lock()
	cell.listeners[listenerID] = cb
	cell.mtx.Unlock()
}
func (cell *eventCell) RemoveListener(listenerID string) int {
	cell.mtx.Lock()
	delete(cell.listeners, listenerID)
	numListeners := len(cell.listeners)
	cell.mtx.Unlock()
	return numListeners
}

func (evl *eventListener) SetRemoved() {
	evl.mtx.Lock()
	evl.removed = true
	evl.mtx.Unlock()
}

func (evl *eventListener) GetEvents() []EventNameSpace {
	evl.mtx.RLock()
	events := make([]EventNameSpace, len(evl.events))
	copy(events, evl.events)
	evl.mtx.RUnlock()
	return events
}

func (cell *eventCell) FireEvent(data EventData) {
	eventCallbacks := make([]EventCallback, 0, len(cell.listeners))
	cell.mtx.RLock()
	for _, cb := range cell.listeners {
		eventCallbacks = append(eventCallbacks, cb)
	}
	cell.mtx.RUnlock()

	for _, cb := range eventCallbacks {
		cb(data)
	}
}
