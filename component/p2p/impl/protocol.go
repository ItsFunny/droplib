/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 7:33 下午
# @File : protocol.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/debug"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/common/errors"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/libp2p/go-libp2p-core/protocol"
	"time"
)

var (
	_ services.IRemoteProtocolService = (*DefaultRemoteProtocolServiceImpl)(nil)
)

// FIXME 名称修改,添加default
type DefaultRemoteProtocolServiceImpl struct {
	*BaseProtocolServiceImpl

	peerAlive  func() bool
	streamInfo services2.IStream

	networkBundle services.IRemoteNetworkBundle
}

func NewDefaultRemoteProtocolServiceImpl(peerId types.NodeID,
	protocolId protocol.ID,
	peerStatus func() bool,
	streamInfo services2.IStream,
	dataSupport services2.StreamDataSupport,
	dataWp services2.IProtocolWrapper,
	helper services.IRemoteNetworkBundle,
) *DefaultRemoteProtocolServiceImpl {
	base := NewBaseProtocolServiceImpl(peerId, protocolId, "remote", dataSupport, dataWp)
	r := &DefaultRemoteProtocolServiceImpl{
		BaseProtocolServiceImpl: base,
		peerAlive:               peerStatus,
		streamInfo:              streamInfo,
		networkBundle:           helper,
	}
	return r
}

func (b *DefaultRemoteProtocolServiceImpl) GetNetworkBundle() services.IRemoteNetworkBundle {
	return b.networkBundle
}

func (b *DefaultRemoteProtocolServiceImpl) SendWithStream(registry services.IStreamRegistry, c <-chan models2.Envelope) {
	stream := b.streamInfo
	cache := false
	for {
		if !b.IsRunning() {
			if nil != stream {
				stream.Reset()
			}
			stream = nil
			return
		}

		// FIXME ,可能只是网络不通顺而已,sleep better than dead
		if !b.peerAlive() {
			b.Stop()
			stream.Reset()
			stream = nil
			return
		}

		if stream == nil {
			v, err := b.networkBundle.StreamPolicy(registry, stream, cache)
			if nil != err {
				// FIXME  dynamic
				time.Sleep(time.Second * 2)
			}
			stream = v
			b.streamInfo = stream
		}

		select {
		case envelope := <-c:
			if envelope.Message == nil {
				b.Logger.Error(fmt.Sprintf("dropping nil message,peer:%s", b.PeerId))
				continue
			}

			streamFlag := envelope.StreamFlag
			cache = envelope.AcquireStreamFromCache
			data, err := b.wrapper.Encode(envelope.Message, func(s services2.IStreamObject) {
				s.SetChannelID(envelope.ChannelID)
			}, func(s services2.IStreamObject) {
				s.SetStreamFlagInfo(int32(streamFlag))
			}).ToBytes()
			if nil != err {
				b.Logger.Error("序列化失败:" + err.Error())
				continue
			}

			bytes, err := b.dataSupport.Pack(data)
			if nil != err {
				b.Logger.Error("pack失败,peerId:%s,err:%s", b.peerAlive, err.Error())
				continue
			}

			if err := b.networkBundle.SendWithStream(stream, bytes); nil != err {
				b.Logger.Error("发送失败,peerId:%s,err:%s", b.peerAlive, err.Error())
				stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
				continue
			}

			b.networkBundle.AfterSendStream()

			if err := b.deferSendStream(envelope.ChannelID, stream, streamFlag); nil != err {
				b.Logger.Error("等待响应失败:" + err.Error())
				// FIXME
				switch err.(type) {
				case *errors.NetworkError:
					// stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
				case *errors.LogicError:
				}
				stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
				continue
			}
			stream = registry.RecycleStream(stream, streamFlag)

			b.networkBundle.WaitBy()
		}
	}
}

func (b *DefaultRemoteProtocolServiceImpl) deferSendStream(cid types.ChannelID, info services2.IStream, flag protos.StreamFlag) error {
	if flag&protos.StreamFlag_RESPONSE > 0 {
		start := time.Now()
		if err := b.networkBundle.WaitForResponse(cid, info, b.dataSupport); nil != err {
			return err
		}
		after := time.Since(start).Seconds()
		b.Logger.Info(debug.Red(fmt.Sprintf("等待接收耗时:%f", after)))
	}
	return nil
}

