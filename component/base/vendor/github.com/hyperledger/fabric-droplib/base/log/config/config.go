/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/14 11:10 上午
# @File : config.go
# @Description :
# @Attention :
*/
package config

import (
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"sync"
)

type LogConfiguration struct {
	// FIXME STATUS
	status byte
	blackList      []string
	lock           sync.Mutex
	LogLevel       common.Level
	blackModuleSet map[string]struct{}
	moduleLevel    map[string]common.Level
}

var (
	logConfig *LogConfiguration
)

func init() {
	logConfig = &LogConfiguration{
		blackList:      []string{"log/base_logger"},
		LogLevel:       common.InfoLevel,
		blackModuleSet: make(map[string]struct{}, 1),
		moduleLevel:    make(map[string]common.Level, 1),
	}
}

func (this *LogConfiguration) GetModuleLevel(m string) common.Level {
	level, exist := this.moduleLevel[m]
	if exist {
		return level
	}
	return this.LogLevel
}

