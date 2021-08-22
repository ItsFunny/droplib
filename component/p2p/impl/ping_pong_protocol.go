/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/30 5:45 下午
# @File : ping_pong_protocol.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"github.com/golang/protobuf/proto"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/p2p/impl/protocol"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	protos2 "github.com/hyperledger/fabric-droplib/protos"
	"github.com/hyperledger/fabric-droplib/protos/p2p/protos"
	"math/rand"
	"time"
)

var (
	_ services.ILocalNetworkBundle  = (*LocalPingProtocolNetworkBundle)(nil)
	_ services.IRemoteNetworkBundle = (*RemotePingPongProtocolNetworkBundle)(nil)
)

type LocalPingProtocolNetworkBundle struct {
	logger v2.Logger
	*protocol.LocalCommonNetworkBundle
}

func NewLocalPingProtocolNetworkBundle() *LocalPingProtocolNetworkBundle {
	responses := make(map[types.ChannelID]protocol.Response)
	responses[p2ptypes.CHANNEL_ID_LIFE_CYCLE] = func() ([]byte, error) {
		resp := &protos.PongMessageProto{
			Pong: "pong",
		}
		return proto.Marshal(resp)
	}
	r := &LocalPingProtocolNetworkBundle{
		logger: logrusplugin.NewLogrusLogger(modules.NewModule("LOCAL_PING_PONG_BUNDLE", 1)),
	}
	r.LocalCommonNetworkBundle = protocol.NewLocalCommonNetworkBundle(responses, r)
	return r
}

type RemotePingPongProtocolNetworkBundle struct {
	*protocol.RemoteCommonNetworkBundle
	sendCount int
}

func NewRemotePingPongProtocolNetworkBundle(peerId types.NodeID) *RemotePingPongProtocolNetworkBundle {
	r := &RemotePingPongProtocolNetworkBundle{}
	lg := logrusplugin.NewLogrusLogger(modules.NewModule("REMOTE_PING_PONG_BUNDLE", 1))
	ver := make(map[types.ChannelID]protocol.VerifyResponse)
	ver[p2ptypes.CHANNEL_ID_LIFE_CYCLE] = func(bytes []byte) (interface{}, error) {
		pong := new(protos.PongMessageProto)
		if err := proto.Unmarshal(bytes, pong); nil != err {
			return nil, err
		}
		return pong, nil
	}
	r.RemoteCommonNetworkBundle = protocol.NewRemoteCommonNetworkBundle(ver, lg, peerId, p2ptypes.PROTOCOL_PING_PONG, r)

	return r
}

func (r *RemotePingPongProtocolNetworkBundle) StreamCondition(service services.IRemoteProtocolService) protos2.StreamFlag {
	if r.sendCount > 10 {
		r.sendCount = 0
		return protos2.StreamFlag_RESPONSE_AND_CLOSE
	}
	return protos2.StreamFlag_RESPONSE_AND_IN_TOUCH
}

func (r *RemotePingPongProtocolNetworkBundle) AfterSendStream() {
	r.sendCount++
}

// FIXME
func (r *RemotePingPongProtocolNetworkBundle) WaitBy() {
	se := time.Second * time.Duration(rand.Intn(3))
	time.Sleep(se)
}
