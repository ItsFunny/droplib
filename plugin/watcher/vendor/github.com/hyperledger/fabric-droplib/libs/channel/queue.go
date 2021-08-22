/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/14 6:22 下午
# @File : queue.go
# @Description :
# @Attention :
*/
package channel

type IData interface {
	ID() interface{}
}
type IChan interface {
	Take() IData
	Push(task IData) (int, error)
}


