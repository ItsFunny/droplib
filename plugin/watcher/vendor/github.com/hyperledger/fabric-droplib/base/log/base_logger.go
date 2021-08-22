/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2019-12-14 12:24
# @File : logger.go
# @Description :
# @Attention :
*/
package logplugin

type CommonBaseLogger struct {
	ReqID  string
	Prefix string
	Module string

	Decorate func(str string) string

	ConcreteLogger ConcreteLogger
}

// func NewCommonBaseLogger(reqID string) *CommonBaseLogger {
// 	l := new(CommonBaseLogger)
//
// 	goLogger := NewLog4goLogger(l)
// 	l.ConcreteLogger = goLogger
// 	l.ReqID = reqID
//
// 	return l
// }

func NewCommonBaseLoggerWithLog4go(reqID string, moduleName string, decorate func(str string) string) *CommonBaseLogger {
	l := new(CommonBaseLogger)

	goLogger := NewLog4goLogger(l)
	l.ConcreteLogger = goLogger
	l.ReqID = reqID
	l.Module = moduleName
	l.Decorate = decorate

	return l
}

func (l *CommonBaseLogger) Info(first interface{}, info ...interface{}) {
	l.ConcreteLogger.RecordInfo(first, info...)
}

func (l *CommonBaseLogger) Warn(first interface{}, info ...interface{}) {
	l.ConcreteLogger.RecordInfo(first, info...)
}

func (l *CommonBaseLogger) Debug(first interface{}, info ...interface{}) {
	l.ConcreteLogger.RecordDebug(first, info...)
}
func (l *CommonBaseLogger) MDebug(module string, first interface{}, info ...interface{}) {
	l.ConcreteLogger.RecordDebug(first, info...)
}
func (l *CommonBaseLogger) Error(first interface{}, info ...interface{}) {
	l.ConcreteLogger.RecordError(first, info...)
}

func (l *CommonBaseLogger) SetPrefix(p string) {
	l.Prefix = p
}

func (l *CommonBaseLogger) GetPrefix() string {
	return l.Prefix
}

func (l *CommonBaseLogger) SetReqID(r string) {
	l.ReqID = r
}

func (l *CommonBaseLogger) GetReqID() string {
	return l.ReqID
}

// FIXME
func (l *CommonBaseLogger) New(module string) Logger {
	cb := NewCommonBaseLoggerWithLog4go("", module, nil)
	cb.Decorate = l.Decorate
	cb.Module = module
	return NewLog4goLogger(cb)
}
