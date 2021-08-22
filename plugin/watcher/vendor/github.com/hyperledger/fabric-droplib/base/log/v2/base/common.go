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
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
)

var (
	_       v2.Logger = (*CommonLogger)(nil)
	gModule           = modules.NewModule("GLOBAL", 1)
)

func init() {
	logcomponent.RegisterBlackList("base/common")
}

// 格式为  时间 + [LOG_LEVEL] + (模块) + (代码处) + msg + kv

type CommonLogger struct {
	module modules.Module
	ModuleLogger
}

type ModuleLogger struct {
	logger v2.IConcreteLogger
}

func NewCommonLogger(module modules.Module, logger v2.IConcreteLogger) *CommonLogger {
	r := &CommonLogger{
		module: module,
		ModuleLogger:ModuleLogger{logger: logger},
	}
	return r
}
func NewModuleCommonLogger(logger v2.IConcreteLogger) *ModuleLogger {
	r := &ModuleLogger{
		logger: logger,
	}
	return r
}

func (c CommonLogger) Info(msg string, keyvals ...interface{}) {
	c.ModuleLogger.MInfo(c.module, msg, keyvals...)
}
func (c CommonLogger) Infof(template string, keyvals ...interface{}) {
	c.ModuleLogger.MInfof(c.module, template, keyvals...)
}

func (c CommonLogger) Debug(msg string, keyvals ...interface{}) {
	c.ModuleLogger.MDebug(c.module, msg, keyvals...)
}

func (c CommonLogger) Panicf(msg string, keyvals ...interface{}) {
	c.ModuleLogger.MPanicf(c.module, msg, keyvals...)
}
func (c CommonLogger) Debugf(template string, keyvals ...interface{}) {
	c.ModuleLogger.MDebug(c.module, template, keyvals...)
}

func (c CommonLogger) Warn(msg string, keyvals ...interface{}) {
	c.ModuleLogger.MWarn(c.module, msg, keyvals...)
}

func (c CommonLogger) Warningf(template string, keyvals ...interface{}) {
	c.ModuleLogger.MWarningf(c.module, template, keyvals...)
}

func (c CommonLogger) Error(msg string, keyvals ...interface{}) {
	c.ModuleLogger.MError(c.module, msg, keyvals...)
}
func (c CommonLogger) Fatal(args ...interface{}) {
	// return c.logger.CFalta(c.module,)
}

func (c CommonLogger) Errorf(template string, keyvals ...interface{}) {
	c.ModuleLogger.MErrorf(c.module, template, keyvals...)
}
func (c CommonLogger) With(fileds map[string]interface{}) v2.Logger {
	return c.ModuleLogger.logger.CWith(c.module, fileds)
}

func (c CommonLogger) Fatalf(format string, args ...interface{}) {
	panic("implement me")
}

func (c CommonLogger) Panic(args ...interface{}) {
	panic("implement me")
}

func (c CommonLogger) Warnf(format string, args ...interface{}) {
	panic("implement me")
}

func (c ModuleLogger) mlog(m modules.Module, l common.Level, msg string, keyvals ...interface{}) {
	if logcomponent.IsLogLevelDisabled(l, m.String()) {
		return
	}
	var line interface{}
	lineStr, ok := GetCodeLineNumber()
	if ok {
		line = lineStr
	}
	switch l {
	case common.DebugLevel:
		c.logger.CDebug(m.String(), line, msg, keyvals...)
	case common.InfoLevel:
		c.logger.CInfo(m.String(), line, msg, keyvals...)
	case common.WarnLevel:
		c.logger.CWarn(m.String(), line, msg, keyvals...)
	case common.ErrorLevel:
		c.logger.CError(m.String(), line, msg, keyvals...)
	default:
		c.logger.CInfo(m.String(), line, msg, keyvals...)
	}
}

func (c ModuleLogger) MInfo(m modules.Module, msg string, keyvals ...interface{}) {
	c.mlog(m, common.InfoLevel, msg, keyvals...)
}

func (m2 ModuleLogger) MPanicf(m modules.Module, msg string, keyvals ...interface{}) {
	panic(fmt.Sprintf(msg, keyvals...))
}
func (c ModuleLogger) MInfof(m modules.Module, template string, keyvals ...interface{}) {
	c.mlog(m, common.InfoLevel, fmt.Sprintf(template, keyvals...))
}

func (c ModuleLogger) MDebug(m modules.Module, msg string, keyvals ...interface{}) {
	c.mlog(m, common.DebugLevel, msg, keyvals...)
}

func (c ModuleLogger) MDebugf(m modules.Module, template string, keyvals ...interface{}) {
	c.mlog(m, common.DebugLevel, fmt.Sprintf(template, keyvals...))
}


func (c ModuleLogger) MWarn(m modules.Module, msg string, keyvals ...interface{}) {
	c.mlog(m, common.WarnLevel, msg, keyvals...)
}



func (c ModuleLogger) MWarningf(m modules.Module, template string, keyvals ...interface{}) {
	c.mlog(m, common.WarnLevel, fmt.Sprintf(template, keyvals...))
}

func (c ModuleLogger) MError(m modules.Module, msg string, keyvals ...interface{}) {
	c.mlog(m, common.ErrorLevel, msg, keyvals...)
}


func (c ModuleLogger) MErrorf(m modules.Module, template string, keyvals ...interface{}) {
	c.mlog(m, common.ErrorLevel, fmt.Sprintf(template, keyvals...))
}

func (c ModuleLogger) MWith(m modules.Module, fileds map[string]interface{}) v2.Logger {
	return c.logger.CWith(m, fileds)
}

func GetCodeLineNumber() (string, bool) {
	return logcomponent.FindCaller(3)
}
//
// func (c ModuleLogger) Info(msg string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.InfoLevel, msg, keyvals...)
// }
// func (m2 ModuleLogger) Panicf(msg string, keyvals ...interface{}) {
// 	panic(fmt.Sprintf(msg, keyvals...))
// }
// func (c ModuleLogger) Infof(template string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.InfoLevel, fmt.Sprintf(template, keyvals...))
// }
// func (c ModuleLogger) Debug(msg string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.DebugLevel, msg, keyvals...)
// }
// func (c ModuleLogger) Debugf(template string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.DebugLevel, fmt.Sprintf(template, keyvals...))
// }
// func (c ModuleLogger) Warn(msg string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.WarnLevel, msg, keyvals...)
// }
// func (c ModuleLogger) Warningf(template string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.WarnLevel, fmt.Sprintf(template, keyvals...))
// }
// func (c ModuleLogger) Error(msg string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.ErrorLevel, msg, keyvals...)
// }
// func (c ModuleLogger) Errorf(template string, keyvals ...interface{}) {
// 	c.mlog(gModule, common.ErrorLevel, fmt.Sprintf(template, keyvals...))
// }
// func (c ModuleLogger) With(fileds map[string]interface{}) v2.Logger {
// 	return c.logger.CWith(gModule, fileds)
// }
