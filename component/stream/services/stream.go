/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 7:05 下午
# @File : stream.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/types"
	"io"
)


type IStream interface {
	io.Reader
	io.Writer
	// Status()uint32 // better atomic
	// Id()int

	Close() error
	CloseWrite()error
	CloseRead() error
	Reset() error
}



type IStreamObject interface {
	GetData() []byte
	SetData([]byte)
	GetChanelID() types.ChannelID
	SetChannelID(id types.ChannelID)
	GetStreamFlagInfo() int32
	SetStreamFlagInfo(flag int32)
	ToBytes() ([]byte, error)
}

type Decorator func(s IStreamObject)

// FIXME: 这个接口设计的很差
type IProtocolWrapper interface {
	Encode(data []byte, decorators ...Decorator) IStreamObject
	Decode(bytes []byte) (IStreamObject, error)
}

type IChannelMessageWrapper interface {
	proto.Message
	Encode(message proto.Message) ([]byte, error)
	Decode(bytes []byte) (proto.Message,services.IMessage, error)
}

type IMessageAble interface {
	ToMessage() services.IMessage
}

// 这个必须在protobuf中定义
type IDataWrapper interface {
	proto.Message
	GetInternalData() interface{}
}

// 消息编码
// type Encoder interface {
// 	Encode(proto.Message) (IDataWrapper, error)
// }

// 消息解码
// type Decoder interface {
// 	Decode(IDataWrapper) (proto.Message, error)
// }
