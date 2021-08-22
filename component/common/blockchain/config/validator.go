/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 3:47 下午
# @File : validator.go
# @Description :
# @Attention :
*/
package config

type ValidatorConfiguration struct {
	PubKeyTypes []string
}

func NewDefaultValidatorConfiguration() *ValidatorConfiguration {
	r := ValidatorConfiguration{
		PubKeyTypes: []string{"ecdsa", "sm2"},
	}
	return &r
}
