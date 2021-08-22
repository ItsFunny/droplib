/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:22 上午
# @File : pubsub.go
# @Description :
# @Attention :
*/
package impl
//
// import (
// 	"context"
// 	"github.com/hyperledger/fabric-droplib/base/log/modules"
// 	"github.com/hyperledger/fabric-droplib/base/types"
// 	libutils "github.com/hyperledger/fabric-droplib/base/utils"
// 	models2 "github.com/hyperledger/fabric-droplib/common/models"
// 	"github.com/hyperledger/fabric-droplib/component/base"
// 	"github.com/hyperledger/fabric-droplib/component/pubsub/models"
// 	"github.com/hyperledger/fabric-droplib/component/pubsub/services"
// 	"github.com/libp2p/go-libp2p-core/host"
// 	"github.com/libp2p/go-libp2p-core/peer"
// 	psub "github.com/libp2p/go-libp2p-pubsub"
// 	"strconv"
// 	"sync"
// )
//
// var (
// 	_ services.ILibP2PPubSubComponent = (*LibP2PPubSubComponentImpl)(nil)
// )
//
// type topic struct {
// 	channels []types.ChannelID
// 	topic    *psub.Topic
// }
//
// type LibP2PPubSubComponentImpl struct {
// 	*base.BaseLogicComponent
// 	sync.RWMutex
//
// 	ctx context.Context
//
// 	ps *psub.PubSub
//
// 	topics map[string]*topic
// }
//
//
// func NewLibP2PPubSubComponentImpl(ctx context.Context,host host.Host, ) *LibP2PPubSubComponentImpl {
// 	ps, err := psub.NewGossipSub(ctx, host)
// 	libutils.PanicIfError("创建gossip sub失败", err)
//
// 	r := &LibP2PPubSubComponentImpl{
// 		ps:      ps,
// 		ctx:     ctx,
// 		RWMutex: sync.RWMutex{},
// 		topics:  make(map[string]*topic),
// 	}
// 	r.BaseLogicComponent = base.NewBaseLogicCompoennt(modules.MODULE_PUBSUB, r)
// 	return r
// }
//
// func (p *LibP2PPubSubComponentImpl) GetBoundServices() []base.ILogicService {
// 	return nil
// }
//
//
// func (p *LibP2PPubSubComponentImpl) RegisterTopic(req models.TopicRegisterReq) {
// 	p.Lock()
// 	defer p.Unlock()
// 	tic, exist := p.topics[req.TopicID]
// 	if !exist {
// 		tic = &topic{}
// 		p.topics[req.TopicID] = tic
// 	} else {
// 		panic("not likely")
// 	}
//
// 	hashM := make(map[types.ChannelID]struct{})
// 	for _, c := range tic.channels {
// 		hashM[c] = struct{}{}
// 	}
// 	for _, node := range req.Chs {
// 		if _, exist = hashM[node.ChannelID]; !exist {
// 			tic.channels = append(tic.channels, node.ChannelID)
// 		}
// 	}
// 	topic, err := p.ps.Join(req.TopicID)
// 	libutils.PanicIfError("订阅事件", err)
// 	tic.topic = topic
// 	// copy
// 	m := make(map[types.ChannelID]chan<- models2.Envelope)
// 	for _, node := range req.Chs {
// 		m[node.ChannelID] = node.In
// 		node.In = nil
// 	}
//
// 	go p.subRecvRoutine(p.ctx, req.LocalNodeId, topic, m)
// }
//
// func (r *LibP2PPubSubComponentImpl) subRecvRoutine(ctx context.Context, localNode types.NodeID, topic *psub.Topic, m map[types.ChannelID]chan<- models2.Envelope) {
// 	sub, err := topic.Subscribe()
// 	if nil != err {
// 		panic(err)
// 	}
// 	for {
// 		msg, err := sub.Next(ctx)
// 		if nil != err {
// 			r.Logger.Error("获取event失败:" + err.Error())
// 			continue
// 		}
// 		id := msg.ReceivedFrom
// 		// FIXME 需要变更
// 		if msg.ReceivedFrom == peer.ID(localNode) {
// 			continue
// 		}
//
// 		r.Logger.Info("收到广播消息:from:" + id.Pretty() + msg.String() + " ")
// 		e := models2.Envelope{
// 			From:      types.NodeID(id),
// 			Broadcast: *msg.Topic,
// 			Message:   msg.Data,
// 		}
// 		for cid, in := range m {
// 			e.ChannelID = cid
// 			select {
// 			case in <- e:
// 			default:
// 				r.Logger.Warn("channelId为:" + strconv.Itoa(int(cid)) + " 的channel阻塞,routine")
// 				go func() { in <- e }()
// 			}
// 		}
// 	}
// }
//
// func (p *LibP2PPubSubComponentImpl) RegisterTopicWithOutRoutine(topicString string) *models.OutTopic {
// 	p.Lock()
// 	defer p.Unlock()
//
// 	tic, exist := p.topics[topicString]
// 	if !exist {
// 		tic = &topic{}
// 		p.topics[topicString] = tic
// 	} else {
// 		panic("not likely")
// 	}
//
// 	topic, err := p.ps.Join(topicString)
// 	libutils.PanicIfError("订阅事件", err)
// 	tic.topic = topic
// 	res := &models.OutTopic{
// 		Topic: topic,
// 	}
// 	return res
// }
//
// func (p *LibP2PPubSubComponentImpl) Publish(ctx context.Context, topicId string, data []byte) error {
// 	return p.topics[topicId].topic.Publish(ctx, data)
// }
