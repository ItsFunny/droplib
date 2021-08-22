/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/31 7:48 下午
# @File : demo.go
# @Description :
# @Attention :
*/
package mock

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services"
)

var (
	_ services.IChannelMessageWrapper = (*DemoC1ChannelMessageWrapper)(nil)
	_ services.IChannelMessageWrapper = (*DemoC2ChannelMessageWrapper)(nil)
)

func (this *DemoC1ChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &DemoC1ChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *DemoC1Message1Proto:
		r.Sum = &DemoC1ChannelDataMessageWrapper_DemoC1Message1{msg}
	case *DemoC1Message2Proto:
		r.Sum = &DemoC1ChannelDataMessageWrapper_DemoC1Message2{DemoC1Message2: msg}
	default:
		return nil, errors.New("未知类型")
	}
	return proto.Marshal(r)
}

func (this *DemoC1ChannelMessageWrapper) Decode(bytes []byte) (proto.Message,services.IMessage, error) {
	r := &DemoC1ChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil,nil, err
	}
	switch msg := r.Sum.(type) {
	case *DemoC1ChannelDataMessageWrapper_DemoC1Message1:
		return msg.DemoC1Message1,nil, nil
	case *DemoC1ChannelDataMessageWrapper_DemoC1Message2:
		return msg.DemoC1Message2,nil, nil
	default:
		return nil,nil, errors.New("asd")
	}
}

func (this *DemoC2ChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &DemoC2ChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *DemoC2Message1Proto:
		r.Sum = &DemoC2ChannelDataMessageWrapper_DemoC2Message1{msg}
	case *DemoC2Message2Proto:
		r.Sum = &DemoC2ChannelDataMessageWrapper_DemoC2Message2{DemoC2Message2: msg}
	default:
		return nil, errors.New("未知类型")
	}
	return proto.Marshal(r)
}

func (this *DemoC2ChannelMessageWrapper) Decode(bytes []byte) (proto.Message,services.IMessage, error) {
	r := &DemoC2ChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil,nil, err
	}
	switch msg := r.Sum.(type) {
	case *DemoC2ChannelDataMessageWrapper_DemoC2Message1:
		return msg.DemoC2Message1,nil, nil
	case *DemoC2ChannelDataMessageWrapper_DemoC2Message2:
		return msg.DemoC2Message2,nil, nil
	default:
		return nil,nil, errors.New("asd")
	}
}
