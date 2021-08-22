/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 4:51 下午
# @File : wrapper.go
# @Description :
# @Attention :
*/
package wrapper

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)



type CallBack func()

type ProtocolWrapper struct {
	PeerId peer.ID
	// p2p 中的协议id
	ProtoCol protocol.ID
	// 是否执行完毕立马close
	CloseQuickly bool
	CloseIndex   int

	// 用于当执行完stream,进行回调
	CallBack CallBack

	Handler network.StreamHandler

	QueueSize int
}
