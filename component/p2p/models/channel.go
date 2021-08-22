/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 10:20 下午
# @File : envelope.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services/types"
	"github.com/hyperledger/fabric-droplib/protos"
	"sync"
)

type ChannelEnvelope struct {
	From      NodeID
	To        NodeID
	Broadcast string
	Message   proto.Message
	ChannelID types.ChannelID

	StreamFlag protos.StreamFlag
	AcquireStreamFromCache bool

	// FIXME
	CallBack func()
}

func (c ChannelEnvelope) Copy() Envelope {
	return Envelope{
		From:       c.From,
		To:         c.To,
		Broadcast:  c.Broadcast,
		Message:    nil,
		StreamFlag: c.StreamFlag,
		AcquireStreamFromCache: c.AcquireStreamFromCache,
		ChannelID:  c.ChannelID,
	}
}

type Envelope struct {
	From      NodeID // sender (empty if outbound)
	To        NodeID // receiver (empty if inbound)
	Broadcast string       // send to all connected peers (ignores To)
	// Message   proto.Message // message payload
	Message []byte
	StreamFlag protos.StreamFlag
	AcquireStreamFromCache bool

	ChannelID types.ChannelID
}

type Channel struct {
	ID   types.ChannelID
	In   <-chan Envelope
	Out  chan<- ChannelEnvelope
	ErrC chan<- PeerError

	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewChannel(id types.ChannelID, inCh <-chan Envelope, outCh chan<- ChannelEnvelope, errCh chan<- PeerError, ) *Channel {
	return &Channel{
		ID: id,
		In:      inCh,
		Out:     outCh,
		ErrC:    errCh,
		closeCh: make(chan struct{}),
	}
}

func (c *Channel) Close() {
	c.closeOnce.Do(func() {
		close(c.closeCh)
		close(c.Out)
		close(c.ErrC)
	})
}

