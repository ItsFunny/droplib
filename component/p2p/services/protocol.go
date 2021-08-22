/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 5:06 下午
# @File : protocol.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/bo"
	"github.com/libp2p/go-libp2p-core/network"
)

type IProtocolServiceSupport interface {
	IRemoteProtocolService
	IProtocolService
}

type IBaseProtocolService interface {
	services.IBaseService
}

type IRemoteProtocolService interface {
	IBaseProtocolService
	SendWithStream(registry IStreamRegistry, c <-chan models2.Envelope)
	GetPeerId() types.NodeID
	GetNetworkBundle() IRemoteNetworkBundle
}

type IRemoteNetworkBundle interface {
	SendWithStream(stream services.IStream, data []byte) error
	StreamPolicy(registry IStreamRegistry, stream services.IStream, cache bool) (services.IStream, error)
	StreamHealth(s services.IStream) bool
	//
	AfterSendStream()
	WaitBy()

	WaitForResponse(cid types.ChannelID, stream services.IStream, support services.StreamDataSupport) error
	// 用于记录
	OnAfterWaitResponse(cid types.ChannelID, responseData interface{}) error
}

type IProtocolService interface {
	IBaseProtocolService
	ReceiveStream(stream services.IStream, )
	GetLocalProtocolNetworkBundle() ILocalNetworkBundle
}

type ILibP2PProtocolService interface {
	IProtocolService
	ReceiveLibP2PStream(stream network.Stream)
}

type ILocalNetworkBundle interface {
	Response(cid types.ChannelID, stream services.IStream, support services.StreamDataSupport) error
	AfterReceive()
}

type ProtocolFactory interface {
	Local(req bo.LocalProtocolCreateBO) (IProtocolService, error)
	BundleLocal(req bo.LocalBundleCreateBO) (ILocalNetworkBundle, error)

	Remote(req bo.RemoteProtocolCreateBO) (IRemoteProtocolService, error)
	BundleRemote(req bo.RemoteBundleCreateBO) (IRemoteNetworkBundle, error)
}


// type IProtocolLifeCycleManager interface {
// 	services.IBaseService
// 	ILogicManager
// 	RegisterRemoteProtocol(remotePeerId peer.ID, registry IStreamRegistry, info models.RemoteProtocolRegistration) (IRemoteProtocolService, error)
// 	SetupSelfProtocol(node *models.Node, info models.SelfProtocolRegistration) (IProtocolService, error)
// 	GetProtocolByPeerIdAndProtocolId(peerId peer.ID, protocolId protocol.ID) IRemoteProtocolService
// 	AppendProtocolProvider(id protocol.ID, provider *ProtocolProvider, helperProvider *ProtocolServiceHelperProvider)
// }
