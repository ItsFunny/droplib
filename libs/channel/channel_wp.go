/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/31 6:26 下午
# @File : channel_wp.go
# @Description :
# @Attention :
*/
package channel

type ChannelID string

type ChannelShim struct {
}
type ChannelDescriptor struct {
	ID                  byte
	Priority            int
	SendQueueCapacity   int
	RecvMessageCapacity int
	RecvBufferCapacity  int
	MaxSendBytes        uint
}

type Channel struct {
	Id ChannelID
	Ch chan IData
}

func (this *Channel) Close() {
	close(this.Ch)
}

type Envelope struct {
	ChannelId ChannelID
	Data      IData
}
