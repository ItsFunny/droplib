/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:27 上午
# @File : consensus.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
)

// type IConsensusReactor interface {
// 	base.ILogicService
// 	BeginPropose(txs types.Txs)
//
// 	ChoseInterestEvents(bus services.IP2PEventBusComponent)
// }

type IConsensusLogicService interface {
	base.ILogicService

}


type IConsensusReactor interface {
	services.IReactor

}

