/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/2/5 17:47
# @File : log.go
# @Description :
# @Attention :
*/
package libutils

import (
	"fmt"
	"runtime"
	"strings"
)

var debugBlackList = []string{"debugutil/debug", "utils/log", "cache/node_cache.go"}
// 查询调用栈
func FindMoreCallers(calls int, blackList ...string) string {
	msg := strings.Builder{}
	ok := false
	bL := make([]string, 0)
	bL = append(bL, debugBlackList...)
	bL = append(bL, blackList...)
	for i := 0; i < calls; i++ {
		s, b := FindCaller(i, bL)
		if b {
			ok = b
			msg.WriteString(s)
			bL = append(bL, s)
		}
	}
	if ok {
		return msg.String()
	}
	return ""
}
func FindCaller(skip int, blackList []string) (string, bool) {
	file := ""
	line := 0
	ok := false
	sp := false
	for i := 0; i < 10; i++ {
		file, line, ok = getCaller(skip + i) //
		if !ok {
			return "", false
		}                                    //
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
