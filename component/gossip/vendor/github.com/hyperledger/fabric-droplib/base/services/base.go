/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/1 4:06 下午
# @File : base.go
# @Description :
# @Attention :
*/
package services

import (
	"fmt"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	"github.com/hyperledger/fabric-droplib/base/services/e"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"reflect"
)


// 基础services,自定义一些基本功能
type IBaseService interface {
	// Start the service.
	// If it's already started or stopped, will return an error.
	// If OnStart() returns an error, it's returned by Start()

	BStart(ctx ...models.StartOption) error
	OnStart(ctx *models.StartCTX) error

	// Stop the service.
	// If it's already stopped, will return an error.
	// OnStop must never error.
	BStop(ctx ...models.StopOption) error
	OnStop(ctx *models.StopCTX)

	BReady(ctx ...models.ReadyOption) error
	OnReady(ctx *models.ReadyCTX) error

	// Reset the service.
	// Panics by default - must be overwritten to enable reset.
	Reset() error
	OnReset() error

	// Return true if the service is running
	IsRunning() bool

	// Quit returns a channel, which is closed once service is stopped.
	Quit() <-chan struct{}

	// String representation of the service
	String() string

	// SetLogger sets a logger.
	SetLogger(logger v2.Logger)
}

type Type int
type BaseType Type

type Template interface {
	ReactorTemplate
	ValidIsMine(typeValue BaseType, excepted BaseType) bool
}

type ReactorTemplate interface {
	fmt.Stringer
	LinkLast(template Template)
	GetNext() Template
	SetNext(template Template)
}

type IBaseValidate interface {
	ValidateBasic() error
}

type ListHook func(node ILinkedList) error
type ILinkedList interface {
	GetNext() ILinkedList
	SetNext(linkInterface ILinkedList)
}

// FIXME  dynamic
func LinkLast(firer, new ILinkedList) ILinkedList {
	if IsNil(firer) {
		return new
	}

	temp := firer
	for !IsNil(temp.GetNext()) {
		temp = temp.GetNext()
	}
	temp.SetNext(new)
	return firer
}
func Iterator(list ILinkedList, hook ListHook) error {
	for tmp := list; !IsNil(tmp); tmp = tmp.GetNext() {
		if err := hook(tmp); nil != err {
			if err == e.STORE_ITERATOR_ERROR {
				break
			}
			return err
		}
	}
	return nil
}
func IsNil(firer interface{}) bool {
	return firer == nil || (reflect.ValueOf(firer).Kind() == reflect.Ptr && reflect.ValueOf(firer).IsNil())
}

type IMessage interface {
	ValidateBasic() error
}
