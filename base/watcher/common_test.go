/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/6 6:30 下午
# @File : common_test.go
# @Description :
# @Attention :
*/
package watcher

import (
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
)

type MockConsumer struct {
	logger v2.Logger
}

func NewMockConsumer() *MockConsumer {
	r := &MockConsumer{
		logger: logrusplugin.NewLogrusLogger(modules.NewModule("MOCK_CONSUMER", 1)),
	}

	return r
}

func (c *MockConsumer) Consume(msg models2.Envelope) error {
	c.logger.Info("收到数据", "msg", msg)
	return nil
}
