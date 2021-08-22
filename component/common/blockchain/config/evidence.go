/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 3:47 下午
# @File : evidence.go
# @Description :
# @Attention :
*/
package config

import (
	"time"
)

type EvidenceConfiguration struct {
	MaxAgeNumBlocks int64
	MaxAgeDuration  time.Duration
	MaxBytes        int64
}

func NewDefaultEvidenceConfiguration() *EvidenceConfiguration {
	r := EvidenceConfiguration{
		MaxAgeNumBlocks: DEFAULT_CONSTANT,
		MaxAgeDuration:  time.Duration(DEFAULT_CONSTANT),
		MaxBytes:        DEFAULT_CONSTANT,
	}

	return &r
}
