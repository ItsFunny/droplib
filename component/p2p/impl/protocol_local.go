/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/13 10:50 上午
# @File : protocol_local.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/debug"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"strconv"
)

var (
	_ services.ILibP2PProtocolService = (*DefaultLocalProtocolServiceImpl)(nil)
)

type DefaultLocalProtocolServiceImpl struct {
	inChannelQueues map[types.ChannelID]chan<- models2.Envelope

	*BaseProtocolServiceImpl

	helper services.ILocalNetworkBundle
}

func NewDefaultLocalProtocolServiceImpl(peerId types.NodeID,
	protocolId protocol.ID,
	dataSupport services2.StreamDataSupport,
	dataWp services2.IProtocolWrapper,
	ins map[types.ChannelID]chan<- models2.Envelope,
	helper services.ILocalNetworkBundle,
) *DefaultLocalProtocolServiceImpl {
	base := NewBaseProtocolServiceImpl(peerId, protocolId, "local", dataSupport, dataWp)
	r := &DefaultLocalProtocolServiceImpl{
		inChannelQueues:         ins,
		BaseProtocolServiceImpl: base,
		helper:                  helper,
	}
	return r
}

func (b *DefaultLocalProtocolServiceImpl) GetLocalProtocolNetworkBundle() services.ILocalNetworkBundle {
	return b.helper
}

func (b *DefaultLocalProtocolServiceImpl) GetLocalProtocolServiceHelper() services.ILocalNetworkBundle {
	return b.helper
}

func (b *DefaultLocalProtocolServiceImpl) ReceiveStream(stream services2.IStream) {
	b.ReceiveLibP2PStream(stream.(network.Stream))
}

func (b *DefaultLocalProtocolServiceImpl) ReceiveLibP2PStream(stream network.Stream) {
	// FIXME  waterLimit
	b.Logger.Info("收到了新的stream,remote:" + stream.Conn().RemotePeer().Pretty())
	for {
		if !b.IsRunning() {
			b.Logger.Info(b.PeerId.Pretty() + ", reset stream")
			stream.Reset()
			return
		}
		pack, err := b.dataSupport.UnPack(context.TODO(), stream, nil)
		if nil != err {
			b.Logger.Info(b.PeerId.Pretty() + ", reset stream")
			stream.Reset()
			return
		}
		wrap, err := b.wrapper.Decode(pack)
		if nil != err {
			b.Logger.Info(b.PeerId.Pretty() + ", reset stream")
			stream.Reset()
			b.Logger.Error("protobuf反序列化失败:" + err.Error())
			return
		}
		b.Logger.Info("收到节点:" + stream.Conn().RemotePeer().Pretty() + "的信息,指向的channel为:" +
			strconv.Itoa(int(wrap.GetChanelID())) +
			"" + ",数据为:" + string(wrap.GetData()))
		cid := wrap.GetChanelID()
		if q, exist := b.inChannelQueues[cid]; !exist {
			b.Logger.Error("对应的channel不存在,cid为:" + strconv.Itoa(int(cid)) + ",本机节点为:" + b.PeerId.Pretty() + ", reset stream")
			stream.Reset()
			return
		} else {
			q <- models2.Envelope{
				From:      types.NodeID(stream.Conn().RemotePeer()),
				Message:   wrap.GetData(),
				ChannelID: wrap.GetChanelID(),
			}
		}
		b.helper.AfterReceive()

		r := protos.StreamFlag(wrap.GetStreamFlagInfo())
		b.Logger.Info(debug.Red("远端节点的streamFlag为:" + strconv.Itoa(int(wrap.GetStreamFlagInfo()))))
		if r&protos.StreamFlag_RESPONSE >= protos.StreamFlag_RESPONSE {
			b.Logger.Info("远端节点需要接收响应:streamFlag为:" + strconv.Itoa(int(wrap.GetStreamFlagInfo())))
			if err := b.helper.Response(cid, stream, b.dataSupport); nil != err {
				// 防止消息重放:
				b.Logger.Info(b.PeerId.Pretty() + ", response 发生错误,reset")
				stream.Reset()
				return
			}
		}
		if r&protos.StreamFlag_CLOSE_IMMEDIATELY >= protos.StreamFlag_CLOSE_IMMEDIATELY {
			b.Logger.Info(b.PeerId.Pretty() + ", 发送方申请断开连接")
			return
		}
	}
}
