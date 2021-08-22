/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/25 8:42 下午
# @File : message.go
# @Description :
# @Attention :
*/
package models

import (
	"encoding/binary"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
)

type StreamMessageHeader struct {
	DataSize  uint64
	IdleBytes []byte
}

func (this StreamMessageHeader) ToBytes() []byte {
	bigI := make([]byte, 8)
	binary.BigEndian.PutUint64(bigI, this.DataSize)
	leftBytes := make([]byte, p2ptypes.STREAM_MESSAGE_IDLE_BYTES_SIZE)
	if len(this.IdleBytes) > p2ptypes.STREAM_MESSAGE_IDLE_BYTES_SIZE {
		copy(leftBytes, this.IdleBytes[0:p2ptypes.STREAM_MESSAGE_IDLE_BYTES_SIZE])
	}
	return append(bigI, leftBytes...)
}

type StreamMessage struct {
	Header StreamMessageHeader
	Data   []byte
}

func (this StreamMessage) ToBytes() []byte {
	l := len(this.Data)
	res := make([]byte, l+p2ptypes.STREAM_MESSAGE_HEADER_SIZE)
	res = append(res, this.Header.ToBytes()...)
	copy(res, this.Header.ToBytes())
	copy(res[p2ptypes.STREAM_MESSAGE_SIZE_END:], this.Data)
	return res
}
