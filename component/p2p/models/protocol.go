/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 1:35 下午
# @File : protocol.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/common/models"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"strconv"
)

type StreamHandlerWrapper struct {
	Handle network.StreamHandler
}

type ChannelBindInfo struct {
	ChannelID   types.ChannelID
	ChannelSize int
	MessageType services.IChannelMessageWrapper
	Topics []ChannelBindTopic
}

type ProtocolRegistrationInfo struct {
	ProtocolId protocol.ID
	MessageType services.IProtocolWrapper
	Binds       []ChannelBindInfo
}

type TopicChannel struct {
	Topic    string
	Channels []TopicChannelNode
}

type TopicChannelNode struct {
	ChannelID   types.ChannelID
	MessageType services.IChannelMessageWrapper
	In          chan<- models.Envelope
}

func GetTopicAllBoundChannels(info ProtocolRegistrationInfo, bounds []BoundChannel) map[string][]TopicChannelNode {
	type Temp struct {
		Cid         types.ChannelID
		MessageType services.IChannelMessageWrapper
	}
	topics := make(map[string]*[]*Temp)
	for _, bind := range info.Binds {
		for _, tic := range bind.Topics {
			cs, exist := topics[tic.Topic]
			c := make([]*Temp, 0)
			if !exist {
				cs = &c
				topics[tic.Topic] = cs
			} else {
				c = *cs
			}
			added := false
			for _, v := range c {
				if v.Cid == bind.ChannelID {
					added = true
					break
				}
			}
			if !added {
				c = append(c, &Temp{
					Cid:         bind.ChannelID,
					MessageType: bind.MessageType,
				})
			}
		}
	}
	res := make(map[string]*[]TopicChannelNode, 0)
	for k, channelIds := range topics {
		nodes := res[k]
		var tNodes []TopicChannelNode
		if nodes == nil {
			nodes = &tNodes
			res[k]=nodes
		}

		c := *channelIds
		for _, ch := range c {
			var in chan<- models.Envelope
			for _, b := range bounds {
				if b.ChannelID == ch.Cid {
					in = b.Ch
				}
			}
			if nil == in {
				panic("error")
			}
			tNodes = append(tNodes, TopicChannelNode{
				ChannelID:   ch.Cid,
				MessageType: ch.MessageType,
				In:          in,
			})
		}
	}

	ret := make(map[string][]TopicChannelNode)
	for k, v := range res {
		ret[k] = *v
	}
	return ret
}

type BoundChannel struct {
	ChannelID types.ChannelID
	// ChannelSize int
	Ch chan<- models.Envelope
}
type SelfProtocolRegistration struct {
	ProtocolId      protocol.ID
	ProtocolWrapper services.IProtocolWrapper
	BoundChannels   []BoundChannel
}
type RemoteProtocolRegistration struct {
	ProtocolId  protocol.ID
	PeerStatus  func() bool
	MessageType services.IProtocolWrapper
}
type ProtocolRegistration struct {
	ProtocolId    protocol.ID
	BoundChannels []BoundChannel
	PeerStatus    func() bool
}



type P2PTopicForLogic struct {
	Topics map[string]<-chan Envelope
}
type P2PProtocol struct {
	Protocols map[p2pProtocol.ID]*P2PProtocolChannel
}

func (this P2PProtocol) GetMessageWrappers() map[types.ChannelID]services.IChannelMessageWrapper {
	m := make(map[types.ChannelID]services.IChannelMessageWrapper)
	for _, p := range this.Protocols {
		for k, v := range p.Channels {
			m[k] = v.MessageType
		}
	}
	return m
}

type ChannelWrapper struct {
	Channel     *Channel
	MessageType services.IChannelMessageWrapper
}
type P2PProtocolChannel struct {
	Channels map[types.ChannelID]*ChannelWrapper

	ChannelTopics map[types.ChannelID][]string
}

func (this P2PProtocol) GetProtocolByProtocolID(protocolId p2pProtocol.ID) *P2PProtocolChannel {
	v, exist := this.Protocols[protocolId]
	if !exist {
		panic("protocolId:" + protocolId + ",not exist")
	}
	return v
}

func (this P2PProtocolChannel) GetChannelWithNoTopicBound() []*Channel {
	m := make(map[types.ChannelID]struct{})
	for k, _ := range this.ChannelTopics {
		m[k] = struct{}{}
	}
	res := make([]*Channel, 0)
	for _, v := range this.Channels {
		if _, exist := m[v.Channel.ID]; exist {
			continue
		}
		res = append(res, v.Channel)
	}
	return res
}
func (this P2PProtocolChannel) GetChannelById(cid types.ChannelID) *Channel {
	channel, exist := this.Channels[cid]
	if !exist {
		panic("channel:" + strconv.Itoa(int(cid)) + ",不存在")
	}
	return channel.Channel
}

// Deprecate
func (this P2PProtocolChannel) GetChannelByTopic(tic string) *Channel {
	for k, topics := range this.ChannelTopics {
		for _, v := range topics {
			if v == tic {
				return this.Channels[k].Channel
			}
		}
	}
	return nil
	//
	// cids := make([]p2ptypes.ChannelID, 0)
	// for k, topics := range this.ChannelTopics {
	// 	for _, v := range topics {
	// 		if v == tic {
	// 			cids = append(cids, k)
	// 		}
	// 	}
	// }
	// l := len(cids)
	// if l == 0 {
	// 	return nil
	// } else if l == 1 {
	// 	return this.Channels[cids[0]].Out
	// }
	// min := cids[0]
	// for i := 1; i < l; i++ {
	// 	if cids[i] < min {
	// 		min = cids[i]
	// 	}
	// }
	// return nil
}
