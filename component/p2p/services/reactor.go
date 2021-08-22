/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 4:06 下午
# @File : reactor.go
# @Description :
# @Attention :
*/
package services

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type ILogicReactor interface {
	services.IReactor
	GetBoundServices() []base.ILogicService
}

type IPeerLifeCycleReactor interface {
	ILogicReactor
	AddPeer(info models.NewPeerInfo)
	PeerStatus(pid types2.NodeID) *models.PeerLifeCycle
	MemberChanged() <-chan models.MemberChanged
	OutMemberChanged(m models.MemberChanged)
}

type IProtocolLifeCycleReactor interface {
	ILogicReactor
	RegisterRemoteProtocol(remotePeerId types2.NodeID, registry IStreamRegistry, info models.RemoteProtocolRegistration) (IRemoteProtocolService, error)
	SetupSelfProtocol(node INode, info models.SelfProtocolRegistration) (IProtocolService, error)
	GetProtocolByPeerIdAndProtocolId(peerId types2.NodeID, protocolId protocol.ID) IRemoteProtocolService
	AppendProtocolProvider(id protocol.ID, proFactory ProtocolFactory)
}
