/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 10:00 上午
# @File : logrus_logger.go
# @Description :
# @Attention :
*/
package logplugin

var (
	_ ConcreteLogger=(*logrusLogger)(nil)
)
type logrusLogger struct {
	*CommonBaseLogger
}

func (l *logrusLogger) RecordInfo(first interface{}, info ...interface{}) {
}

func (l *logrusLogger) RecordDebug(first interface{}, info ...interface{}) {
	panic("implement me")
}

func (l *logrusLogger) RecordError(first interface{}, info ...interface{}) {
	panic("implement me")
}

func (l *logrusLogger) RecordWarn(first interface{}, info ...interface{}) {
	panic("implement me")
}

