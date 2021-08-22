/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 9:58 上午
# @File : dial.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/common/models"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"time"
)

type AsyncStreamHandler func(info *StreamInfo)
type TopicDataHandler func()

type SubConvertFunc func(msg *pubsub.Message) (models.Envelope, error)

type ChannelBindTopic struct {
	Topic string
}

type TopicInfo struct {
	Topic string
}

type ChannelInfo struct {
	ChannelId   types.ChannelID
	ChannelSize int
	MessageType services.IChannelMessageWrapper
	Topics      []TopicInfo
}

type ProtocolInfo struct {
	ProtocolId  protocol.ID
	MessageType services.IProtocolWrapper
	Channels    []ChannelInfo
}

type DialWrapper struct {
	// Pa                peer.AddrInfo
	Addr              string
	Ttl               time.Duration
	InterestProtocols []ProtocolInfo
}
