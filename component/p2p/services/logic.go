/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/27 8:50 下午
# @File : logic.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/component/base"
	models2 "github.com/hyperledger/fabric-droplib/component/p2p/models"
)


type IPeerLifeCycleService interface {
	base.ILogicService
	Record() <-chan models2.PeerRecord
}
