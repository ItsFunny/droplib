/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 12:00 下午
# @File : error.go
# @Description :
# @Attention :
*/
package libutils

import "errors"

func PanicIfError(prefix string, err error) {
	if nil != err {
		panic(errors.New(prefix + "," + err.Error()))
	}
}

