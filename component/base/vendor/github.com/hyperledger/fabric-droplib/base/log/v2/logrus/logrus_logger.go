/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 3:13 下午
# @File : logrus_logger.go
# @Description :
# @Attention :
*/
package logrusplugin

import (
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	"github.com/hyperledger/fabric-droplib/base/log/v2/base"
	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	v2.Logger
	log    *logrus.Logger
	fields map[string]interface{}
}

func NewLogrusLogger(module modules.Module) v2.Logger {
	return newLogrus(module)
}
func newLogrus(module modules.Module) *logrusLogger {
	r := &logrusLogger{}
	r.Logger = base.NewCommonLogger(module, r)
	r.log = logrus.New()
	r.log.SetFormatter(NewTextFormmater())
	var ll logrus.Level
	l:=module.LogLevel()
	if l==common.DebugLevel{
		ll=logrus.DebugLevel
	}else if l==common.InfoLevel{
		ll=logrus.InfoLevel
	}else if l==common.ErrorLevel{
		ll=logrus.ErrorLevel
	}else if l==common.FatalLevel{
		ll=logrus.FatalLevel
	}else{
		ll=logrus.InfoLevel
	}
	r.log.SetLevel(ll)
	return r
}

func (l *logrusLogger) CInfo(module string, lineNo interface{}, msg string, keyvals ...interface{}) {
	fields, more := l.buildFields(module, lineNo, keyvals...)
	if nil != more {
		l.log.WithFields(fields).Info(msg, more)
		return
	}
	l.log.WithFields(fields).Info(msg)
}

func (l *logrusLogger) CDebug(module string, lineNo interface{}, msg string, keyvals ...interface{}) {
	fields, more := l.buildFields(module, lineNo, keyvals...)

	if nil != more {
		l.log.WithFields(fields).Debug(msg, more)
		return
	}
	l.log.WithFields(fields).Debug(msg)
}

func (l *logrusLogger) CWarn(module string, lineNo interface{}, msg string, keyvals ...interface{}) {
	fields, more := l.buildFields(module, lineNo, keyvals...)
	if nil != more {
		l.log.WithFields(fields).Warn(msg, more)
		return
	}
	l.log.WithFields(fields).Warn(msg)
}

func (l *logrusLogger) CError(module string, lineNo interface{}, msg string, keyvals ...interface{}) {
	fields, more := l.buildFields(module, lineNo, keyvals...)
	if nil != more {
		l.log.WithFields(fields).Error(msg, more)
		return
	}
	l.log.WithFields(fields).Error(msg)
}

func (l *logrusLogger) CWith(module modules.Module, fields map[string]interface{}) v2.Logger {
	prevF := l.fields
	if nil != prevF && nil != fields {
		for k, v := range prevF {
			fields[k] = v
		}
	}
	l2 := newLogrus(module)
	l2.fields = fields
	return l2
}

func (l *logrusLogger) buildFields(module string, lineNo interface{}, keyvals ...interface{}) (res logrus.Fields, more interface{}) {
	res = make(map[string]interface{})
	res[MODULE] = module
	res[CODE_LINE_NUMBER] = lineNo
	ll := len(keyvals)
	if len(keyvals)&1 != 0 {
		ll -= 1
		more = keyvals[len(keyvals)-1]
	}
	for i := 0; i < ll; i += 2 {
		if k, ok := keyvals[i].(string); ok {
			res[k] = keyvals[i+1]
		}
	}
	if l.fields != nil {
		for k, v := range l.fields {
			res[k] = v
		}
	}

	return
}
