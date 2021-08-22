/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2019-12-14 12:33
# @File : log4gologger.go
# @Description :
# @Attention :
*/
package logplugin

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/log4go"
	"runtime"
	"strings"
	"sync"
)

type log4goLogger struct {
	*CommonBaseLogger
}

func NewLog4goLogger(b *CommonBaseLogger) *log4goLogger {
	l := new(log4goLogger)
	l.CommonBaseLogger = b
	return l
}

var (
	blackList = []string{"log/base_logger"}
	lock      sync.Mutex
)

// 废弃
func RegisterBlackListaaa(path string) {
	lock.Lock()
	defer lock.Unlock()
	blackList = append(blackList, path)
}

func findCaller(skip int) (string, bool) {
	file := ""
	line := 0
	ok := false
	sp := false
	for i := 0; i < 10; i++ {
		file, line, ok = getCaller(skip + i) //
		if !ok {
			return "", false
		}
		for _, bl := range blackList {
			if strings.HasPrefix(file, bl) {
				sp = true
				break
			}
		}
		if !sp {
			break
		}
		sp = false
	}
	return fmt.Sprintf("%s:%d", file, line), true
}

func getCaller(skip int) (string, int, bool) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0, ok
	}
	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line, true
}
func (l *log4goLogger) RecordInfo(first interface{}, info ...interface{}) {
	caller, ok := findCaller(3)
	src := ""

	if ok {
		// src += fmt.Sprintf("(%s) [requestId:%v]", caller,l.ReqID)
		src += fmt.Sprintf("(%s)", caller)
	}

	if l.Prefix != "" {
		src += fmt.Sprintf("[%s] ", l.Prefix)
	}

	isStr := false
	switch temp := first.(type) {
	case string:
		isStr = true
		src += temp
	}

	if nil != l.Decorate {
		src = l.Decorate(src)
	}

	if isStr {
		// if len(info) > 0 {
		// 	msg := fmt.Sprint(fmt.Sprintf(strings.Repeat(" %v", len(info)), info...))
		// 	src += msg
		// }
		// logrus.Info(src)
		log4go.InfoWithModule(l.Module, src, info...)
	} else {
		temp := []interface{}{first}
		info = append(temp, info...)
		log4go.InfoWithModule(l.Module, src, info)
	}
}
func (l *log4goLogger) RecordWarn(first interface{}, info ...interface{}) {
	caller, ok := findCaller(3)
	src := ""

	if ok {
		// src += fmt.Sprintf("(%s) [requestId:%v]", caller,l.ReqID)
		src += fmt.Sprintf("(%s)", caller)
	}

	if l.Prefix != "" {
		src += fmt.Sprintf("[%s] ", l.Prefix)
	}

	isStr := false
	switch temp := first.(type) {
	case string:
		isStr = true
		src += temp
	}

	if isStr {
		// if len(info) > 0 {
		// 	msg := fmt.Sprint(fmt.Sprintf(strings.Repeat(" %v", len(info)), info...))
		// 	src += msg
		// }
		// logrus.Info(src)
		log4go.WarnWithModule(l.Module, src, info...)
	} else {
		temp := []interface{}{first}
		info = append(temp, info...)
		log4go.WarnWithModule(l.Module, src, info)
	}
}

func (l *log4goLogger) RecordDebug(first interface{}, info ...interface{}) {
	caller, ok := findCaller(3)
	src := ""

	if l.Prefix != "" {
		src = fmt.Sprintf("[%s] ", l.Prefix)
	}

	if ok {
		src += fmt.Sprintf("(%s)", caller)
	}

	isStr := false
	switch temp := first.(type) {
	case string:
		isStr = true
		src += temp
	}

	if isStr {
		// if len(info) > 0 {
		// 	msg := fmt.Sprint(fmt.Sprintf(strings.Repeat(" %v", len(info)), info...))
		// 	src += msg
		// }
		// logrus.Debug(src)
		log4go.DebugWithModule(l.Module, src, info...)
	} else {
		temp := []interface{}{first}
		info = append(temp, info...)
		log4go.DebugWithModule(l.Module, src, info)
	}
}

func (l *log4goLogger) RecordError(first interface{}, info ...interface{}) {
	caller, ok := findCaller(3)
	src := ""

	if l.Prefix != "" {
		src = fmt.Sprintf("[%s] ", l.Prefix)
	}

	if ok {
		src += fmt.Sprintf("(%s)", caller)
	}

	isStr := false
	switch temp := first.(type) {
	case string:
		isStr = true
		src += temp
	}


	if isStr {
		// if len(info) > 0 {
		// 	msg := fmt.Sprint(fmt.Sprintf(strings.Repeat(" %v", len(info)), info...))
		// 	src += msg
		// }
		// logrus.Error(src)
		log4go.ErrorWithOptions(l.Module, src, info...)
	} else {
		temp := []interface{}{first}
		info = append(temp, info...)
		log4go.ErrorWithOptions(l.Module, src, info)
	}
}
