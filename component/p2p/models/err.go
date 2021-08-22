/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 10:22 下午
# @File : err.go
# @Description :
# @Attention :
*/
package models

type PeerError struct {
	NodeID NodeID
	Err    error
}

type NetworkError struct {
	error
}

type LogicError struct {
	error
}

