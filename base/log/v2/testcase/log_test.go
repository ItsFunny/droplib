/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 3:37 下午
# @File : log_test.go
# @Description :
# @Attention :
*/
package testcase

import (
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"testing"
)

func Test_LogRus(t *testing.T) {
	m1 := modules.NewModule("demo", 1)
	logger := logrusplugin.NewLogrusLogger(m1)
	logger.Info("123", 4, 5, 6)
	logplugin.Info("asd")
	m2 := modules.NewModule("newmodule", 2)
	logplugin.MInfo(m2, "asd")
}
