/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/27 8:53 下午
# @File : p2p.go
# @Description :
# @Attention :
*/
package p2p

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/p2p/bo"
	streamimpl "github.com/hyperledger/fabric-droplib/component/p2p/impl"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	protos2 "github.com/hyperledger/fabric-droplib/protos"
	"github.com/hyperledger/fabric-droplib/protos/p2p/protos"
	"github.com/libp2p/go-libp2p"
	p2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"sync/atomic"
)

type P2P struct {
	sync.RWMutex
	*impl.BaseServiceImpl
	router   *streamimpl.Router
	services map[types.ServiceID]base.ILogicService

	protocols   *models2.P2PProtocol
	cfg         *config.P2PConfiguration
	nodeFactory services.INodeFactory

	status uint32
}

func init() {
	common.RegisterBlackList("runtime/asm_amd64")
}

// default lib p2p
func NewP2P(cfg *config.P2PConfiguration, opts ...libp2p.Option) *P2P {
	ctx := context.Background()
	r := &P2P{
		cfg:         cfg,
		nodeFactory: streamimpl.NewDefaultNodeFactory(),
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_P2P, r)
	//  FIXME 添加ctx
	node := r.nodeFactory.CreateNode(ctx, r.cfg, opts...)
	r.Logger.Info("创建的节点信息为:" + node.String())
	bus := node.NewP2PEventBusComponent()

	// registry
	registry := node.NewStreamRegistry()

	// lifecycle
	support := streamimpl.NewPackerUnPacker()
	lifecycle := streamimpl.NewProtocolLifeCycleManagerImpl(support)
	// peer lifecycle
	peerLifeCycle := streamimpl.NewPeerLifeCycleManagerImpl(cfg, bus, node)

	// pubsub
	pubsubI := node.NewPubSub()

	r.router = streamimpl.NewRouter(node, bus, registry, lifecycle, peerLifeCycle, pubsubI)

	r.services = make(map[types.ServiceID]base.ILogicService)
	r.AppendDefaultPingPong()
	return r
}

func (p *P2P) GetNode() services.INode {
	return p.router.GetNode()
}

// should be called when the system is online immediately
func (p *P2P) GetRegisteredService(id types.ServiceID) base.ILogicService {
	p.WaitUntilReady()

	if s, exist := p.services[id]; !exist {
		panic("service不存在,id为:" + string(id))
	} else {
		return s
	}
}
func (p *P2P) GetRouter() *streamimpl.Router {
	return p.router
}
func (p *P2P) GetEventBusComponent() base.IP2PEventBusComponent {
	return p.router.GetEventBusComponent()
}
func (n *P2P) GetMultiaddr() multiaddr.Multiaddr {
	node := n.router.GetNode()
	addr := n.cfg.IdentityProperty.Address
	port := n.cfg.IdentityProperty.Port
	dest := fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", addr, port, node.ID().Pretty())
	newMultiaddr, _ := multiaddr.NewMultiaddr(dest)
	return newMultiaddr
}
func (n *P2P) GetExternalAddress() string {
	return n.GetRouter().GetNode().GetExternAddress()
}
func (n *P2P) GetNodeID() types.NodeID {
	return n.GetRouter().GetNode().ID()
}

func (this *P2P) OnReady() error {
	return nil
}
func (this *P2P) Status() uint32 {
	return atomic.LoadUint32(&this.status)
}

type ProtocolProviderRegister struct {
	ProtocolId      p2pProtocol.ID
	MessageType     services2.IProtocolWrapper
	ProtocolFactory services.ProtocolFactory
	Channels        []models.ChannelInfo
	// FIXME , add logic here ?
}

func (this *P2P) AppendProtocol(protoc ProtocolProviderRegister) {
	this.Lock()
	defer this.Unlock()
	node := this.router.GetNode()
	for _, v := range node.GetProtocols() {
		if v.ProtocolId == protoc.ProtocolId {
			this.Logger.Warn("重复注册protocol:" + string(protoc.ProtocolId))
			continue
		}
	}
	info := models.ProtocolInfo{
		ProtocolId:  protoc.ProtocolId,
		MessageType: protoc.MessageType,
		Channels:    protoc.Channels,
	}
	node.AddProtocol(info)
	this.router.GetProtocolLifeCycleManager().AppendProtocolProvider(protoc.ProtocolId, protoc.ProtocolFactory)
}

