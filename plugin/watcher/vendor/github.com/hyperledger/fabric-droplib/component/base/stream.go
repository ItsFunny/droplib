/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 3:26 下午
# @File : stream.go
# @Description :
# @Attention :
*/
package base

import "io"

type IStream interface {
	io.Reader
}
