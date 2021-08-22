/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 9:04 上午
# @File : stream.go
# @Description :
# @Attention :
*/
package models

import (
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/libp2p/go-libp2p-core/network"
	"sync/atomic"
)

// 
type StreamInfo struct {
	Id uint64

	network.Stream
	Status uint32
}

// FIXME CAS
func (this *StreamInfo) Close() error {
	atomic.StoreUint32(&this.Status, p2ptypes.STREAM_STATUS_CLOSE_ALL)
	return this.Stream.Close()
}

func (this *StreamInfo) CloseWrite() error {
	atomic.StoreUint32(&this.Status, p2ptypes.STREAM_STATUS_CLOSE_WRITE)
	return this.Stream.CloseWrite()
}

func (this *StreamInfo) CloseRead() error {
	atomic.StoreUint32(&this.Status, p2ptypes.STREAM_STATUS_CLOSE_READ)
	return this.Stream.CloseRead()
}

func (this *StreamInfo) Reset() error {
	atomic.StoreUint32(&this.Status, p2ptypes.STREAM_STATUS_RESET)
	return this.Stream.Reset()
}

func NewStreamInfo(stream network.Stream) *StreamInfo {
	s := &StreamInfo{
		Stream: stream,
		Status: p2ptypes.STREAM_STATUS_ALIVE,
	}
	return s
}

func (this *StreamInfo) IsHealthy() bool {
	return atomic.LoadUint32(&this.Status)&p2ptypes.STREAM_STATUS_NOT_HEALTHY == 0
}

