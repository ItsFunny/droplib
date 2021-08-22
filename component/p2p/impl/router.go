/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 1:29 下午
# @File : node_registry.go
# @Description :
*/
package streamimpl

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	libutils "github.com/hyperledger/fabric-droplib/base/utils"
	"github.com/hyperledger/fabric-droplib/common/errors"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	services3 "github.com/hyperledger/fabric-droplib/component/p2p/services"
	models3 "github.com/hyperledger/fabric-droplib/component/pubsub/models"
	services2 "github.com/hyperledger/fabric-droplib/component/switch/services"
	"github.com/libp2p/go-libp2p-core/peer"
	p2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type protocolChannel map[types.ChannelID]impl.Queue

type peerProtocolQueue struct {
	queues map[types.ChannelID]impl.Queue
	sync.RWMutex
	status uint32
}

func (this *peerProtocolQueue) waitUntilReady() {
	c := func() bool {
		return atomic.LoadUint32(&this.status) == 1
	}
	for !c() {
		time.Sleep(time.Second * 3)
	}
}

type Router struct {
	*impl.BaseServiceImpl

	swch services2.ISwitch

	// node *models.Node
	node services3.INode
	ctx  context.Context

	bus base.IP2PEventBusComponent

	peerMtx sync.RWMutex
	// FIXME ,更应该是1个peer 一个queue,而不是1个peer 对应n*m个queue
	peerProtocols map[types.NodeID]*peerProtocolQueue

	peerLifeCycleReactor services3.IPeerLifeCycleReactor

	streamRegistry services3.IStreamRegistry

	pubsubManager services3.IPubSubManager

	protoMtx         sync.RWMutex
	protocolChannels map[p2pProtocol.ID]protocolChannel
	protocolMessages map[p2pProtocol.ID]services.IProtocolWrapper

	protocolLifecycle services3.IProtocolLifeCycleReactor

	status uint32
}

func (r *Router) GetEventBusComponent() base.IP2PEventBusComponent {
	return r.bus
}

func (r *Router) OutMemberChanged(m models.MemberChanged) {
	r.peerLifeCycleReactor.OutMemberChanged(m)
}
func (r *Router) GetProtocolLifeCycleManager() services3.IProtocolLifeCycleReactor {
	return r.protocolLifecycle
}

func (r *Router) GetBoundServices() []base.ILogicService {
	res := r.peerLifeCycleReactor.GetBoundServices()
	res = append(res, r.pubsubManager.GetBoundServices()...)
	return res
}

func NewRouter(node services3.INode,
	bus base.IP2PEventBusComponent,
	registry services3.IStreamRegistry,
	cycle services3.IProtocolLifeCycleReactor,
	managerImpl services3.IPeerLifeCycleReactor,
	pubsubManager services3.IPubSubManager) *Router {
	ctx := context.Background()
	r := &Router{
		node:                 node,
		ctx:                  ctx,
		bus:                  bus,
		peerMtx:              sync.RWMutex{},
		peerProtocols:        make(map[types.NodeID]*peerProtocolQueue),
		peerLifeCycleReactor: managerImpl,
		streamRegistry:       registry,
		pubsubManager:        pubsubManager,
		protocolChannels:     make(map[p2pProtocol.ID]protocolChannel),
		protocolMessages:     make(map[p2pProtocol.ID]services.IProtocolWrapper),
		protocolLifecycle:    cycle,
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_ROUTER, r)

	return r
}
func (r *Router) GetNode() services3.INode {
	return r.node
}

func (r *Router) OnStart() error {
	r.peerLifeCycleReactor.Start(services.ASYNC_START_WAIT_READY)
	r.streamRegistry.Start(services.ASYNC_START_WAIT_READY)
	r.pubsubManager.Start(services.ASYNC_START_WAIT_READY)
	r.protocolLifecycle.Start(services.ASYNC_START_WAIT_READY)
	r.bus.Start(services.ASYNC_START_WAIT_READY)

	go r.dialPeers()

	return nil
}
func (r *Router) OnReady() error {
	r.peerLifeCycleReactor.ReadyIfPanic(services.READY_UNTIL_START)
	r.streamRegistry.ReadyIfPanic(services.READY_UNTIL_START)
	r.pubsubManager.ReadyIfPanic(services.READY_UNTIL_START)
	r.protocolLifecycle.ReadyIfPanic(services.READY_UNTIL_START)
	r.bus.ReadyIfPanic(services.READY_UNTIL_START)

	return nil
}

