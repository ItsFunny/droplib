/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 2:19 下午
# @File : core.go
# @Description :
# @Attention :
*/
package gossip

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
)

type MessageFilter func(msg GossipMessag) bool
type IFilter interface {
	Filter(msg GossipMessag) bool
}
type filter struct {
	f MessageFilter
	*impl.BaseLinkedList
}

func wrapFilters(filters ...MessageFilter) *FilterChain {
	var f *filter
	for _, v := range filters {
		f = services.LinkLast(f, &filter{f: v}).(*filter)
	}
	return &FilterChain{
		f,
	}
}

type FilterChain struct {
	filter *filter
}

func (f *FilterChain) Filter(msg GossipMessag) bool {
	var r bool
	err := services.Iterator(f.filter, func(node services.ILinkedList) error {
		r = node.(*filter).f(msg)
		if !r {
			return errors.New("123")
		}
		return nil
	})
	return err != nil
}

type byteHandler struct {
	h func(data []byte) error
	*impl.BaseLinkedList
}
type ByteHandlerChain struct {
	byteHandler *byteHandler
}

func (b *ByteHandlerChain) Handle(data []byte) error {
	return b.byteHandler.h(data)
}
