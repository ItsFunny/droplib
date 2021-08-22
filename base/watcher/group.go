/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/6 3:21 下午
# @File : group.go
# @Description :
# @Attention :
*/
package watcher

// 运行中的group
type OnRunningChannelGroup struct {
	OnClose func()
}