func (r *Router) incase() {
	r.WaitUntilReady()
}

func (r *Router) Setup(notify chan struct{}) *models2.P2PProtocol {
	res := &models2.P2PProtocol{
		Protocols: make(map[p2pProtocol.ID]*models2.P2PProtocolChannel),
	}

	for _, proc := range r.node.GetProtocols() {
		binds := make([]models.ChannelBindInfo, 0)
		for _, bind := range proc.Channels {
			topics := make([]models.ChannelBindTopic, 0)
			for _, tic := range bind.Topics {
				topics = append(topics, models.ChannelBindTopic{
					Topic: tic.Topic,
					// SubConvertFunc: tic.Conv,
				})
			}
			binds = append(binds, models.ChannelBindInfo{
				ChannelID:   bind.ChannelId,
				ChannelSize: bind.ChannelSize,
				MessageType: bind.MessageType,
				Topics:      topics,
			})
		}
		req := models.ProtocolRegistrationInfo{
			ProtocolId:  proc.ProtocolId,
			MessageType: proc.MessageType,
			Binds:       binds,
		}
		res.Protocols[proc.ProtocolId] = r.setupProtocol(req)
	}
	close(notify)

	if err := r.Ready(services.READY_UNTIL_START); nil != err {
		panic(err)
	}

	return res
}

func (r *Router) dialPeers() {
	r.Logger.Info("starting dial routine")
	// ctx := r.signalCtx()
	protocols := r.getLocalProtocols()
	for {
		if !r.IsRunning() {
			return
		}
		select {
		case addr := <-r.peerLifeCycleReactor.MemberChanged():
			r.Logger.Info("收到了新的节点加入请求:" + addr.ExternMultiAddress)

			exist := false
			r.peerMtx.Lock()
			_, exist = r.peerProtocols[addr.PeerId]
			if !exist {
				r.peerProtocols[addr.PeerId] = &peerProtocolQueue{
					queues: make(map[types.ChannelID]impl.Queue),
				}
			}
			r.peerMtx.Unlock()

			// go func() {
			// FIXME error ?  drop or not  ?
			if _, e := r.streamRegistry.Dial(models.DialWrapper{
				Addr:              addr.ExternMultiAddress,
				Ttl:               addr.Ttl,
				InterestProtocols: protocols,
			}); nil != e {
				r.Logger.Error("本地节点为:" + r.node.GetExternAddress() + ",远端节点为:" + addr.ExternMultiAddress + ",dial peer failed for :" + e.Error() + "")
			} else {
				r.peerLifeCycleReactor.AddPeer(models.NewPeerInfo{
					ID:                 addr.PeerId,
					ExternMultiAddress: addr.ExternMultiAddress,
				})
				// FIXME OPTIMIZE
				// 初始化相关信息
				protocolReg := models.RemoteProtocolRegistration{
					PeerStatus: r.peerLifeCycleReactor.PeerStatus(addr.PeerId).Alive,
				}
				for _, protocol := range protocols {
					protocolReg.ProtocolId = protocol.ProtocolId
					protocolReg.MessageType = protocol.MessageType

					remoteProtocol, err := r.protocolLifecycle.RegisterRemoteProtocol(addr.PeerId, r.streamRegistry, protocolReg)
					if nil != err {
						if err != ERR_REPEAT_REGISTER_PROTOCOL {
							r.Logger.Error("注册remote protocol失败:" + err.Error())
							continue
						}
					}

					protocolService := r.protocolLifecycle.GetProtocolByPeerIdAndProtocolId(addr.PeerId, protocol.ProtocolId)
					if nil == protocolService {
						panic("not likely ,should not happen" + fmt.Sprintf("peerId:%s,protocolId:%s", addr.PeerId, protocol.ProtocolId))
					}
					r.protoMtx.Lock()
					protocolQueue := r.peerProtocols[addr.PeerId]
					r.protoMtx.Unlock()

					protocolQueue.Lock()
					for _, cid := range protocol.Channels {
						q := impl.NewFIFOQueue(0)
						protocolQueue.queues[cid.ChannelId] = q
						r.Logger.Info("为远端节点:" + addr.PeerId.Pretty() + ",protocol:" + string(protocol.ProtocolId) + ",注册channel:" + strconv.Itoa(int(cid.ChannelId)))
						go remoteProtocol.SendWithStream(r.streamRegistry, q.Dequeue())
					}
					// FIXME
					protocolQueue.status = 1

					protocolQueue.Unlock()
				}

			}
			// }()
		}
	}
}

