/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:08 上午
# @File : event.go
# @Description :
# @Attention :
*/
package types

type EventData interface{}
type EventCallback func(data EventData)
type EventNameSpace string
func (this EventNameSpace) String() string {
	return string(this)
}