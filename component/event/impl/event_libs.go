/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:09 上午
# @File : event_libs.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/libp2p/go-libp2p-core/event"
)

const (
	AVAILABLE = 0
	ACQUIRED  = 1
)

type SubscriptionWrapper struct {
	subscription event.Subscription
	status       int8
}


