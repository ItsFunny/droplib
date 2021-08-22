/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/16 1:35 下午
# @File : base.go
# @Description :
# @Attention :
*/
package constants

type START_FLAG int8
type READY_FALG int8

func (this START_FLAG) Sync() bool {
	return int8(this)&SYNC >= SYNC
}
func (this READY_FALG) Sync() bool {
	return int8(this)&SYNC >= SYNC
}

var (
	SYNC  int8 = 1 << 0
	ASYNC int8 = 1 << 1
)
var (
	SYNC_START                        = START_FLAG(1<<0 | SYNC)
	ASYNC_START                       = START_FLAG(1<<1 | ASYNC)
	WAIT_READY             START_FLAG = 1 << 2
	ASYNC_START_WAIT_READY            = ASYNC_START | WAIT_READY
	SYNC_START_WAIT_READY             = SYNC_START | WAIT_READY
)

var (
	READY_ERROR_IF_NOT_STARTED = READY_FALG(1 << 2)
	READY_UNTIL_START          = READY_FALG(1 << 3)
	SYNC_READY_UNTIL_START     = READY_UNTIL_START | READY_FALG(SYNC)
	ASYNC_READY_UTIL_START     = READY_UNTIL_START | READY_FALG(ASYNC)
)

const (
	NONE     = 0
	STARTED  = 1 << 0
	ON_READY = 1<<1 | STARTED
	READY    = 1<<2 | ON_READY

	STOP  = 1 << 6
	FLUSH = 1<<7 | STOP
)
