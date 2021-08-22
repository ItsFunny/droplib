/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 10:03 上午
# @File : logger.go
# @Description :
# @Attention :
*/
package v2

import "github.com/hyperledger/fabric-droplib/base/log/modules"


type Logger interface {
	Info(msg string, keyvals ...interface{})
	Panicf(msg string, keyvals ...interface{})
	Infof(template string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Debugf(template string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Warningf(template string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Errorf(template string, keyvals ...interface{})
	With(fileds map[string]interface{}) Logger

	// Fatal(args ...interface{})
	// Fatalf(format string, args ...interface{})
	// Panic(args ...interface{})
	// Warnf(format string, args ...interface{})
}

type MLogger interface {
	Logger
	MInfo(m modules.Module,msg string, keyvals ...interface{})
	MPanicf(m modules.Module,msg string, keyvals ...interface{})
	MInfof(m modules.Module,template string, keyvals ...interface{})
	MDebug(m modules.Module,msg string, keyvals ...interface{})
	MDebugf(m modules.Module,template string, keyvals ...interface{})
	MWarn(m modules.Module,msg string, keyvals ...interface{})
	MWarningf(m modules.Module,template string, keyvals ...interface{})
	MError(m modules.Module,msg string, keyvals ...interface{})
	MErrorf(m modules.Module,template string, keyvals ...interface{})
	MWith(m modules.Module,fileds map[string]interface{}) Logger
}

type IConcreteLogger interface {
	CDebug(module string, lineNo interface{}, msg string, keyvals ...interface{})
	CInfo(module string, lineNo interface{}, msg string, keyvals ...interface{})
	CWarn(module string, lineNo interface{}, msg string, keyvals ...interface{})
	CError(module string, lineNo interface{}, msg string, keyvals ...interface{})
	CFalta(module string, lineNo interface{}, msg string, keyvals ...interface{})
	CWith(module modules.Module, fields map[string]interface{}) Logger
}
