/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 2:13 下午
# @File : gossip.go
# @Description :
# @Attention :
*/
package gossip

import "github.com/hyperledger/fabric-droplib/component/base"

type GossipMessag interface {
	GetIncNumber() uint64
	GetSeqNumber() uint64
}
type IGossipComponent interface {
	base.IComponent
}
