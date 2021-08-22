/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/2/20 9:46 上午
# @File : log.go
# @Description :
# @Attention :
*/
package logplugin

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/log4go"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
)

// 全局log
var (
	logger v2.MLogger
)

func init() {
	logger = logrusplugin.NewGlobalLogrusLogger()
	// logger = NewLog4goLogger(NewCommonBaseLoggerWithLog4go("LOG", "ALL", nil))
	logcomponent.RegisterBlackList("log/log.go")
	logcomponent.RegisterBlackList("base/log")
	logcomponent.RegisterBlackList("base/base")
	// common.RegisterBlackList("impl/service_impl")
	log4go.SetModule("ALL")
}
func GlobalLogger(moduleName string) Logger {
	return NewLog4goLogger(NewCommonBaseLoggerWithLog4go("", moduleName, nil))
}
func GlobalLoggerWithDecorate(moduleName string, decorate func(str string) string) Logger {
	return NewLog4goLogger(NewCommonBaseLoggerWithLog4go("", moduleName, decorate))
}

func Info(msg string, kv ...interface{}) {
	logger.Info(msg, kv...)
}
func Debug(msg string, kv ...interface{}) {
	logger.Debug(msg, kv...)
}
func Warn(msg string, kv ...interface{}) {
	logger.Info(msg, kv...)
}
func InfoF(args ...interface{}) {
	logger.Info(fmt.Sprintf(args[0].(string), args[1:]...))
}

func Error(msg string, kv ...interface{}) {
	logger.Error(msg, kv...)
}

func With(fs map[string]interface{}) v2.Logger {
	return logger.With(fs)
}

func MInfo(m modules.Module, msg string, kv ...interface{}) {
	logger.MInfo(m, msg, kv...)
}

func MDebug(m modules.Module, msg string, kv ...interface{}) {
	logger.MDebug(m, msg, kv...)
}
func MWarn(m modules.Module, msg string, kv ...interface{}) {
	logger.MWarn(m, msg, kv...)
}
func MInfoF(m modules.Module, args ...interface{}) {
	logger.MInfof(m, fmt.Sprintf(args[0].(string), args[1:]...))
}

func MError(m modules.Module, msg string, kv ...interface{}) {
	logger.MError(m, msg, kv...)
}

func MErrorF(m modules.Module, msg string, kv ...interface{}) {
	logger.MErrorf(m, msg, kv...)
}
func MWith(m modules.Module, fs map[string]interface{}) v2.Logger {
	return logger.MWith(m, fs)
}
