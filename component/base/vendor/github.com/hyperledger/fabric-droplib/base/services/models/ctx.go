/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/16 1:34 下午
# @File : ctx.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/hyperledger/fabric-droplib/base/services/constants"
)

type StartCTX struct {
	Flag constants.START_FLAG

	PreStart func()
	PostStart func()
}

type ReadyCTX struct {
	ReadyFlag constants.READY_FALG

	PreReady func()
	PostReady func()
}

type StopCTX struct {
}