func (r *Router) getLocalProtocols() []models.ProtocolInfo {
	return r.node.GetProtocols()
}
func (r *Router) signalCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-r.Quit()
		cancel()
	}()
	return ctx
}

func (r *Router) setupProtocol(prot models.ProtocolRegistrationInfo) *models2.P2PProtocolChannel {
	res := &models2.P2PProtocolChannel{
		Channels:      make(map[types.ChannelID]*models2.ChannelWrapper),
		ChannelTopics: make(map[types.ChannelID][]string),
	}

	bounds := make([]models.BoundChannel, 0)

	for _, c := range prot.Binds {
		channel, queue, err := r.OpenChannel(prot.ProtocolId, c.ChannelID, prot.MessageType, c.MessageType, c.ChannelSize)
		libutils.PanicIfError("openChannel", err)
		cwp := &models2.ChannelWrapper{
			Channel:     channel,
			MessageType: c.MessageType,
		}
		res.Channels[c.ChannelID] = cwp
		bounds = append(bounds, models.BoundChannel{
			ChannelID: c.ChannelID,
			Ch:        queue.Enqueue(),
		})
		for _, tic := range c.Topics {
			res.ChannelTopics[c.ChannelID] = append(res.ChannelTopics[c.ChannelID], tic.Topic)
		}
	}

	r.bindTopics(models.GetTopicAllBoundChannels(prot, bounds))

	selfR := models.SelfProtocolRegistration{
		ProtocolId:      prot.ProtocolId,
		ProtocolWrapper: prot.MessageType,
		BoundChannels:   bounds,
	}

	protocol, err := r.protocolLifecycle.SetupSelfProtocol(r.node, selfR)
	libutils.PanicIfError("注册protocol", err)
	r.node.SetStreamHandler(prot.ProtocolId, protocol)

	return res
}

func (r *Router) OpenChannel(id p2pProtocol.ID, cid types.ChannelID, protocolType services.IProtocolWrapper, messageType services.IChannelMessageWrapper, size int) (*models2.Channel, impl.Queue, error) {
	queue := impl.NewFIFOQueue(size)
	outCh := make(chan models2.ChannelEnvelope, size)
	errCh := make(chan errors.PeerError, size)
	channel := models2.NewChannel(cid, queue.Dequeue(), outCh, errCh)

	r.protoMtx.Lock()
	defer r.protoMtx.Unlock()
	var protocolCh protocolChannel
	if v, exist := r.protocolChannels[id]; !exist {
		protocolCh = make(map[types.ChannelID]impl.Queue)
		r.protocolChannels[id] = protocolCh
	} else {
		protocolCh = v
	}

	if _, ok := protocolCh[cid]; ok {
		return nil, nil, fmt.Errorf("channel %v already exists", id)
	}
	protocolCh[cid] = queue
	r.protocolMessages[id] = protocolType

	go func() {
		defer func() {
			// FIXME
			r.protoMtx.Lock()
			delete(r.protocolChannels[id], cid)
			r.protoMtx.Unlock()
			queue.Close()
		}()

		// TODO ctx ,context 需要修改为一个公共的
		r.routeChannel(context.Background(), id, cid, outCh, errCh, messageType)
	}()

	return channel, queue, nil
}

