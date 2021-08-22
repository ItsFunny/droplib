/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 4:40 下午
# @File : reflect_test.go.go
# @Description :
# @Attention :
*/
package watcher

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/common"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"time"
)

func Test_ReflectBaseUse(t *testing.T) {
	mockBaseUse(t, func() ChannelWatcher {
		return newReflectChannelWatcher(DefaultForeverOpt)
	})
	time.Sleep(time.Second * 2)
}
func Test_ReflectSleep(t *testing.T) {
	w := newReflectChannelWatcher(DefaultForeverOpt)
	w.BStart()
	ms, wp := mockChannels(func(v channel.IData) {
	}, 2)
	w.WatchMemberChanged(ms[0])
	time.Sleep(time.Second * 5)
	w.WatchMemberChanged(ms[1])
	wp.BlockWaitPanic()

}
func Test_ReflectClose(t *testing.T) {
	watcher := newReflectChannelWatcher(DefaultForeverOpt)
	mockBaseUse(t, func() ChannelWatcher {
		return watcher
	})
	time.Sleep(time.Second)
	require.Equal(t, 0, watcher.Size())
}

func Test_Reflect100Routine(t *testing.T) {
	testWithN(t, testN{
		count:        1,
		routineLimit: 100,
		consumerF: func(v channel.IData) {
			time.Sleep(time.Second)
		},
		channelWatcherF: func() ChannelWatcher {
			return newInternalTestReflectChannelWatcher()
		},
		receiverF:     nil,
		sleepInterval: math.MaxInt32,
	})
}

func Test_AsyncReflect100Routine(t *testing.T) {
	debug_async = true
	testWithN(t, testN{
		count:        1,
		routineLimit: 100,
		consumerF: func(v channel.IData) {
			time.Sleep(time.Second)
		},
		channelWatcherF: func() ChannelWatcher {
			return newInternalTestReflectChannelWatcher()
		},
		receiverF:     nil,
		sleepInterval: math.MaxInt32,
	})
}

func Test_ReflectUpgradeOne(t *testing.T) {
	c := mockUpgrade(1, func() ChannelWatcher {
		return newInternalTestReflectChannelWatcher()
	})
	if _, ok := c.(*selectNChannelWatcher); !ok {
		t.Error("not right")
	}
}
func Test_ReflectUpgrade(t *testing.T) {

	for i := 0; i < 10; i++ {
		c := mockUpgrade(1, func() ChannelWatcher {
			return newInternalTestReflectChannelWatcher()
		})
		if _, ok := c.(*selectNChannelWatcher); !ok {
			t.Error("not right")
		}
	}
}
func TestReflectConcurrentockUpgrade100(t *testing.T) {
	c := mockUpgrade(100, func() ChannelWatcher {
		return newInternalTestReflectChannelWatcher()
	})
	if _, ok := c.(*selectNChannelWatcher); !ok {
		t.Error("not right")
	}
}
func Test_MoreReflectConcurrentReflectUpgrade(t *testing.T) {
	commonTestNCounts(t, 10, func() {
		TestReflectConcurrentockUpgrade100(t)
	})
}

func Test_ReflectRollback(t *testing.T) {
	back := mockRollBack(1, func() ChannelWatcher {
		return newInternalTestReflectChannelWatcher()
	})
	fmt.Println(back)
}
func Test_ReflectConcurrentRollBack(t *testing.T) {
	back := mockRollBack(4096, func() ChannelWatcher {
		return newInternalTestReflectChannelWatcher()
	})
	fmt.Println(back)
}
func Test_MoreReflectConcurrentRollBack(t *testing.T) {
	commonTestNCounts(t, 100, func() {
		Test_ReflectConcurrentRollBack(t)
	})
}
func Test_MoreReflectRollBack(t *testing.T) {
	for i := 0; i < 100; i++ {
		back := mockRollBack(1, func() ChannelWatcher {
			return newInternalTestReflectChannelWatcher()
		})
		if _, ok := back.(*routingChannelWatcher); !ok {
			t.Error("panic")
		}
	}
}
func Test_ReflectConcurrentUpgrade(t *testing.T) {
	logcomponent.SetGlobalLogLevel(common.InfoLevel)
	commonTestNCounts(t, 100, func() {
		back := mockUpgrade(4096, func() ChannelWatcher {
			return newReflectChannelWatcher(DefaultForeverOpt)
		})
		if _, ok := back.(*selectNChannelWatcher); !ok {
			t.Error("panic")
		}
	})
}

func Test_ReflectRollBackDeleta(t *testing.T) {
	mockTestDeleta(func(opt Opt) ChannelWatcher {
		return newReflectChannelWatcher(opt)
	}, true)
}
