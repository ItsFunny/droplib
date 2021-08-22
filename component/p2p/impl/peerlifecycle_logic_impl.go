/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/27 8:51 下午
# @File : logic_impl.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/debug"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/events"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	protos2 "github.com/hyperledger/fabric-droplib/protos"
	"github.com/hyperledger/fabric-droplib/protos/p2p/protos"
	"github.com/libp2p/go-libp2p-core/peer"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	_ services.IPeerLifeCycleService = (*PeerLifeCycleServiceImpl)(nil)
)

type PeerLifeCycleServiceImpl struct {
	*BaseLogicServiceImpl

	lifeCycleChan      *models2.Channel
	externHandleMember func(mems *protos.MemberInfoBroadCastMessageProto)

	notify    chan models.PeerRecord
	eventChan <-chan interface{}

	Addr        string
	localNodeId types.NodeID

	enablePingPong bool
}

func (p *PeerLifeCycleServiceImpl) ChoseInterestEvents(bus base.IP2PEventBusComponent) {
	go func() {
		p.eventChan = bus.AcquireEventChanByNameSpace(PEER_LIFECYCLE_EVENTNAMESPACE)
	}()
}

func NewPeerLifeCycleServiceImpl(
	enablePingPong bool,
	node services.INode,
	externHandleMember func(mems *protos.MemberInfoBroadCastMessageProto),
) *PeerLifeCycleServiceImpl {
	r := &PeerLifeCycleServiceImpl{}
	r.BaseLogicServiceImpl = NewBaseLogicServiceImpl(modules.P2P_MODULE_PING_PONG, r)
	// FIXME
	r.notify = make(chan models.PeerRecord, 20)
	r.localNodeId = node.ID()
	r.enablePingPong = enablePingPong
	r.Addr = node.GetExternAddress()
	r.externHandleMember = externHandleMember
	return r
}

func (p *PeerLifeCycleServiceImpl) Record() <-chan models.PeerRecord {
	return p.notify
}

func (p *PeerLifeCycleServiceImpl) OnStart() error {
	if p.enablePingPong {
		go p.broadcastOrP2P()
	}

	go p.receive()
	return nil
}

func (p *PeerLifeCycleServiceImpl) broadcastOrP2P() {
	broadCastTicker := time.NewTicker(time.Second * 60)

	go func() {
		time.Sleep(time.Second * 4)
		p.lifeCycleChan.Out <- models2.ChannelEnvelope{
			Broadcast: p2ptypes.TOPIC_LIFE_CYCLE,
			Message: &protos.PingPongBroadCastMessageProto{
				Info: p.Addr,
			},
			ChannelID: p2ptypes.CHANNEL_ID_LIFE_CYCLE,
		}
	}()

	for {
		if !p.Pre() {
			time.Sleep(time.Second * 3)
			continue
		}
		select {
		case <-p.Quit():
			return
		case <-broadCastTicker.C:
			p.lifeCycleChan.Out <- models2.ChannelEnvelope{
				From:      p.localNodeId,
				Broadcast: p2ptypes.TOPIC_LIFE_CYCLE,
				Message: &protos.PingPongBroadCastMessageProto{
					Info: p.Addr,
				},
				ChannelID: p2ptypes.CHANNEL_ID_LIFE_CYCLE,
			}
		}
	}
}

func (p *PeerLifeCycleServiceImpl) chatWithNext() {

}
func (p *PeerLifeCycleServiceImpl) receive() {
	directP2PStarted := uint32(0)
	for {
		if !p.Pre() {
			time.Sleep(time.Second * 3)
			continue
		}

		select {
		case <-p.Quit():
			return
		case r := <-p.eventChan:
			eve := r.(events.LifeCyclePeerEvent)
			switch eventData := eve.Event.(type) {
			case events.PingPongPeerEvent:
				p.handlePingPong(eventData, &directP2PStarted)
			case events.MemberInfoEvent:
				p.handleMemberInfo(eventData)
			}
		case v := <-p.lifeCycleChan.In:
			if v.ChannelID != p.lifeCycleChan.ID {
				p.Logger.Warn("收到了channel不匹配的msg,from:" + strconv.Itoa(int(v.ChannelID)) + ",self:" + strconv.Itoa(int(p.lifeCycleChan.ID)))
				continue
			}
			protoMsg, logicMsg, err := p.UnwrapFromBytes(p.lifeCycleChan.ID, v.Message)
			if nil != err {
				p.Logger.Error("拆包失败:" + err.Error())
				continue
			}
			p.Logger.Info("from:" + peer.ID(v.From).Pretty() + " 收到msg:" + protoMsg.String())

			p.handleMsg(v.From, protoMsg, logicMsg)
		}
	}
}

