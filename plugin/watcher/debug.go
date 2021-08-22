/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 12:53 下午
# @File : debug.go
# @Description :
# @Attention :
*/
package watcher

import (
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
)

// 用于调试
var (
	debug_async = true
	debug_print = false
	debugModule = modules.NewModule("debug", 1)
)

func debugPrint(msg string, kv ...interface{}) {
	if debug_print {
		logplugin.Warn(msg, kv...)
	}
}
