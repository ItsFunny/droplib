/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 5:13 下午
# @File : protocol.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/bo"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/hyperledger/fabric-droplib/component/switch/impl"
	"github.com/libp2p/go-libp2p-core/protocol"
	"sync"
	"sync/atomic"
)

var (
	_ services.IProtocolLifeCycleReactor = (*ProtocolLifeCycleManagerImpl)(nil)
)

var ()

var (
	ERR_REPEAT_REGISTER_PROTOCOL = errors.New("repeat ")
)

type PeerProtocols struct {
	sync.RWMutex
	protocols map[protocol.ID]services.IRemoteProtocolService
}
type ProtocolLifeCycleManagerImpl struct {
	*impl.NoneSwitchReactor

	sync.RWMutex
	// self
	selfProtocols map[protocol.ID]services.IProtocolService

	protocolFactory map[protocol.ID]services.ProtocolFactory

	dataSupport services2.StreamDataSupport

	peerProtocols map[types.NodeID]*PeerProtocols

	status uint32
}

func (p *ProtocolLifeCycleManagerImpl) OnReady() error {
	for _, v := range p.selfProtocols {
		if err := v.Ready(services2.READY_UNTIL_START); nil != err {
			panic(err)
		}
	}
	return nil
}

func (p *ProtocolLifeCycleManagerImpl) AppendProtocolProvider(id protocol.ID, proFactory services.ProtocolFactory) {
	p.Lock()
	defer p.Unlock()

	if _, exist := p.protocolFactory[id]; exist {
		panic("重复注册,proFactory")
	}
	p.protocolFactory[id] = proFactory
}

func (p *ProtocolLifeCycleManagerImpl) Status() uint32 {
	return atomic.LoadUint32(&p.status)
}

func (p *ProtocolLifeCycleManagerImpl) Notify() {
	atomic.StoreUint32(&p.status, p2ptypes.MANAGER_READY)
}

func (p *ProtocolLifeCycleManagerImpl) GetBoundServices() []base.ILogicService {
	return nil
}

func NewProtocolLifeCycleManagerImpl(dataSupport services2.StreamDataSupport, ) *ProtocolLifeCycleManagerImpl {
	r := &ProtocolLifeCycleManagerImpl{
		RWMutex:         sync.RWMutex{},
		selfProtocols:   make(map[protocol.ID]services.IProtocolService),
		protocolFactory: make(map[protocol.ID]services.ProtocolFactory),
		dataSupport:     dataSupport,
		peerProtocols:   make(map[types.NodeID]*PeerProtocols),
	}
	r.NoneSwitchReactor = impl.NewNonSwitchReactor(modules.MODULE_PROTOCOL_LIFECYCLE, r)
	return r
}

func (p *ProtocolLifeCycleManagerImpl) SetupSelfProtocol(node services.INode, info models.SelfProtocolRegistration) (services.IProtocolService, error) {
	p.Logger.Info("开始初始化自身,注册protocol:" + string(info.ProtocolId) + ",host为:" + node.GetExternAddress())
	p.Lock()
	defer p.Unlock()

	if _, exist := p.selfProtocols[info.ProtocolId]; exist {
		return nil, errors.New("重复注册,protocolId:" + string(info.ProtocolId))
	}

	cq := make(map[types.ChannelID]chan<- models2.Envelope)
	for _, ch := range info.BoundChannels {
		cq[ch.ChannelID] = ch.Ch
	}

	proFactory, exist := p.protocolFactory[info.ProtocolId]
	if !exist {
		return nil, errors.New("未注册protocolFactory")
	}

	proReq := bo.LocalProtocolCreateBO{
		NodeID:          node.ID(),
		ProtocolId:      info.ProtocolId,
		DataSupport:     p.dataSupport,
		ProtocolWrapper: info.ProtocolWrapper,
		Channels:        cq,
	}
	proImpl, err := proFactory.Local(proReq)
	if nil != err {
		return nil, err
	}
	p.selfProtocols[info.ProtocolId] = proImpl

	proImpl.Start(services2.ASYNC_START_WAIT_READY)

	return proImpl, nil
}

func (p *ProtocolLifeCycleManagerImpl) RegisterRemoteProtocol(remotePeerId types.NodeID, registry services.IStreamRegistry, info models.RemoteProtocolRegistration) (services.IRemoteProtocolService, error) {
	p.Logger.Info("开始为remotePeer:" + remotePeerId.Pretty() + "注册protocol:" + string(info.ProtocolId))

	var pr *PeerProtocols
	var exist bool

	p.Lock()
	if pr, exist = p.peerProtocols[remotePeerId]; !exist {
		pr = &PeerProtocols{
			protocols: make(map[protocol.ID]services.IRemoteProtocolService),
		}
		p.peerProtocols[remotePeerId] = pr
	}
	p.Unlock()

	var proService services.IRemoteProtocolService

	pr.Lock()
	if proService, exist = pr.protocols[info.ProtocolId]; exist {
		p.Logger.Warn("peerId:" + remotePeerId.Pretty() + "protocolId为:" + string(info.ProtocolId) + ",重复注册")
		return nil, ERR_REPEAT_REGISTER_PROTOCOL
	}
	stream, err := registry.AcquireNewStream(remotePeerId, info.ProtocolId, false)
	if nil != err {
		return nil, errors.New("获取newStream失败")
	}

	protocolReq := bo.RemoteProtocolCreateBO{
		NodeID:      remotePeerId,
		ProtocolId:  info.ProtocolId,
		Status:      info.PeerStatus,
		Stream:      stream,
		DataSupport: p.dataSupport,
		MessageType: info.MessageType,
	}
	pro, exist := p.protocolFactory[info.ProtocolId]
	if !exist {
		return nil, errors.New("未注册protocol factory")
	}
	proService, err = pro.Remote(protocolReq)
	if nil != err {
		return nil, err
	}
	pr.protocols[info.ProtocolId] = proService
	pr.Unlock()

	proService.Start(services2.ASYNC_START)
	proService.Ready(services2.READY_UNTIL_START)

	return proService, nil
}

func (p *ProtocolLifeCycleManagerImpl) GetProtocolByPeerIdAndProtocolId(peerId types.NodeID, protocolId protocol.ID) services.IRemoteProtocolService {
	p.RLock()
	pr, exist := p.peerProtocols[peerId]
	if !exist {
		p.RUnlock()
		return nil
	}
	p.RUnlock()

	pr.RLock()
	defer pr.RUnlock()
	return pr.protocols[protocolId]
}