func (b *DefaultRemoteProtocolServiceImpl) OnStop() {
	if nil != b.streamInfo {
		b.streamInfo.Reset()
	}
}

// func (b *DefaultRemoteProtocolServiceImpl) SendWithStream(registry services.IStreamRegistry, c <-chan models2.Envelope) {
// 	stream := b.streamInfo
// 	cache := true
// 	for {
// 		if !b.IsRunning() {
// 			if nil != stream {
// 				stream.Reset()
// 			}
// 			stream = nil
// 			return
// 		}
//
// 		// FIXME ,这里抽象到helper中
// 		if stream == nil {
// 			v, err := b.networkBundle.StreamPolicy(stream)
// 			if nil != err {
// 				// FIXME  dynamic
// 				time.Sleep(time.Second * 2)
// 			}
// 			stream = v
// 			b.streamInfo = stream
// 			continue
// 		}
//
// 		if stream == nil {
// 			v, err := registry.AcquireNewStream(b.PeerId, b.ProtocolId, cache)
// 			if nil != err {
// 				// FIXME  dynamic
// 				time.Sleep(time.Second * 2)
// 			}
// 			stream = v
// 			b.streamInfo = stream
// 			continue
// 		}
//
// 		// FIXME ,可能只是网络不通顺而已,sleep better than dead
// 		if !b.peerAlive() {
// 			b.Stop()
// 			stream.Reset()
// 			stream = nil
// 			return
// 		}
//
// 		if !stream.IsHealthy() {
// 			stream.Reset()
// 			stream = nil
// 			continue
// 		}
//
// 		select {
// 		case envelope := <-c:
// 			if envelope.Message == nil {
// 				b.Logger.Error(fmt.Sprintf("dropping nil message,peer:%s", b.PeerId))
// 				continue
// 			}
//
// 			streamFlag := envelope.StreamFlag
// 			cache = envelope.AcquireStreamFromCache
// 			data, err := b.wrapper.Encode(envelope.Message, func(s services2.IStreamObject) {
// 				s.SetChannelID(envelope.ChannelID)
// 			}, func(s services2.IStreamObject) {
// 				s.SetStreamFlagInfo(int32(streamFlag))
// 			}).ToBytes()
// 			if nil != err {
// 				b.Logger.Error("序列化失败:" + err.Error())
// 				continue
// 			}
//
// 			bytes, err := b.dataSupport.Pack(data)
// 			if nil != err {
// 				b.Logger.Error("pack失败,peerId:%s,err:%s", b.peerAlive, err.Error())
// 				continue
// 			}
//
// 			if err := b.networkBundle.SendWithStream(stream, bytes); nil != err {
// 				b.Logger.Error("发送失败,peerId:%s,err:%s", b.peerAlive, err.Error())
// 				continue
// 			}
//
// 			rw := bufio.NewWriter(stream)
// 			if _, err = rw.Write(bytes); nil != err {
// 				b.Logger.Error("发送失败,peerId:%s,err:%s", b.peerAlive, err.Error())
// 				stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
// 				continue
// 			}
// 			if err = rw.Flush(); nil != err {
// 				b.Logger.Error("flush失败,peerId:%s,err:%s", b.peerAlive, err.Error())
// 				stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
// 				continue
// 			}
//
// 			b.helper.AfterSendStream()
//
// 			if err := b.deferSendStream(envelope.ChannelID, stream, streamFlag); nil != err {
// 				b.Logger.Error("等待响应失败:" + err.Error())
// 				// FIXME
// 				switch err.(type) {
// 				case *errors.NetworkError:
// 					// stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
// 				case *errors.LogicError:
// 				}
// 				stream = registry.RecycleStream(stream, protos.StreamFlag_CLOSE_IMMEDIATELY)
// 				continue
// 			}
// 			stream = registry.RecycleStream(stream, streamFlag)
//
// 			b.helper.WaitBy()
// 		}
// 	}
// }
