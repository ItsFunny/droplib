/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 11:29 下午
# @File : switch.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
)

type IReactor interface {
	services.IBaseService

	GetReactorId() types.ReactorId
	SetSwitch(iSwitch ISwitch)
}

type ISwitch interface {
	services.IBaseService
	SwitchTo(id types.ReactorId) IReactor

	// panic if error
	RegisterReactor(reactor IReactor)
}
