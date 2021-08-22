/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/1 5:24 下午
# @File : listener.go
# @Description :
# @Attention :
*/
package v2

import (
	"github.com/hyperledger/fabric-droplib/base/services"
)

type FireEvent interface {
	Fire(wrapper DefaultEventWrapper)
}

type IListenerComponentV2 interface {
	services.IBaseService
	FireEvent
	RegisterListenerForEvent(listenerName string, eventNameSpace EventNameSpace, cb EventCallback) error
	RemoveListenerForEvent( listenerID string,event EventNameSpace,)
	RemoveListener(listenerID string)
}

