/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 9:18 上午
# @File : stream_registry.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/debug"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const cmgrTag = "stream-fwd"

var (
	_ services.IStreamRegistry = (*LibP2PStreamRegistry)(nil)
	_ services2.IStream        = (*LibP2PStreamInfo)(nil)
)

type LibP2PStreamInfo struct {
	id uint64

	network.Stream
	status uint32
}

func (this *LibP2PStreamInfo) Status() uint32 {
	return atomic.LoadUint32(&this.status)
}

func (this *LibP2PStreamInfo) Id() int {
	return int(this.id)
}

func NewLibP2PStreamInfo(stream network.Stream) *LibP2PStreamInfo {
	s := &LibP2PStreamInfo{
		Stream: stream,
		status: p2ptypes.STREAM_STATUS_ALIVE,
	}
	return s
}

type ProtoStreamCache struct {
	sync.RWMutex
	status uint32

	ProtocolId protocol.ID
	Streams    map[uint64]*LibP2PStreamInfo
}

func (this *LibP2PStreamInfo) IsHealthy() bool {
	return atomic.LoadUint32(&this.status)&p2ptypes.STREAM_STATUS_NOT_HEALTHY == 0
}

func (this *ProtoStreamCache) Available() bool {
	return atomic.LoadUint32(&this.status) == p2ptypes.PROTOCOL_STATUS_AVALIABLE
}

type PeerProtocolCache struct {
	sync.RWMutex
	status uint32

	PeerId    peer.ID
	Protocols map[protocol.ID]*ProtoStreamCache
}

func (c *PeerProtocolCache) Alive() bool {
	return atomic.LoadUint32(&c.status) == p2ptypes.PEER_STATUS_ALIVE
}

type StreamHookNode struct {
	ToPa        *pubsub.Topic
	DataHandler models.TopicDataHandler
}
type StreamHook struct {
	ToPacs map[string]*StreamHookNode
	info   *models.StreamInfo
}
type LibP2PStreamRegistry struct {
	sync.RWMutex
	*impl.BaseServiceImpl

	connManager connmgr.ConnManager

	ctx   context.Context
	host  host.Host
	sweep func()

	nextId uint64
	conns  map[peer.ID]int

	streams      []*PeerProtocolCache
	streamsIndex map[peer.ID]*PeerProtocolCache
}

func NewLibP2PStreamRegistry(ctx context.Context, host host.Host) *LibP2PStreamRegistry {
	r := &LibP2PStreamRegistry{
		connManager:  host.ConnManager(),
		ctx:          ctx,
		host:         host,
		nextId:       0,
		conns:        make(map[peer.ID]int),
		streamsIndex: make(map[peer.ID]*PeerProtocolCache),
	}
	r.sweep = r.defaultSweep

	r.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_STREAM_REGISTRY, r)
	return r
}

func (r *LibP2PStreamRegistry) Dial(wp models.DialWrapper) (map[protocol.ID]services2.IStream, error) {
	r.WaitUntilReady()

	newMultiaddr, err := multiaddr.NewMultiaddr(wp.Addr)
	if nil != err {
		r.Logger.Error("MemberChanged 收到了新的对象,但是解析为分层地址失败:" + err.Error())
		return nil, err
	}
	addrInfo, err := peer.AddrInfoFromP2pAddr(newMultiaddr)
	if nil != err {
		r.Logger.Error("参数不合法,无法解析为addr:" + wp.Addr)
		return nil, err
	}
	pi := *addrInfo

	err = r.host.Connect(r.ctx, pi)
	if nil != err {
		return nil, err
	}

	r.Logger.Info("连接新的节点:" + pi.String())
	r.host.Peerstore().AddAddr(pi.ID, pi.Addrs[0], wp.Ttl)

	res := make(map[protocol.ID]services2.IStream)

	for _, protocol := range wp.InterestProtocols {
		stream, err := r.host.NewStream(r.ctx, pi.ID, protocol.ProtocolId)
		if nil != err {
			r.Logger.Error("创建stream失败,protocolid为:" + string(protocol.ProtocolId) + ",err:" + err.Error())
			return nil, err
		}
		//
		info := NewLibP2PStreamInfo(stream)

		r.registerStream(pi.ID, protocol, info)
		res[protocol.ProtocolId] = info
	}

	return res, nil
}