func (p *PeerLifeCycleServiceImpl) handleMemberInfo(eventData events.MemberInfoEvent) {
	p.Logger.Info("广播成员信息,当前已知的成员为:" + eventData.String())
	members := make([]*protos.MemberInfoNodeMessageProto, 0)
	for _, v := range eventData.Members {
		members = append(members, &protos.MemberInfoNodeMessageProto{
			ExternMultiAddress: v.ExternMultiAddress,
		})
	}
	p.lifeCycleChan.Out <- models2.ChannelEnvelope{
		Broadcast: p2ptypes.TOPIC_LIFE_CYCLE,
		Message: &protos.MemberInfoBroadCastMessageProto{
			Members: members,
		},
		ChannelID: p2ptypes.CHANNEL_ID_LIFE_CYCLE,
	}
}

func (p *PeerLifeCycleServiceImpl) handleMsg(from types.NodeID, msg proto.Message, logicMsg services2.IMessage) {
	switch m := msg.(type) {
	case *protos.MemberInfoBroadCastMessageProto:
		p.Logger.Info("收到了member信息:" + m.String())
		p.externHandleMember(m)
	case *protos.PingPongBroadCastMessageProto:
		p.notify <- models.PeerRecord{
			PeerId:             from,
			ExternMultiAddress: m.Info,
		}
	case *protos.PingMessageProto:
		p.Logger.Info("收到了来自于:" + from.Pretty() + ",的ping request")
	case *protos.PongMessageProto:
		p.Logger.Info("收到了来自与:" + from.Pretty() + ",的回应")
	}
}

func (p *PeerLifeCycleServiceImpl) ChoseInterestProtocolsOrTopics(pProtocol *models2.P2PProtocol) error {
	channels := pProtocol.Protocols[p2ptypes.PROTOCOL_PING_PONG]
	if channels == nil {
		return errors.New("ping 协议不存在")
	}
	p.lifeCycleChan = channels.GetChannelById(p2ptypes.CHANNEL_ID_LIFE_CYCLE)
	if p.lifeCycleChan == nil {
		return errors.New("topic 所ping channel不存在")
	}
	return nil
}

func (p *PeerLifeCycleServiceImpl) GetServiceId() types.ServiceID {
	return p2ptypes.LOGIC_PING_PONG
}

func (p *PeerLifeCycleServiceImpl) handlePingPong(eve events.PingPongPeerEvent, directP2PStarted *uint32) {
	if atomic.LoadUint32(directP2PStarted) == 1 {
		p.Logger.Info(debug.Green("当前有节点正在直连通信中"))
		return
	}

	p.Logger.Info(debug.Green("收到event,目标peer为:" + eve.PeerId.Pretty()))
	go func() {
		atomic.StoreUint32(directP2PStarted, 1)
		// 每10秒,或者完成
		streamFlag := protos2.StreamFlag_RESPONSE_AND_IN_TOUCH
		ticker := time.NewTicker(time.Second * 15)
		count := 0
		next := func() {
			atomic.StoreUint32(directP2PStarted, 0)
			p.notify <- models.PeerRecord{
				PeerId:             p.localNodeId,
				ExternMultiAddress: "",
			}
		}
		send := func(n bool) {
			p.lifeCycleChan.Out <- models2.ChannelEnvelope{
				From: types.NodeID(p.localNodeId),
				To:   types.NodeID(eve.PeerId),
				Message: &protos.PingMessageProto{
					Ping: "ping",
				},
				ChannelID:  p2ptypes.CHANNEL_ID_LIFE_CYCLE,
				StreamFlag: streamFlag,
			}
			if n {
				next()
			}
		}

		for {
			select {
			case <-ticker.C:
				if streamFlag&protos2.StreamFlag_RESPONSE < protos2.StreamFlag_RESPONSE {
					streamFlag = protos2.StreamFlag_RESPONSE_AND_CLOSE
					send(true)
				}
				return
			default:
				if count >= 10 {
					streamFlag = protos2.StreamFlag_RESPONSE_AND_CLOSE
					send(true)
					return
				} else {
					count++
					send(false)
					time.Sleep(time.Second * 2)
				}
			}
		}
	}()
}
