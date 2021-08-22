/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 8:59 上午
# @File : consensus.go
# @Description :
# @Attention :
*/
package config

var (
	COMPONENT_BLOCKCHAIN = NewModule("BLOCKCHAIN", 1)
)
var (
	SERVICE_MODULE_CONSENSUS = NewModule("CONSENSUS", 1)
)
var (
	LOGICSERVICE_MODULE_BLOCKCHAIN = NewModule("BLOCKCHAIN", 1)
	LOGICSERVICE_MODULE_CONSENSUS  = NewModule("CONSENSUS", 1)
)

var (
	PROTOCOL_CONSENSUS = "/consensus/1.0.0"
)
