/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 3:34 下午
# @File : block.go
# @Description :
# @Attention :
*/
package config

type BlockConfiguration struct {
	MaxBytes int64 `json:"maxBytes"`
	MaxGas   int64
}

func NewDefaultBlockConfiguration() *BlockConfiguration {
	c := BlockConfiguration{
		MaxBytes: 104857600,
		MaxGas:   104857600,
	}
	return &c
}
