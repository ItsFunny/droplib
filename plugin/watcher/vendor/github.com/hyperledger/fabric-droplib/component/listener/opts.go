/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/20 8:33 上午
# @File : opts.go
# @Description :
# @Attention :
*/
package listener

type Opt func(component *listenerComponent)

func ClientId(cid string) Opt {
	return func(component *listenerComponent) {
		component.clientId = cid
	}
}
