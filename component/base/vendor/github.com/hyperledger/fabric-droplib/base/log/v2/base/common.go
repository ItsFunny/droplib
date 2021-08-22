/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 10:06 上午
# @File : common.go
# @Description :
# @Attention :
*/
package base

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"github.com/hyperledger/fabric-droplib/base/log/config"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
)

var (
	_ v2.Logger = (*CommonLogger)(nil)
)

func init() {
	config.RegisterBlackList("base/common")
}

// 格式为  时间 + [LOG_LEVEL] + (模块) + (代码处) + msg + kv

type CommonLogger struct {
	module modules.Module

	logger v2.IConcreteLogger
}

func NewCommonLogger(module modules.Module, logger v2.IConcreteLogger) *CommonLogger {
	r := &CommonLogger{
		module: module,
		logger: logger,
	}
	return r
}

func (c CommonLogger) Info(msg string, keyvals ...interface{}) {
	c.log(common.InfoLevel, msg, keyvals...)
}
func (c CommonLogger) Infof(template string, keyvals ...interface{}) {
	c.log(common.InfoLevel, fmt.Sprintf(template, keyvals...))
}

func (c CommonLogger) Debug(msg string, keyvals ...interface{}) {
	c.log(common.DebugLevel, msg, keyvals...)
}

func (c CommonLogger) Panicf(msg string, keyvals ...interface{}) {
	panic(fmt.Sprintf(msg, keyvals...))
}
func (c CommonLogger) Debugf(template string, keyvals ...interface{}) {
	c.log(common.DebugLevel, fmt.Sprintf(template, keyvals...))
}

func (c CommonLogger) Warn(msg string, keyvals ...interface{}) {
	c.log(common.WarnLevel, msg, keyvals...)
}

func (c CommonLogger) Warningf(template string, keyvals ...interface{}) {
	c.log(common.WarnLevel, fmt.Sprintf(template, keyvals...))
}

func (c CommonLogger) Error(msg string, keyvals ...interface{}) {
	c.log(common.ErrorLevel, msg, keyvals...)
}
func (c CommonLogger) Errorf(template string, keyvals ...interface{}) {
	c.log(common.ErrorLevel, fmt.Sprintf(template, keyvals...))
}
func (c CommonLogger) With(fileds map[string]interface{}) v2.Logger {
	return c.logger.CWith(c.module, fileds)
}

func (c CommonLogger) log(l common.Level, msg string, keyvals ...interface{}) {
	if config.IsLogLevelDisabled(l, c.module.String()) {
		return
	}
	var line interface{}
	lineStr, ok := c.GetCodeLineNumber()
	if ok {
		line = lineStr
	}
	switch l {
	case common.DebugLevel:
		c.logger.CDebug(c.module.String(), line, msg, keyvals...)
	case common.InfoLevel:
		c.logger.CInfo(c.module.String(), line, msg, keyvals...)
	case common.WarnLevel:
		c.logger.CWarn(c.module.String(), line, msg, keyvals...)
	case common.ErrorLevel:
		c.logger.CError(c.module.String(), line, msg, keyvals...)
	default:
		c.logger.CInfo(c.module.String(), line, msg, keyvals...)
	}
}

func (c CommonLogger) GetCodeLineNumber() (string, bool) {
	return config.FindCaller(3)
}
