/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 8:27 上午
# @File : intergration_test.go
# @Description :
# @Attention :
*/
package watcher

import (
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/log/common"
	logcomponent "github.com/hyperledger/fabric-droplib/base/log/v2/component"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	"math"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"
)

var watchers = []struct {
	name string
	f    func() ChannelWatcher
}{
	{"routine", func() ChannelWatcher {
		return newRoutineChannelWatcher(DefaultForeverOpt)
	}},
	{
		"reflect", func() ChannelWatcher {
		return newReflectChannelWatcher(DefaultForeverOpt)
	}},
	{
		"selectn", func() ChannelWatcher {
		return newSelectNChannelWatcher(DefaultForeverOpt)
	},
	},
}

func Test_BaseUse(t *testing.T) {
	debug_async = true
	for _, watcher := range watchers {
		t.Run(watcher.name, func(t *testing.T) {
			mockBaseUse(t, watcher.f)
		})
	}
}

func Test_MoreBaseUse(t *testing.T) {
	defaultTestSleepInterval = math.MaxInt32
	for _, w := range watchers {
		t.Run(w.name, func(t *testing.T) {
			testWithN(t, testN{
				count:           1,
				routineLimit:    4096,
				consumerF:       nil,
				channelWatcherF: w.f,
				receiverF:       nil,
			})
		})
	}
}

func TestSmallJob(t *testing.T) {
	defaultTestSleepInterval = math.MaxInt32
	for _, w := range watchers {
		t.Run(w.name, func(t *testing.T) {
			commonTestSmallJob(t, w.f)
		})
	}
}

func TestRollBack(t *testing.T) {
	for _, w := range watchers {
		if strings.Contains(w.name, "routine") {
			continue
		}
		t.Run(w.name, func(t *testing.T) {
			mockRollBack(1, func() ChannelWatcher {
				return w.f()
			})
		})
	}
}
func Test_ConcurrentRollBack(t *testing.T) {
	for _, w := range watchers {
		if strings.Contains(w.name, "routine") {
			continue
		}
		t.Run(w.name, func(t *testing.T) {
			mockRollBack(100, func() ChannelWatcher {
				return w.f()
			})
		})
	}
}

func Test_MoreConcurrentRollBack(t *testing.T) {
	for _, w := range watchers {
		if strings.Contains(w.name, "routine") {
			continue
		}
		t.Run(w.name, func(t *testing.T) {
			mockRollBack(4096, func() ChannelWatcher {
				return w.f()
			})
		})
	}
}

func TestOneUpgrade(t *testing.T) {
	for _, w := range watchers {
		if strings.Contains(w.name, "selectn") {
			continue
		}
		t.Run(w.name, func(t *testing.T) {
			mockUpgrade(1, func() ChannelWatcher {
				return w.f()
			})
		})
	}
}

func TestConcurrentUpgrade(t *testing.T) {
	logcomponent.SetGlobalLogLevel(common.WarnLevel)

	for _, w := range watchers {
		if strings.Contains(w.name, "selectn") {
			continue
		}
		t.Run(w.name, func(t *testing.T) {
			mockUpgrade(100, func() ChannelWatcher {
				return w.f()
			})
		})
	}
}
func TestMoreConcurrentUpgrade(t *testing.T) {
	commonTestNCounts(t, 2, func() {
		for _, w := range watchers {
			if strings.Contains(w.name, "selectn") {
				continue
			}
			t.Run(w.name, func(t *testing.T) {
				mockUpgrade(4096, func() ChannelWatcher {
					return w.f()
				})
			})
		}
	})
}

// 测试随机的rollBack 和upgrade
func TestFrequencyRollbackUpgrade(t *testing.T) {
	logcomponent.SetGlobalLogLevel(common.InfoLevel)
	commonTestNCounts(t, 10, func() {
		opts := foreverOptions()
		wh := watchers[0].f()
		defaultTestSleepInterval = math.MaxInt32
		wh.BStart()
		// 100个 member
		members, wp := mockChannels(func(v channel.IData) {
			time.Sleep(time.Second)
		}, 1024)
		go func() {
			upgrade := false
			index := 0
			add := func(w ChannelWatcher) {
				for i := 0; i < rand.Intn(5) && index < len(members); i++ {
					retryCount := 0
					for w.WatchMemberChanged(members[index]) {
						if retryCount >= 5 {
							return
						}
						logplugin.Info("添加member重试", "name", members[index].name)
						time.Sleep(time.Millisecond * 150)
						retryCount++
					}
					logplugin.Info("添加member", "name", members[index].name, "index", index)
					index++
				}
			}
			changeCount := 0
			for {
				time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(100, 1500)))
				if changeCount%5 == 0 {
					switch wh.(type) {
					case *selectNChannelWatcher:
						upgrade = false
						wh = wh.RollBack(opts...)
					case *routingChannelWatcher:
						upgrade = true
						wh = wh.Upgrade(opts...)
					case *reflectChannelWatcher:
						if upgrade {
							upgrade = false
							wh = wh.Upgrade(opts...)
						} else {
							upgrade = true
							wh = wh.RollBack(opts...)
						}
					}
				}
				add(wh)
				changeCount++
				if index >= len(members) {
					logplugin.Info("退出")
					return
				}
			}
		}()
		wp.BlockWaitPanic()
	})
}

