/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/13 11:36 上午
# @File : protocol.go
# @Description :
# @Attention :
*/
package bo

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/types"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	services2 "github.com/hyperledger/fabric-droplib/component/stream/services"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type LocalProtocolCreateBO struct {
	NodeID          models.NodeID
	ProtocolId      protocol.ID
	DataSupport     services2.StreamDataSupport
	ProtocolWrapper services2.IProtocolWrapper
	Channels        map[types.ChannelID]chan<- models2.Envelope
}

type RemoteProtocolCreateBO struct {
	NodeID types2.NodeID
	ProtocolId      protocol.ID
	Status func()bool
	Stream services.IStream
	DataSupport     services.StreamDataSupport
	MessageType services.IProtocolWrapper
}



type LocalBundleCreateBO struct {
	NodeID          types2.NodeID
	ProtocolId      protocol.ID
	DataSupport     services.StreamDataSupport
	ProtocolWrapper services.IProtocolWrapper
	Channels        map[types2.ChannelID]chan<- models2.Envelope
}


type RemoteBundleCreateBO struct {
	NodeID types2.NodeID
	ProtocolId      protocol.ID
	Status func()bool
	Stream services.IStream
	DataSupport     services.StreamDataSupport
	MessageType services.IProtocolWrapper
}