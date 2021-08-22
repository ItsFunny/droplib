/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/29 12:57 下午
# @File : base_logic_impl.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	"strings"
)

type ChannelInfo struct {
}
type BaseLogicServiceImpl struct {
	*impl.BaseServiceImpl
	channelMessages map[types.ChannelID]services2.IChannelMessageWrapper
}

func NewBaseLogicServiceImpl(module modules.Module, concrete services2.IBaseService) *BaseLogicServiceImpl {
	r := &BaseLogicServiceImpl{}
	name := module.String()
	name = strings.ToUpper(name)
	if !strings.Contains(name, "LOGIC_SERVICE") {
		name = name + "_LOGIC_SERVICE"
		module = modules.NewModule(name, module.Index())
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, module, concrete)
	return r
}
func (b *BaseLogicServiceImpl) SetChannelMessagesWrapper(r map[types.ChannelID]services2.IChannelMessageWrapper) {
	b.channelMessages = r
}
func (b *BaseLogicServiceImpl) UnwrapFromBytes(cid types.ChannelID, data []byte) (proto.Message, services2.IMessage, error) {
	return b.channelMessages[cid].Decode(data)
}
func (b *BaseLogicServiceImpl) Wrap2Bytes(cid types.ChannelID, msg proto.Message) ([]byte, error) {
	return b.channelMessages[cid].Encode(msg)
}
func (b *BaseLogicServiceImpl) Pre() bool {
	if !b.IsRunning() {
		return false
	}
	return true
}