func TestSelectNRollbackUpgrade(t *testing.T) {
	debug_async = true
	logcomponent.SetGlobalLogLevel(common.InfoLevel)
	commonTestNCounts(t, 3, func() {
		opts := foreverOptions()
		wh := watchers[1].f()
		defaultTestSleepInterval = math.MaxInt32
		wh.BStart()
		// 100个 member
		members, wp := mockChannels(func(v channel.IData) {
			time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(100, 1300)))
			// time.Sleep(time.Second)
		}, 4096)
		index := 0
		go func() {
			upgrade := true
			tt := time.NewTicker(time.Second * 3)
			add := func(w ChannelWatcher) {
				for i := 0; i < 5 && index < len(members); i++ {
					retryCount := 0
					for !w.Started() || w.WatchMemberChanged(members[index]) {
						if retryCount >= 5 {
							return
						}
						logplugin.Info("添加member重试", "name", members[index].name)
						time.Sleep(time.Millisecond * 150)
						retryCount++
					}
					logplugin.Info("add member", "name", members[index].name, "index", index)
					index++
				}
			}
			for {
				select {
				case <-tt.C:
					if upgrade {
						upgrade = false
						wh = wh.Upgrade(opts...)
					} else {
						upgrade = true
						wh = wh.RollBack(opts...)
					}
					add(wh)
					if index >= len(members) {
						logplugin.Info("退出")
						return
					}
				}
			}
		}()
		wp.BlockWaitPanic()
	})
}

// func Test_RefSelcRollbackUpgrade(t *testing.T) {
// 	defaultTestSleepInterval = math.MaxInt32
// 	opts := foreverOptions()
// 	wh := watchers[1].f()
// 	// defaultTestSleepInterval = 1000
// 	wh.BStart()
// 	// 100个 member
// 	members, wp := mockChannels(func(v channel.IData) {
// 		// time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(100, 400)))
// 		time.Sleep(time.Second * 1)
// 	}, 4096)
//
// 	index := 0
// 	go func() {
// 		upgrade := true
// 		tt := time.NewTicker(time.Second * 3)
// 		add := func(w ChannelWatcher) {
// 			for i := 0; i < 5 && index < len(members); i++ {
// 				retryCount := 0
// 				for !w.Started() || w.WatchMemberChanged(members[index]) {
// 					if retryCount >= 5 {
// 						return
// 					}
// 					logplugin.Info("添加member重试", "name", members[index].name)
// 					time.Sleep(time.Millisecond * 150)
// 					retryCount++
// 				}
// 				logplugin.Info("add member", "name", members[index].name, "index", index)
// 				index++
// 			}
// 		}
// 		for {
// 			select {
// 			case <-tt.C:
// 				if upgrade {
// 					upgrade = false
// 					wh = wh.Upgrade(opts...)
// 				} else {
// 					upgrade = true
// 					wh = wh.RollBack(opts...)
// 				}
// 				add(wh)
// 				if index >= len(members) {
// 					logplugin.Info("退出")
// 					return
// 				}
// 			}
// 		}
// 	}()
// 	wp.BlockWaitPanic()
// }

func Test_RouteReflectRollbackUpgrade(t *testing.T) {
	defaultTestSleepInterval = math.MaxInt32
	opts := foreverOptions()
	wh := watchers[0].f()
	defaultTestSleepInterval = 10000
	wh.BStart()
	// 100个 member
	members, wp := mockChannels(func(v channel.IData) {
		// time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(100, 400)))
		time.Sleep(time.Second * 1)
	}, 4096)

	index := 0
	go func() {
		upgrade := true
		tt := time.NewTicker(time.Second * 3)
		add := func(w ChannelWatcher) {
			for i := 0; i < 5 && index < len(members); i++ {
				retryCount := 0
				for !w.Started() || w.WatchMemberChanged(members[index]) {
					if retryCount >= 5 {
						return
					}
					logplugin.Info("添加member重试", "name", members[index].name)
					time.Sleep(time.Millisecond * 150)
					retryCount++
				}
				logplugin.Info("add member", "name", members[index].name, "index", index)
				index++
			}
		}
		for {
			select {
			case <-tt.C:
				if upgrade {
					upgrade = false
					wh = wh.Upgrade(opts...)
				} else {
					upgrade = true
					wh = wh.RollBack(opts...)
				}
				add(wh)
				if index >= len(members) {
					logplugin.Info("退出")
					return
				}
			}
		}
	}()
	wp.BlockWaitPanic()
}
