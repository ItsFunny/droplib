/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 2:18 下午
# @File : gossip_impl.go
# @Description :
# @Attention :
*/
package gossip

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/transport"
	"time"
)

var (
	_ IGossipComponent = (*gossipImpl)(nil)
)

type gossipImpl struct {
	*base.BaseComponent
	incTime uint64
	seqNum  uint64

	transport transport.ITransportComponent

	filter IFilter
}

func NewGossipComponent(ts transport.ITransportComponent, filters ...MessageFilter) *gossipImpl {
	r := &gossipImpl{
		incTime: uint64(time.Now().UnixNano()),
		seqNum:  uint64(0),
	}
	r.BaseComponent = base.NewBaseComponent(modules.NewModule("GOSSIP", 1), r)
	r.filter = wrapFilters(filters...)

	return r
}
