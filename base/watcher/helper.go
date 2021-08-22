/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/6 6:35 下午
# @File : helper.go
# @Description :
# @Attention :
*/
package watcher

func ConvGroupT2OnRunning(group *ChannelGroup) *OnRunningChannelGroup {
	r := &OnRunningChannelGroup{
		OnClose: group.OnClose,
	}
	return r
}
