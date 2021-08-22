/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/27 7:47 下午
# @File : peer_lifecycle.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"github.com/hyperledger/fabric-droplib/base/debug"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/event/types"
	"github.com/hyperledger/fabric-droplib/component/p2p/events"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/hyperledger/fabric-droplib/component/switch/impl"
	"github.com/hyperledger/fabric-droplib/protos/p2p/protos"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"sync"
	"time"
)

const (
	PEER_LIFECYCLE_EVENTNAMESPACE = types.EventNameSpace("PEER_LIFECYCLE_EVENTNAMESPACE")
)

var (
	_ services.IPeerLifeCycleReactor = (*peerLifeCycleManagerImpl)(nil)
)

type peerLifeCycleRecorder interface {
	record() <-chan models.PeerRecord
}

type peerLifeCycleManagerImpl struct {
	*impl.NoneSwitchReactor
	peerLock sync.RWMutex
	peers    map[types2.NodeID]*models.PeerLifeCycle

	// 这个bus 也需要修改为通用的
	bus base.IP2PEventBusComponent

	peerLifeCycleService services.IPeerLifeCycleService

	memberChanged chan models.MemberChanged

	status uint32

	localNodeId types2.NodeID
	// FIXME 冗余
	localExternAddress string
	externAddrParse    func(str string) types2.NodeID
}

func NewPeerLifeCycleManagerImpl(cfg *config.P2PConfiguration, bus base.IP2PEventBusComponent, node services.INode) *peerLifeCycleManagerImpl {
	r := &peerLifeCycleManagerImpl{
		peers:              make(map[types2.NodeID]*models.PeerLifeCycle),
		bus:                bus,
		memberChanged:      make(chan models.MemberChanged, 5),
		status:             0,
		localNodeId:        node.ID(),
		localExternAddress: node.GetExternAddress(),
		externAddrParse:    node.FromExternAddress,
	}
	err := bus.SubscribeWithEmit(PEER_LIFECYCLE_EVENTNAMESPACE, new(events.LifeCyclePeerEvent), nil)
	if nil != err {
		panic(err)
	}
	r.NoneSwitchReactor = impl.NewNonSwitchReactor(modules.MODULE_PEER_LIFECYCLE, r)
	r.bus = bus

	r.peerLifeCycleService = NewPeerLifeCycleServiceImpl(cfg.LifeCycleProperty.EnablePingPong, node, func(mems *protos.MemberInfoBroadCastMessageProto) {
		l := len(mems.Members)
		ids := make([]types2.NodeID, l)
		for i := 0; i < l; i++ {
			if id := r.externAddrParse(mems.Members[i].ExternMultiAddress); id == r.localNodeId {
				continue
			}
			ids[i] = r.externAddrParse(mems.Members[i].ExternMultiAddress)
		}

		if len(ids) > 0 {
			r.peerLock.RLock()
			for index, id := range ids {
				if len(id) == 0 {
					continue
				}
				if _, exist := r.peers[id]; !exist {
					go r.OutMemberChanged(models.MemberChanged{
						PeerId:             id,
						ExternMultiAddress: mems.Members[index].ExternMultiAddress,
					})
				}
			}
			r.peerLock.RUnlock()
		}
	})

	return r
}

func (p *peerLifeCycleManagerImpl) MemberChanged() <-chan models.MemberChanged {
	return p.memberChanged
}

// 用于bootstrap
func (p *peerLifeCycleManagerImpl) OutMemberChanged(m models.MemberChanged) {
	p.memberChanged <- m
}

func (p *peerLifeCycleManagerImpl) Status() uint32 {
	return p.status
}

func (p *peerLifeCycleManagerImpl) GetBoundServices() []base.ILogicService {
	return []base.ILogicService{p.peerLifeCycleService}
}

func (p *peerLifeCycleManagerImpl) OnStart() error {
	p.peerLifeCycleService.Start(services2.ASYNC_START_WAIT_READY)
	go p.record()
	return nil
}
func (p *peerLifeCycleManagerImpl) OnStop() {
}
func (p *peerLifeCycleManagerImpl) OnReady() error {
	return nil
}

