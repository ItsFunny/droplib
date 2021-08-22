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

var (
	SYNC_START             START_FLAG = 1 << 0
	ASYNC_START            START_FLAG = 1 << 1
	WAIT_READY             START_FLAG = 1 << 2
	ASYNC_START_WAIT_READY            = ASYNC_START | WAIT_READY
	SYNC_START_WAIT_READY             = SYNC_START | WAIT_READY
)

var (
	READY_ERROR_IF_NOT_STARTED READY_FALG = 1 << 1
	READY_UNTIL_START          READY_FALG = 1 << 2
)

const (
	NONE    = 0
	STARTED = 1 << 0
	READY   = 1<<1 | STARTED

	STOP  = 1
	FLUSH = 1<<1 | STOP
)
