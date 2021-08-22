/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/16 10:04 上午
# @File : ticket.go
# @Description :
# @Attention :
*/
package config

import "time"

const (
	DEFAULT_TICKET_INTERVAL = 20
	DEFAULT_TIME_DURATION   = 10
)

type TicketConfiguration struct {
	TickInterval  time.Duration
	HeartbeatTick int
}

func NewDefaultTicketConfiguration() *TicketConfiguration {
	return &TicketConfiguration{
		TickInterval:  time.Second * 20,
		HeartbeatTick: DEFAULT_TIME_DURATION,
	}
}
