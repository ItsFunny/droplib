/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/2 1:00 下午
# @File : component_test.go
# @Description :
# @Attention :
*/
package test

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"testing"
)

func TestCompoennt_No_Ready(t *testing.T) {
	logger := logrusplugin.NewLogrusLogger(modules.NewModule("1", 1))
	fmt.Println(logger)
}
