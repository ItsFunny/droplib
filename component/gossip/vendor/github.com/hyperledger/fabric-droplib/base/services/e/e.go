/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/16 1:36 下午
# @File : e.go
# @Description :
# @Attention :
*/
package e

import "errors"

var (
	STORE_ITERATOR_ERROR = errors.New("STOP_ITERATOR")

	FORCELOSE_ERROR=errors.New("FORCE_CLOSE")
)
