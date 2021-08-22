/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/15 10:49 上午
# @File : sync.go
# @Description :
# @Attention :
*/
package libsync

import "sync"

// A Mutex is a mutual exclusion lock.
type Mutex struct {
	sync.Mutex
}

// An RWMutex is a reader/writer mutual exclusion lock.
type RWMutex struct {
	sync.RWMutex
}