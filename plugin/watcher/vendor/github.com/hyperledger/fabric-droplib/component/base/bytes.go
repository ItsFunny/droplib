/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 3:01 下午
# @File : bytes.go
# @Description :
# @Attention :
*/
package base

type ByteHandler interface {
	Handle(data []byte) error
}