func (r *Router) lazyRoute(q impl.Queue,
	id p2pProtocol.ID, cid types.ChannelID,
	outCh <-chan models2.ChannelEnvelope,
	errCh chan errors.PeerError,
	messageType services.IChannelMessageWrapper) {

	queue := q
	q = nil
	go func() {
		defer func() {
			// FIXME
			r.protoMtx.Lock()
			delete(r.protocolChannels[id], cid)
			r.protoMtx.Unlock()
			queue.Close()
		}()

		// TODO ctx ,context 需要修改为一个公共的
		r.routeChannel(context.Background(), id, cid, outCh, errCh, messageType)
	}()
}

func (r *Router) bindTopics(topics map[string][]models.TopicChannelNode) {
	req := models3.TopicRegisterReq{
		LocalNodeId: r.node.ID(),
	}
	for k, nodes := range topics {
		req.TopicID = k
		for _, c := range nodes {
			n := models3.TopicRegisterChannelNode{
				ChannelID: c.ChannelID,
				In:        c.In,
			}
			req.Chs = append(req.Chs, n)
		}
		r.pubsubManager.RegisterTopic(req)
		req.Chs = nil
	}
}

func (r *Router) routeChannel(ctx context.Context,
	protocolId p2pProtocol.ID, channelId types.ChannelID,
	outC <-chan models2.ChannelEnvelope, errC chan errors.PeerError, wrapper services.IChannelMessageWrapper) {
	for {
		select {
		case enve, ok := <-outC:
			if !ok {
				return
			}
			streamEnv := enve.Copy()
			enve.ChannelID = channelId
			if msg, err := wrapper.Encode(enve.Message); nil != err {
				r.Logger.Error("encode失败:" + err.Error())
				continue
			} else {
				streamEnv.Message = msg
			}

			if len(streamEnv.Broadcast) > 0 {
				// pubsub
				if err := r.pubsubManager.Publish(ctx, streamEnv.Broadcast, streamEnv.Message); nil != err {
					r.Logger.Error("发送消息失败:" + err.Error())
				}
				continue
			}
			pid := streamEnv.To

			// FIXME ,不需要用锁
			r.peerMtx.RLock()
			protocols, exist := r.peerProtocols[pid]
			r.peerMtx.RUnlock()
			if !exist {
				// FIXME ,调试结束之后不可以panic
				time.Sleep(time.Second * 5)
				r.peerMtx.RLock()
				protocols, exist = r.peerProtocols[pid]
				r.peerMtx.RUnlock()
				if !exist {
					panic(fmt.Sprintf("本节点id:%s,peerid:%s,protocolId为:%s 对应的peerProtocol不存在,panic", r.node.ID(), pid, protocolId))
				}
			}
			protocols.waitUntilReady()

			protocols.RLock()
			q, exist := protocols.queues[streamEnv.ChannelID]
			protocols.RUnlock()
			if !exist {
				// FIXME ,调试结束之后不可以panic
				panic(fmt.Sprintf("peerid:%s,protocolId为:%s,chId:%d 对应的channel不存在,panic", pid, protocolId, streamEnv.ChannelID))
			}
			select {
			case <-r.Quit():
				return
			case <-q.Closed():
				r.Logger.Debug("dropping message for unconnected peer",
					"peer", streamEnv.To, "channel", channelId)
			case q.Enqueue() <- streamEnv:
			}
		case peerError, ok := <-errC:
			if !ok {
				return
			}
			r.Logger.Error("peer error, evicting", "peer", peerError.NodeID, "err", peerError.Err)
			// FIXME ,需要关闭连接,移除addr
			// }

		case <-r.Quit():
			return
		}
	}
}

func (r *Router) fromNodeId(id types.NodeID) peer.ID {
	return peer.ID(id)
}
