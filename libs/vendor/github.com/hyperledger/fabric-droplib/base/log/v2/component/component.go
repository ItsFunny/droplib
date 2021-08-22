/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/2 12:40 下午
# @File : component.go
# @Description :
# @Attention :
*/
package logcomponent

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"github.com/hyperledger/fabric-droplib/base/log/config"
	"sync/atomic"
	"time"
)

var logManager *LogManager

func init() {
	logManager = new(LogManager)
	logManager.cfg = config.NewLogConfiguration()
	RegisterBlackList("component/component.go")
}

type LogManager struct {
	status int32
	cfg    *config.LogConfiguration
}

func GetLogConfig() *config.LogConfiguration {
	return logManager.cfg
}
// FIXME FILTER
// 只有比配置大的时候打印
func IsLogLevelDisabled(level common.Level, moduleName string) bool {
	return logManager.cfg.IsLogLevelDisabled(level, moduleName)
}
func FindCaller(skip int) (string, bool) {
	return logManager.cfg.FindCaller(skip)
}

func GetModuleLevel(m string) common.Level {
	return logManager.cfg.GetModuleLevel(m)
}

func RegisterBlackList(pathes ...string) {
	logManager.cfg.RegisterBlackList(pathes...)
}

func RegisterBlackModule(modules ...string) {
	logManager.cfg.RegisterBlackModule(modules...)
}

func RegisterModuleLevel(m map[string]common.Level) {
	logManager.cfg.RegisterModuleLevel(m)
}

func SetGlobalLogLevel(l common.Level) {
	logManager.cfg.SetGlobalLogLevel(l)
}

func NotifyAsReady() {
	atomic.CompareAndSwapInt32(&logManager.status, common.NONE, common.READY)
}

func WaitUntilReady(name string) {
	c := func() bool {
		return logManager.status == common.READY
	}
	for !c() {
		fmt.Println(name+",log manager 未就绪,阻塞中, 唤醒需要调用 NotifyAsReady")
		time.Sleep(time.Second * 1)
	}
}