func (this *LibP2PStreamRegistry) registerStream(peerId peer.ID, pw models.ProtocolInfo, info *LibP2PStreamInfo) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	this.connManager.TagPeer(peerId, cmgrTag, 20)
	this.conns[peerId]++
	info.id = this.nextId
	this.nextId++

	var pc *PeerProtocolCache
	if peerPro, exist := this.streamsIndex[peerId]; exist {
		pc = peerPro
	} else {
		pc = &PeerProtocolCache{
			RWMutex:   sync.RWMutex{},
			status:    p2ptypes.PEER_STATUS_ALIVE,
			PeerId:    peerId,
			Protocols: make(map[protocol.ID]*ProtoStreamCache),
		}
		this.streamsIndex[peerId] = pc
		this.streams = append(this.streams, pc)
	}

	pc.Lock()
	var prot *ProtoStreamCache
	if v, exist := pc.Protocols[pw.ProtocolId]; exist {
		prot = v
	} else {
		prot = &ProtoStreamCache{
			ProtocolId: pw.ProtocolId,
			Streams:    make(map[uint64]*LibP2PStreamInfo),
		}
		pc.Protocols[pw.ProtocolId] = prot
	}
	prot.Lock()
	_, exist := prot.Streams[info.id]
	if exist {
		// FIXME
	}
	prot.Streams[info.id] = info
	prot.Unlock()
	pc.Unlock()
}

func (this *LibP2PStreamRegistry) AcquireNewStream(pid types2.NodeID, protocolId protocol.ID, useCache bool) (services2.IStream, error) {
	peerId := peer.ID(pid)
	if !useCache {
		this.Logger.Info("disable cache 尝试或者创建新的stream,peerId为:" + peerId.Pretty() + ",protocolId为:" + string(protocolId))
		stream, err := this.host.NewStream(this.ctx, peerId, protocolId)
		if nil != err {
			this.Logger.Info(debug.Red("peerId:" + peerId.Pretty() + ",protocolId:" + string(protocolId) + ",获取stream失败"))
			return nil, err
		}
		res := NewLibP2PStreamInfo(stream)
		return res, nil
	}
	this.Logger.Info("use cache 尝试或者创建新的stream,peerId为:" + peerId.Pretty() + ",protocolId为:" + string(protocolId))
	idleStream := this.getIdleStream(peerId, protocolId)
	if idleStream != nil {
		this.Logger.Info(debug.Cyan("peerId:" + peerId.Pretty() +
			",protocolId:" + string(protocolId) + ",取用缓存着的stream,streamId为:" + strconv.Itoa(int(idleStream.id))))
		return idleStream, nil
	}

	stream, err := this.host.NewStream(this.ctx, peerId, protocolId)
	if nil != err {
		this.Logger.Info(debug.Red("peerId:" + peerId.Pretty() + ",protocolId:" + string(protocolId) + ",获取stream失败"))
		return nil, err
	}
	res := NewLibP2PStreamInfo(stream)
	this.Logger.Info("peerId:" + peerId.Pretty() + ",protocolId:" + string(protocolId) + ",创建新的stream,streamId为:" + strconv.Itoa(int(res.id)))
	this.registerStream(peerId, models.ProtocolInfo{
		ProtocolId: protocolId,
	}, res)

	return res, nil
}

