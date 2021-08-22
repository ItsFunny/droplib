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
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type LogConfiguration struct {
	// FIXME STATUS
	status         int32
	blackList      []string
	lock           sync.Mutex
	LogLevel       common.Level
	blackModuleSet map[string]struct{}
	moduleLevel    map[string]common.Level
	// loggers map[string]v2.Logger
}

// var (
// 	logConfig *LogConfiguration
// )
//
// func init() {
// 	logConfig = NewLogConfiguration()
// }

func NewLogConfiguration() *LogConfiguration {
	r := &LogConfiguration{
		blackList:      []string{"log/base_logger"},
		LogLevel:       common.InfoLevel,
		blackModuleSet: make(map[string]struct{}, 1),
		moduleLevel:    make(map[string]common.Level, 1),
		// loggers: make(map[string]v2.Logger),
	}
	return r
}

func (this *LogConfiguration) IsLogLevelDisabled(level common.Level, moduleName string) bool {
	if this.LogLevel > level {
		return true
	}
	_, exist := this.blackModuleSet[moduleName]
	return exist
}
func (this *LogConfiguration) FindCaller(skip int) (string, bool) {
	return common.FindCaller(skip, this.blackList)
}
func (this *LogConfiguration) RegisterBlackList(pathes ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, path := range pathes {
		this.blackList = append(this.blackList, path)
	}
}
func (this *LogConfiguration) RegisterBlackModule(modules ...string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, module := range modules {
		module = strings.ToUpper(module)
		this.blackModuleSet[module] = struct{}{}
	}
}
func (this *LogConfiguration) RegisterModuleLevel(m map[string]common.Level) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for module, l := range m {
		module = strings.ToUpper(module)
		this.moduleLevel[module] = l
	}
}

func (this *LogConfiguration) SetGlobalLogLevel(l common.Level) {
	this.LogLevel = l
	logrus.SetLevel(l.GetLogrusLevel())
}

func (this *LogConfiguration) GetModuleLevel(m string) common.Level {
	level, exist := this.moduleLevel[m]
	if exist {
		return level
	}
	return this.LogLevel
}
