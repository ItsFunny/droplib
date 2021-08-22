/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/12 7:02 下午
# @File : types.go
# @Description :
# @Attention :
*/
package watcher

type WatcherType byte

const (
	WATCHER_TYPE_ROUTINE WatcherType = iota
	WATCHER_TYPE_REFLECT
	WATCHER_TYPE_SELECTN
)
