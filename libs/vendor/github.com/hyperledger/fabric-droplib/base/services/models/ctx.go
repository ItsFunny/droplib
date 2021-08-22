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
	"context"
	"github.com/hyperledger/fabric-droplib/base/services/constants"
)

// FIXME ,错误位置
type StartCTX struct {
	Ctx context.Context

	Flag      constants.START_FLAG
	PreStart  func()
	PostStart func()
}

type ReadyCTX struct {
	Ctx context.Context

	ReadyFlag constants.READY_FALG

	PreReady  func()
	PostReady func()
}

type StopCTX struct {
	Force bool
	Value map[string]interface{}
}

func (this *StartCTX) GetValue(key string) interface{} {
	if nil == this.Ctx {
		return nil
	}
	return this.Ctx.Value(key)
}
