/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/13 3:05 下午
# @File : watcher_test.go
# @Description :
# @Attention :
*/
package watcher

import (
	"github.com/hyperledger/fabric-droplib/base/log/common"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	"math"
	"testing"
	"time"
)

var specificWartchers = []struct {
	name      string
	watchType WatcherType
}{
	{
		name:      "routine",
		watchType: WATCHER_TYPE_ROUTINE,
	},
	{
		name:      "reflect",
		watchType: WATCHER_TYPE_REFLECT,
	},
	{
		name:      "selectn",
		watchType: WATCHER_TYPE_SELECTN,
	},
}

func Test_SmallJob(t *testing.T) {
	channels, wp := mockChannels(func(v channel.IData) {
	}, 100)
	for _, wh := range specificWartchers {
		t.Run(wh.name, func(t *testing.T) {
			watcher := NewForeverWatcher(wh.watchType)
			watcher.BStart()
			for i := 0; i < len(channels); i++ {
				go func(index int) {
					watcher.RegisterNewChannel(channels[index].name, channels[index].c, channels[index].consumer)
				}(i)
			}
		})
	}
	wp.BlockWaitPanic()
}

func Test_LongJob(t *testing.T) {
	debug_async = false
	defaultTestSleepInterval = math.MaxInt32
	channels, wp := mockChannels(func(v channel.IData) {
		time.Sleep(time.Second * 10)
	}, 100)
	for _, wh := range specificWartchers {
		t.Run(wh.name, func(t *testing.T) {
			watcher := NewForeverWatcher(wh.watchType)
			watcher.BStart()
			for i := 0; i < len(channels); i++ {
				go func(index int) {
					watcher.RegisterNewChannel(channels[index].name, channels[index].c, channels[index].consumer)
				}(i)
			}
		})
	}
	wp.BlockWaitPanic()
}

func TestRoutineToReflect(t *testing.T) {
	defaultTestSleepInterval = math.MaxInt32
	routeUpgrade := RoutineUpgradeLimitOption(1)
	reflectNoRollback := ReflectNoRollbackOption()
	watcher := NewChannelWatcher(routeUpgrade, reflectNoRollback)
	watcher.BStart()
	channels, wp := mockChannels(func(v channel.IData) {
		time.Sleep(time.Second * 1)
	}, 2)

	go watcher.RegisterNewChannel(channels[0].name, channels[0].c, channels[0].consumer)
	go func() {
		time.Sleep(time.Second * 5)
		watcher.RegisterNewChannel(channels[1].name, channels[1].c, channels[1].consumer)
	}()
	wp.BlockWaitPanic()
}

// func TestRoutineToReflectToSelecn(t *testing.T) {
// 	debug_async = false
//
// 	defaultTestSleepInterval = math.MaxInt32
// 	routeUpgrade := RoutineUpgradeLimitOption(1)
// 	reflectUpgrade := ReflectUpgradeLimitOption(2)
// 	reflectNoRollback := ReflectNoRollbackOption()
// 	selectnNoRollBack := SelectNNoRollBackOption()
//
// 	watcher := NewChannelWatcher(routeUpgrade, reflectNoRollback, selectnNoRollBack, reflectUpgrade)
// 	watcher.BStart()
// 	channels, wp := mockChannels(func(v channel.IData) {
// 		time.Sleep(time.Second * 2)
// 	}, 3)
//
// 	go watcher.RegisterNewChannel(channels[0].name, channels[0].c, channels[0].consumer)
// 	go func() {
// 		time.Sleep(time.Second * 1)
// 		watcher.RegisterNewChannel(channels[1].name, channels[1].c, channels[1].consumer)
// 	}()
// 	go func() {
// 		time.Sleep(time.Second * 5)
// 		watcher.RegisterNewChannel(channels[2].name, channels[2].c, channels[2].consumer)
// 	}()
// 	wp.BlockWaitPanic()
// }=>

func TestConcurrent(t *testing.T) {
	logcomponent.SetGlobalLogLevel(common.DebugLevel)
	defaultTestSleepInterval = 600
	commonTestNCounts(t, 50, func() {
		debug_async = true
		watcher := NewChannelWatcher()
		watcher.BStart()
		channels, wp := mockChannels(func(v channel.IData) {
			time.Sleep(time.Millisecond * 2)
		}, 8092)
		for i := 0; i < len(channels); i++ {
			go func(index int) {
				watcher.RegisterNewChannel(channels[index].name, channels[index].c, channels[index].consumer)
			}(i)
		}
		wp.BlockWaitPanic()
	})

}

func TestRandomJobConcurrent(t *testing.T) {
	logcomponent.SetGlobalLogLevel(common.DebugLevel)
	defaultTestSleepInterval = 6000
	commonTestNCounts(t, 50, func() {
		debug_async = true

		watcher := NewChannelWatcher()
		watcher.BStart()
		channels, wp := mockChannels(func(v channel.IData) {
			time.Sleep(time.Millisecond * (time.Duration(rand.RandInt32(10, 3000))))
		}, 4096)
		for i := 0; i < len(channels); i++ {
			go func(index int) {
				watcher.RegisterNewChannel(channels[index].name, channels[index].c, channels[index].consumer)
			}(i)
		}
		wp.BlockWaitPanic()
	})
}

