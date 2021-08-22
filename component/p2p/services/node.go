/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/12 1:23 下午
# @File : node.go
# @Description :
# @Attention :
*/
package services

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type StreamFunc func(stream services.IStream)


type INode interface {
	fmt.Stringer
	SetStreamHandler(pid protocol.ID,service IProtocolService)
	GetProtocols() []models.ProtocolInfo
	AddProtocol(info models.ProtocolInfo)
	GetExternAddress() string
	// FIXME 不应该在这
	FromExternAddress(str string)types.NodeID
	ID() types.NodeID

	NewP2PEventBusComponent()base.IP2PEventBusComponent
	NewStreamRegistry()IStreamRegistry
	NewPubSub()IPubSubManager
}