// type ProtocolProperty struct {
// 	Protocols []*ProtocolPropertyNode
// }
//
// func (this *ProtocolProperty) GetMessageTypeByChannelId(cid p2ptypes.ChannelID) libs.IChannelMessageWrapper {
// 	for _, v := range this.Protocols {
// 		for _, cc := range v.Channels {
// 			if cc.ChannelId == cid {
// 				return cc.MessageType
// 			}
// 		}
// 	}
// 	panic("找不到匹配的messageType")
// }
//
// type ProtocolPropertyNode struct {
// 	ProtocolId   protocol.ID
// 	StreamFlag   int
// 	ProtocolType libs.IProtocolWrapper
// 	Channels     []ProtocolBindChannel
// }
// type ProtocolBindChannel struct {
// 	ChannelId p2ptypes.ChannelID
// 	MessageType libs.IChannelMessageWrapper
// 	ChannelSize int
// 	Topics []ChannelBindTopic
// }
// type ChannelBindTopic struct {
// 	Topic string
// }

func (this *P2P) OnStart() error {
	this.router.Start(services2.ASYNC_START)

	notify := make(chan struct{})
	this.protocols = this.router.Setup(notify)
	// FIXME
	<-notify

	boundServices := this.router.GetBoundServices()
	for _, logicService := range boundServices {
		if nil == logicService {
			continue
		}
		this.registerService(logicService)
	}
	return this.Ready(services2.READY_UNTIL_START)
}

func (this *P2P) RegisterService(s base.ILogicService) {
	this.WaitUntilReady()
	s.Start(services2.ASYNC_START_WAIT_READY)
	this.registerService(s)
}

func (this *P2P) appendPingpong() {

}

func (this *P2P) registerService(s base.ILogicService) {
	this.Lock()
	defer this.Unlock()

	if _, exist := this.services[s.GetServiceId()]; exist {
		this.Logger.Warn("重复注册service:" + string(s.GetServiceId()))
		return
	}

	if err := s.ChoseInterestProtocolsOrTopics(this.protocols); nil != err {
		panic("设定service 协议失败:" + err.Error())
	}
	s.SetChannelMessagesWrapper(this.protocols.GetMessageWrappers())
	s.ChoseInterestEvents(this.router.GetEventBusComponent())

	if err := s.Ready(services2.READY_UNTIL_START); nil != err {
		this.Logger.Error("启动logicService失败,logicService:", s, "error:", err.Error())
	}
}
func (p *P2P) AppendDefaultPingPong() {
	property := pingPongProtocolProtocolProperty()
	p.AppendProtocol(property)
}
func pingPongProtocolProtocolProperty() ProtocolProviderRegister {
	r := buildPingPong()
	return r
}

func buildPingPong() ProtocolProviderRegister {
	pingPong := ProtocolProviderRegister{
		ProtocolId:  p2ptypes.PROTOCOL_PING_PONG,
		MessageType: &protos2.StreamCommonProtocolWrapper{},
	}
	pingPong.ProtocolFactory = &pingPongProtocolFactory{}

	// channels
	pingPongCs := make([]models.ChannelInfo, 0)
	pingPongc := models.ChannelInfo{
		ChannelId:   p2ptypes.CHANNEL_ID_LIFE_CYCLE,
		ChannelSize: 30,
		MessageType: &protos.LifeCycleDefaultChannelMessageWrapper{},
	}

	topics := []models.TopicInfo{{
		Topic: p2ptypes.TOPIC_LIFE_CYCLE,
	}}
	pingPongc.Topics = topics
	pingPongCs = append(pingPongCs, pingPongc)
	pingPong.Channels = pingPongCs
	return pingPong
}

type pingPongProtocolFactory struct {
}

func (p *pingPongProtocolFactory) BundleLocal(req bo.LocalBundleCreateBO) (services.ILocalNetworkBundle, error) {
	return streamimpl.NewLocalPingProtocolNetworkBundle(), nil
}

func (p *pingPongProtocolFactory) BundleRemote(req bo.RemoteBundleCreateBO) (services.IRemoteNetworkBundle, error) {
	return streamimpl.NewRemotePingPongProtocolNetworkBundle(req.NodeID), nil
}

func (p *pingPongProtocolFactory) Local(req bo.LocalProtocolCreateBO) (services.IProtocolService, error) {
	local, _ := p.BundleLocal(bo.LocalBundleCreateBO{})
	return streamimpl.NewDefaultLocalProtocolServiceImpl(req.NodeID, req.ProtocolId,
		req.DataSupport,
		req.ProtocolWrapper,
		req.Channels, local), nil
}

func (p *pingPongProtocolFactory) Remote(req bo.RemoteProtocolCreateBO) (services.IRemoteProtocolService, error) {
	remote, _ := p.BundleRemote(bo.RemoteBundleCreateBO{
		NodeID: req.NodeID,
	})
	return streamimpl.NewDefaultRemoteProtocolServiceImpl(req.NodeID,
		req.ProtocolId,
		req.Status,
		req.Stream,
		req.DataSupport,
		req.MessageType,
		remote), nil
}
