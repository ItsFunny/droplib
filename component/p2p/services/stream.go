/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 4:49 下午
# @File : stream.go
# @Description :
# @Attention :
*/
package services

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/wrapper"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type P2PStreamHandlerWrapper struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	Host    host.Host
	Stream  network.Stream
	Handler network.StreamHandler

	PeerId    peer.ID
	ProtoCol  protocol.ID
	FifoQueue impl.Queue
}
type streamHandlerFactory interface {
	createP2PStreamHandler(h host.Host, wrapper wrapper.ProtocolWrapper) (*P2PStreamHandlerWrapper, error)
}

type Sweeper interface {
	Sweep()
}

type ITimeWheel interface {
}

type IStreamRegistry interface {
	services.IBaseService
	Dial(wp models.DialWrapper) (map[protocol.ID]services.IStream, error)
	AcquireNewStream(peerId types2.NodeID, protocolId protocol.ID, useCache bool) (services.IStream, error)
	RecycleStream(stream services.IStream, flag protos.StreamFlag) services.IStream
}
