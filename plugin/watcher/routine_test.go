/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 3:21 下午
# @File : routine_test.go.go
# @Description :
# @Attention :
*/
package watcher

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_one(t *testing.T) {
	mockBaseUse(t, func() ChannelWatcher {
		return newRoutineChannelWatcher(DefaultForeverOpt)
	})
}

func Test_Close(t *testing.T) {
	watcher := newRoutineChannelWatcher(DefaultForeverOpt)
	mockBaseUse(t, func() ChannelWatcher {
		return watcher
	})
	time.Sleep(time.Second)
	require.Equal(t, 0, watcher.Size())
}
func Test_100Routine(t *testing.T) {
	testWithN(t, testN{
		count:        1,
		routineLimit: 100,
		channelWatcherF: func() ChannelWatcher {
			return newRoutineChannelWatcher(DefaultForeverOpt)
		},
	})
}
func Test_Routine(t *testing.T) {
	watcher := newRoutineChannelWatcher(DefaultForeverOpt)
	watcher.BStart()
}
func Test_RoutineOneUpgrade(t *testing.T) {
	c := mockUpgrade(1, func() ChannelWatcher {
		return newRoutineChannelWatcher(DefaultForeverOpt)
	})
	if _, ok := c.(*reflectChannelWatcher); !ok {
		t.Error("not right")
	}
}

func Test_MoreUpgrade(t *testing.T) {
	for i := 0; i < 100; i++ {
		c := mockUpgrade(1, func() ChannelWatcher {
			return newRoutineChannelWatcher(DefaultForeverOpt)
		})
		if _, ok := c.(*reflectChannelWatcher); !ok {
			t.Error("not right")
		}
	}
}

func Test_ConcurrentUpgrade(t *testing.T) {
	commonTestNCounts(t, 20, func() {
		c := mockUpgrade(4096, func() ChannelWatcher {
			return newRoutineChannelWatcher(DefaultForeverOpt)
		})
		if _, ok := c.(*reflectChannelWatcher); !ok {
			t.Error("not right")
		}
	})
}
func TestRoutineDelta(t *testing.T) {
	mockTestDeleta(func(opt Opt) ChannelWatcher {
		return newRoutineChannelWatcher(opt)
	}, false)
}