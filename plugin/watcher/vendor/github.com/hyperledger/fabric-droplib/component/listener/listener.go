/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/20 8:14 上午
# @File : listener.go
# @Description :
# @Attention :
*/
package listener

import (
	"github.com/hyperledger/fabric-droplib/component/base"
)

type IListenerComponent interface {
	base.IComponent
	RegisterListener(topic ...string) <-chan interface{}
	NotifyListener(data interface{}, listenerIds ...string)
}