func (p *peerLifeCycleManagerImpl) record() {
	buffers := make([]models.PeerRecord, 0)
	// 定期广播已知的成员信息
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case r := <-p.peerLifeCycleService.Record():
			// if len(buffers) > 10 {
			if r.PeerId != p.localNodeId {
				buffers = append(buffers, r)
				p.peerLock.RLock()
				if _, exist := p.peers[r.PeerId]; !exist {
					// dest := r.Addr + r.PeerId.Pretty()
					p.Logger.Info("从广播中收到了新的节点:" + r.ExternMultiAddress + ",peerId:" + debug.Green(r.PeerId.Pretty()))
					p.memberChanged <- models.MemberChanged{
						PeerId:             r.PeerId,
						ExternMultiAddress: r.ExternMultiAddress,
						Type:               p2ptypes.MEMBER_CHANGED_ADD,
						Ttl:                peerstore.PermanentAddrTTL,
					}
				}
				p.peerLock.RUnlock()
			}

			if len(buffers) > 0 {
				p.flush(buffers)
				buffers = buffers[0:0]
			}

			per := p.choseRandomPeer()
			if per.Size() > 0 {
				if err := p.bus.Publish(PEER_LIFECYCLE_EVENTNAMESPACE, events.LifeCyclePeerEvent{Event: events.PingPongPeerEvent{PeerId: per}}); nil != err {
					p.Logger.Error("发布event失败:" + err.Error())
				}
			}
		case <-ticker.C:
			data := events.MemberInfoEvent{}

			p.peerLock.RLock()
			for _, v := range p.peers {
				data.Members = append(data.Members, events.MemberInfoNode{
					ExternMultiAddress: v.ExternMultiAddress,
				})
			}
			p.peerLock.RUnlock()
			if len(data.Members) == 0 {
				p.Logger.Debug("该节点尚未已知其他节点,不广播event")
				continue
			}
			data.Members = append(data.Members, events.MemberInfoNode{ExternMultiAddress: p.localExternAddress})
			p.bus.Publish(PEER_LIFECYCLE_EVENTNAMESPACE, events.LifeCyclePeerEvent{Event: data})
		}
	}
}
func (p *peerLifeCycleManagerImpl) choseRandomPeer() types2.NodeID {
	for k, _ := range p.peers {
		return k
	}
	return ""
}

// FIXME
func (p *peerLifeCycleManagerImpl) flush(records []models.PeerRecord) {
	p.Logger.Info(" begin flush peers")

	m := make(map[types2.NodeID]struct{})
	p.peerLock.Lock()
	defer p.peerLock.Unlock()

	for _, r := range records {
		_, exist := p.peers[r.PeerId]
		if !exist {
			p.addPeer(models.NewPeerInfo{
				ID:                 r.PeerId,
				ExternMultiAddress: r.ExternMultiAddress,
			})
			continue
		}
		if _, exist = m[r.PeerId]; exist {
			continue
		} else {
			m[r.PeerId] = struct{}{}
		}
	}

	for k, _ := range m {
		v := p.peers[k]
		v.Lock()
		v.Status = p2ptypes.PEER_STATUS_ALIVE
		v.LastSeen = time.Now()
		v.Unlock()
	}
}

func (p *peerLifeCycleManagerImpl) PeerStatus(pid types2.NodeID) *models.PeerLifeCycle {
	p.peerLock.RLock()
	defer p.peerLock.RUnlock()
	return p.peers[pid]
}

func (p *peerLifeCycleManagerImpl) AddPeer(info models.NewPeerInfo) {
	p.peerLock.Lock()
	defer p.peerLock.Unlock()
	p.addPeer(info)
}
func (p *peerLifeCycleManagerImpl) addPeer(info models.NewPeerInfo) {
	if _, exist := p.peers[info.ID]; exist {
		p.Logger.Warn("重复添加peer,peerid为:" + info.ID.Pretty())
		return
	}
	p.peers[info.ID] = &models.PeerLifeCycle{
		PeerId:             info.ID,
		ExternMultiAddress: info.ExternMultiAddress,
		LastSeen:           time.Now(),
		Status:             p2ptypes.PEER_STATUS_ALIVE,
	}
}



