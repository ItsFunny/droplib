/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 3:57 下午
# @File : logic.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/types"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/component/stream/services"
)

type ILogicComponent interface {
	base.IComponent
	GetBoundServices() []ILogicService
}

type Choser interface {
	AcquireEventChanByNameSpace(data interface{}) chan<- interface{}
}
type ILogicService interface {
	services2.IBaseService
	GetServiceId() types.ServiceID
	ChoseInterestProtocolsOrTopics(pProtocol *models.P2PProtocol) error
	SetChannelMessagesWrapper(r map[types.ChannelID]services.IChannelMessageWrapper)
	// 最好是async
	ChoseInterestEvents(Choser)
}

type BaseLogicComponent struct {
	*base.BaseComponent
}

func NewBaseLogicCompoennt(m modules.Module, i services2.IBaseService) *BaseLogicComponent {
	r := &BaseLogicComponent{}
	r.BaseComponent = base.NewBaseComponent(m, i)

	return r
}
