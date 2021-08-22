/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 2:41 下午
# @File : mem_pool.go
# @Description :
# @Attention :
*/
package commonerrors

import "errors"

var (
	// ErrTxInCache is returned to the client if we saw tx earlier
	ErrTxInCache = errors.New("tx already exists in cache")
)

