/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/12 7:21 上午
# @File : property.go
# @Description :
# @Attention :
*/
package watcher

type ChannelWatcherProperty struct {
	MaxWorker int

	MaxMsgInFlight int
	StepOneLimit   int16
	StepTwoLimit   int16
	SpecialName    []string
}
