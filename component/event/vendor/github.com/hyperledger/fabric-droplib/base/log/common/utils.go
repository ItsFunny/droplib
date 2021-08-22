/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 10:29 上午
# @File : utils.go
# @Description :
# @Attention :
*/
package common

import (
	"fmt"
	"runtime"
	"strings"
)



func FindCaller(skip int,blackList []string) (string, bool) {
	file := ""
	line := 0
	ok := false
	sp := false
	for i := 0; i < 10; i++ {
		file, line, ok = getCaller(skip + i) //
		if !ok {
			return "", false
		}
		if strings.Contains(file,"log@"){
			sp=true
		}else{
			for _, bl := range blackList {
				if strings.HasPrefix(file, bl) {
					sp = true
					break
				}
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
