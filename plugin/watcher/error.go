/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/16 5:44 下午
# @File : error.go
# @Description :
# @Attention :
*/
package watcher

import (
	"errors"
	"fmt"
)

var (
	PROGRAMA_ERROR = errors.New("编码错误")
)

// FIXME ,丢到log中
func PanicWithMsg(e error, msg string) {
	panic(fmt.Sprintf("err:%v,msg:%s", e, msg))
}
