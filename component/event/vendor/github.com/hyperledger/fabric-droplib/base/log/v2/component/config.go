/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/9 9:05 下午
# @File : config.go
# @Description :
# @Attention :
*/
package logcomponent

import (
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"os"
	"strings"
)

// FIXME
const (
	ENV_LOG_BLOCK_MODULE  = "BLOCK_MODULE"
	ENV_DEFAULT_LOG_LEVEL = "DEFAULT_LOG_LEVEL"
)

// fixme
func InitLog() {
	loglevl := os.Getenv(ENV_DEFAULT_LOG_LEVEL)
	defaultLogLevel := common.DebugLevel
	if len(loglevl) != 0 {
		loglevl = strings.ToLower(loglevl)
		switch loglevl {
		case "debug":
			defaultLogLevel = common.DebugLevel
		case "info":
			defaultLogLevel = common.InfoLevel
		case "warn":
			defaultLogLevel = common.WarnLevel
		case "error":
			defaultLogLevel = common.ErrorLevel

		}
	}
	blockModule := os.Getenv(ENV_LOG_BLOCK_MODULE)
	if len(blockModule) > 0 {
		blocks := strings.Split(blockModule, ",")
		for index, block := range blocks {
			blocks[index] = strings.ToUpper(block)
		}
		RegisterBlackModule(blocks...)
	}
	SetGlobalLogLevel(defaultLogLevel)
	RegisterBlackList("go-log/log.go")
}
