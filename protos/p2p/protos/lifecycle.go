/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/29 8:45 上午
# @File : pingpong.go
# @Description :
# @Attention :
*/
package protos

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services"
)

var (
	_ services.IChannelMessageWrapper = (*LifeCycleDefaultChannelMessageWrapper)(nil)
)

func (this *LifeCycleDefaultChannelMessageWrapper) Decode(bytes []byte) (proto.Message,services.IMessage, error) {
	r := &LifeCycleDefaultChannelDataMessageWrapper{}
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil,nil, err
	}
	switch msgWp := r.GetSum().(type) {
	case *LifeCycleDefaultChannelDataMessageWrapper_MemberInfoBroadcastMessageProto:
		return msgWp.MemberInfoBroadcastMessageProto,nil, nil
	case *LifeCycleDefaultChannelDataMessageWrapper_PingMessageProto:
		return msgWp.PingMessageProto,nil, nil
	case *LifeCycleDefaultChannelDataMessageWrapper_PongMessageProto:
		return msgWp.PongMessageProto,nil, nil
	case *LifeCycleDefaultChannelDataMessageWrapper_BroadcastMessageProto:
		return msgWp.BroadcastMessageProto,nil, nil
	default:
		return nil,nil, fmt.Errorf("未知的数据类型: %T", msgWp)
	}
}

func (this *LifeCycleDefaultChannelMessageWrapper) Encode(message proto.Message) ([]byte, error) {
	r := &LifeCycleDefaultChannelDataMessageWrapper{}
	switch msg := message.(type) {
	case *MemberInfoBroadCastMessageProto:
		r.Sum = &LifeCycleDefaultChannelDataMessageWrapper_MemberInfoBroadcastMessageProto{
			MemberInfoBroadcastMessageProto: msg,
		}
	case *PingPongBroadCastMessageProto:
		r.Sum = &LifeCycleDefaultChannelDataMessageWrapper_BroadcastMessageProto{
			BroadcastMessageProto: msg,
		}
	case *PingMessageProto:
		r.Sum = &LifeCycleDefaultChannelDataMessageWrapper_PingMessageProto{
			PingMessageProto: msg,
		}
	case *PongMessageProto:
		r.Sum = &LifeCycleDefaultChannelDataMessageWrapper_PongMessageProto{PongMessageProto: msg}
	default:
		return nil, fmt.Errorf("未知的 message类型: %T", msg)
	}
	return proto.Marshal(r)
}