func (this *LibP2PStreamRegistry) getIdleStream(peerId peer.ID, protocolId protocol.ID) *LibP2PStreamInfo {
	this.RWMutex.RLock()
	peerC, exist := this.streamsIndex[peerId]
	if !exist {
		this.RWMutex.RUnlock()
		return nil
	}
	this.RWMutex.RUnlock()

	peerC.RLock()
	p, exist := peerC.Protocols[protocolId]
	if !exist {
		peerC.RUnlock()
		return nil
	}
	p.RLock()
	for _, stream := range p.Streams {
		if stream.IsHealthy() {
			p.RUnlock()
			peerC.RUnlock()
			return stream
		}
	}
	p.RUnlock()
	peerC.RUnlock()

	return nil
}

func (this *LibP2PStreamRegistry) RecycleStream(stream services2.IStream, flag protos.StreamFlag) services2.IStream {
	close := flag&protos.StreamFlag_CLOSE_IMMEDIATELY >= protos.StreamFlag_CLOSE_IMMEDIATELY
	if close {
		this.Logger.Info("回收stream,close该stream,streamId为:", stream)
		stream.Reset()
		return nil
	}
	return stream
}

func (this *LibP2PStreamRegistry) OnStart() error {
	go this.sweep()
	return nil
}

func (this *LibP2PStreamRegistry) defaultSweep() {
	this.Logger.Info("gc running ")
	for {
		if !this.IsRunning() {
			return
		}
		externRLock := func() {
			this.RWMutex.RLock()
		}
		externRUnlock := func() {
			this.RWMutex.RUnlock()
		}

		externRLock()
		// outLocked := true
		for _, c := range this.streams {
			c.RWMutex.RLock()
			// if outLocked {
			// 	externRUnlock()
			// 	outLocked = false
			// }

			if !c.Alive() {
				this.Logger.Error("死了")
				c.RWMutex.RUnlock()
				continue
			}
			// peerLocked := true
			for _, v := range c.Protocols {
				// if !v.Available() {
				// 	this.Logger.Error(debug.Red("外层检测为alive,内部全是dead,peer:" + string(c.PeerId) + ",protocol:" + string(v.ProtocolId)))
				// 	continue
				// }
				v.Lock()
				// if peerLocked {
				// 	c.RWMutex.RUnlock()
				// 	peerLocked = false
				// }

				for _, s := range v.Streams {
					if s.IsHealthy() {
						this.Logger.Info(debug.Red("[healthy]:peerId:" + c.PeerId.Pretty() + ",protocolId:" + string(v.ProtocolId) + ",streamId:" + strconv.Itoa(int(s.id))))
						continue
					}
					this.Logger.Info(debug.Red("[回收]:peerId:" + c.PeerId.Pretty() + ",protocolId:" + string(v.ProtocolId) + ",streamId:" + strconv.Itoa(int(s.id))))
					// remove
					delete(v.Streams, s.id)

				}
				v.Unlock()
			}
			c.RWMutex.RUnlock()
		}
		externRUnlock()

		// if outLocked{
		// 	this.RWMutex.RUnlock()
		// }
		// FIXME,tikTok
		time.Sleep(time.Second * 20)
	}
}

// FIXME CAS
func (this *LibP2PStreamInfo) Close() error {
	atomic.StoreUint32(&this.status, p2ptypes.STREAM_STATUS_CLOSE_ALL)
	return this.Stream.Close()
}

func (this *LibP2PStreamInfo) CloseWrite() error {
	atomic.StoreUint32(&this.status, p2ptypes.STREAM_STATUS_CLOSE_WRITE)
	return this.Stream.CloseWrite()
}

func (this *LibP2PStreamInfo) CloseRead() error {
	atomic.StoreUint32(&this.status, p2ptypes.STREAM_STATUS_CLOSE_READ)
	return this.Stream.CloseRead()
}

func (this *LibP2PStreamInfo) Reset() error {
	atomic.StoreUint32(&this.status, p2ptypes.STREAM_STATUS_RESET)
	return this.Stream.Reset()
}
