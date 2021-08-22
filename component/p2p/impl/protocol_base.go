/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 5:14 下午
# @File : stream.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/libp2p/go-libp2p-core/protocol"
	"strings"
)

type BaseProtocolServiceImpl struct {
	*impl.BaseServiceImpl
	PeerId     types.NodeID
	ProtocolId protocol.ID

	dataSupport services2.StreamDataSupport
	wrapper     services2.IProtocolWrapper
}

func NewBaseProtocolServiceImpl(
	peerId types.NodeID,
	protocolId protocol.ID,
	localOrRemote string,
	dataSupport services2.StreamDataSupport,
	dataWp services2.IProtocolWrapper,
) *BaseProtocolServiceImpl {
	r := &BaseProtocolServiceImpl{
		PeerId:      peerId,
		ProtocolId:  protocolId,
		dataSupport: dataSupport,
		wrapper:     dataWp,
	}
	localOrRemote = strings.ToUpper(localOrRemote)
	r.BaseServiceImpl = impl.NewBaseService(nil, modules.NewModule(localOrRemote+"_"+string(protocolId), 1), r)
	return r
}
func (b *BaseProtocolServiceImpl) GetPeerId() types.NodeID {
	return b.PeerId
}
