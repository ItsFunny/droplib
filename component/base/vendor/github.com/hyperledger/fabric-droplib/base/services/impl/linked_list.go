/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 9:04 下午
# @File : linked_list.go
# @Description :
# @Attention :
*/
package impl

import "github.com/hyperledger/fabric-droplib/base/services"

type BaseLinkedList struct {
	Next services.ILinkedList
}

func NewBaseLinkedList() *BaseLinkedList {
	r := &BaseLinkedList{}
	return r
}

func (b *BaseLinkedList) GetNext() services.ILinkedList {
	return b.Next
}

func (b *BaseLinkedList) SetNext(linkInterface services.ILinkedList) {
	b.Next = linkInterface
}
