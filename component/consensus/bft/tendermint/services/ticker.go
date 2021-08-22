/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 5:11 下午
# @File : ticker.go
# @Description :
# @Attention :
*/
package services

import (
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
)

type ITimeoutTicker interface {
	Start(flag services.START_FLAG) error
	Stop() error
	Chan() <-chan models.TimeoutInfo       // on which to receive a timeout
	ScheduleTimeout(ti models.TimeoutInfo) // reset the timer
	SetLogger(logger v2.Logger)
}
