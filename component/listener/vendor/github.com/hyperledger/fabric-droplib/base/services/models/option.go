/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/16 1:45 下午
# @File : option.go
# @Description :
# @Attention :
*/
package models

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/services/constants"
)

type StartOption func(c *StartCTX)
type ReadyOption func(c *ReadyCTX)
type StopOption func(c *StopCTX)

func CtxStartOpt(ctx context.Context) StartOption {
	return func(c *StartCTX) {
		c.Ctx = ctx
	}
}
func PreStartOpt(f func()) StartOption {
	return func(c *StartCTX) {
		c.PreStart = f
	}
}
func PostStartOpt(f func()) StartOption {
	return func(c *StartCTX) {
		c.PostStart = f
	}
}
func AsyncStartWaitReadyOpt(c *StartCTX) {
	c.Flag = constants.ASYNC_START_WAIT_READY
}
func SyncStartWaitReadyOpt(c *StartCTX) {
	c.Flag = constants.SYNC_START_WAIT_READY
}
func SyncStartOpt(c *StartCTX) {
	c.Flag = constants.SYNC_START
}
func AsyncStartOpt(c *StartCTX) {
	c.Flag = constants.ASYNC_START
}

func ReadyWaitStartOpt(c *ReadyCTX) {
	c.ReadyFlag = constants.READY_UNTIL_START
}
func ReadyPanicIfErrOpt(c *ReadyCTX) {
	c.ReadyFlag = constants.READY_ERROR_IF_NOT_STARTED
}
func PreReadyOpt(f func()) ReadyOption {
	return func(c *ReadyCTX) {
		c.PreReady = f
	}
}
func PostReadyOpt(f func()) ReadyOption {
	return func(c *ReadyCTX) {
		c.PostReady = f
	}
}
