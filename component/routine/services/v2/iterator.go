/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 4:50 下午
# @File : iterator.go
# @Description :
# @Attention :
*/
package v2

type TaskIterator interface {
	HasNext() bool
	Next() ITask
}
