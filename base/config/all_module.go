/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 2:25 下午
# @File : module.go
# @Description :
# @Attention :
*/
package config

var (
	MODULE_EVENT_BUS        = NewModule("EVENT_BUS", 1)
	MODULE_COMMON_EVENT_BUS = NewModule("COMMON_EVENTBUS", 1)
	MODULE_PUBSUB           = NewModule("PUBSUB", 1)

	MODULE_MEM_POOL = NewModule("MEM_POOL", 2)

	MODULE_BLOCK_STORE = NewModule("BLOCK_STORE", 3)

	MODULE_STATE          = NewModule("STATE", 4)
	MODULE_TIMEOUT_TICKER = NewModule("TIMEOUT_TICKER", 5)

	MODULE_WAL = NewModule("WAL", 6)

	MODULE_EVENT_MANAGER = NewModule("EVENT_MANAGER", 7)

	MODULE_P2P = NewModule("P2P", 8)

	MODULE_ROUTER = NewModule("ROUTER", 9)

	MODULE_STREAM_REGISTRY = NewModule("REGISTRY", 10)

	MODULE_AUTO_FILE_GROUP = NewModule("AUTO_FILE_GROUP", 11)
)
var (
	MODULE_BLOCK_EXECUTOR = NewModule("BLOCK_EXECUTOR", 12)
)
var (
	MODULE_PROTOCOL_LIFECYCLE = NewModule("PROTOCOL_LIFECYCLE", 1)
	MODULE_PEER_LIFECYCLE     = NewModule("PEER_LIFECYCLE", 2)
)
