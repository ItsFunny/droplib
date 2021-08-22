/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 5:28 下午
# @File : selectn_test.go.go
# @Description :
# @Attention :
*/
package watcher

import (
	"context"
	"fmt"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test_BaseUseSelectN(t *testing.T) {
	debug_async = true
	mockBaseUse(t, func() ChannelWatcher {
		return newSelectNChannelWatcher(DefaultForeverOpt)
	})
}
func Test_SelectNClose(t *testing.T) {
	debug_async = true
	watcher := newSelectNChannelWatcher(DefaultForeverOpt)
	mockBaseUse(t, func() ChannelWatcher {
		return watcher
	})
	time.Sleep(time.Second * 2)
	ch := watcher.getSelectNByName("test")
	require.Nil(t, ch)
	require.Equal(t, 0, watcher.Size())
}

func Test_Merge(t *testing.T) {
	debug_async=false
	defaultTestSleepInterval = math.MaxInt32
	n := 4
	values := asValues(n, 1)
	watcher := newSelectNChannelWatcher(DefaultForeverOpt)
	watcher.BStart()

	chs := make([]chan struct{}, n)
	booleans := make([]bool, n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for index := 0; index < n; index++ {
		chs[index] = make(chan struct{})
		booleans[index] = false
		go func(index int) {
			var f func(v channel.IData)
			f = func(v channel.IData) {
				if !booleans[index] {
					booleans[index] = true
					chs[index] <- struct{}{}
					close(chs[index])
				}
				time.Sleep(time.Millisecond * 2000)
			}
			name := "test" + strconv.Itoa(index)
			test, member := newMockTest(f, nil, name, values[index]...)
			watcher.WatchMemberChanged(member)
			test.BlockWaitPanic()
			wg.Done()
			logplugin.InfoF("%s,done", name)
		}(index)
	wait:
		for {
			select {
			case <-chs[index]:
				var size int
				size = watcher.GetRegionSize()
				if index == 0 {
					require.Equal(t, 1, size) // 1
				} else if index == 1 {
					require.Equal(t, 1, size) // 2
				} else if index == 2 {
					require.Equal(t, 2, size) // 2,1
				} else if index == 3 {
					require.Equal(t, 1, size) // 4
				}
				time.Sleep(time.Second * 1)
				break wait
			}
		}
	}
	wg.Wait()
	time.Sleep(time.Second)
	require.Equal(t, 0, watcher.Size())
	fmt.Println("done")
}

func Test_MergeMore(t *testing.T) {
	debug_async=false
	defaultTestSleepInterval = math.MaxInt32
	n := 8
	values := asValues(8, 1)
	watcher := newSelectNChannelWatcher(DefaultForeverOpt)
	watcher.BStart()

	chs := make([]chan struct{}, n)
	booleans := make([]bool, n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for index := 0; index < n; index++ {
		chs[index] = make(chan struct{})
		booleans[index] = false
		go func(index int) {
			var f func(v channel.IData)
			f = func(v channel.IData) {
				if !booleans[index] {
					booleans[index] = true
					chs[index] <- struct{}{}
					close(chs[index])
				}
				time.Sleep(time.Millisecond * 2000)
			}
			name := "test" + strconv.Itoa(index)
			test, member := newMockTest(f, nil, name, values[index]...)
			watcher.WatchMemberChanged(member)
			test.BlockWaitPanic()
			wg.Done()
			logplugin.InfoF("%s done", name)
		}(index)
	wait:
		for {
			select {
			case <-chs[index]:
				size := watcher.GetRegionSize()
				if index == 0 {
					require.Equal(t, 1, size)
				} else if index == 1 {
					require.Equal(t, 1, size)
				} else if index == 2 {
					require.Equal(t, 2, size)
				} else if index == 3 {
					require.Equal(t, 1, size)
				} else if index == 4 {
					require.Equal(t, 2, size)
				} else if index == 5 {
					require.Equal(t, 2, size)
				} else if index == 6 {
					require.Equal(t, 3, size)
				} else if index == 7 {
					require.Equal(t, 1, size)
				}
				break wait
			}
		}
	}
	wg.Wait()
	time.Sleep(time.Second)
	require.Equal(t, 0, watcher.Size())
	fmt.Println("done")
}

func Test_100SelectNRoutine(t *testing.T) {
	debug_async = false
	w := newSelectNChannelWatcher(DefaultForeverOpt)
	testWithN(t, testN{
		count:        1,
		routineLimit: 4096,
		consumerF:    nil,
		channelWatcherF: func() ChannelWatcher {
			return w
		},
		receiverF: nil,
	})
	fmt.Println("done")
}
func Test_More100SelectNRoutine(t *testing.T) {
	debug_async = true
	time.Sleep(time.Second)
	for i := 0; i < 200; i++ {
		t.Run("count_"+strconv.Itoa(i), func(t *testing.T) {
			w := newSelectNChannelWatcher(DefaultForeverOpt)
			testWithN(t, testN{
				count:        i,
				routineLimit: 100,
				channelWatcherF: func() ChannelWatcher {
					return w
				},
				receiverF:     nil,
				sleepInterval: 20,
			})
			w = nil
			fmt.Println("done")
			time.Sleep(time.Second)
		})
	}
}

func Test_TestRandomSleepSelectNRoutine(t *testing.T) {
	w := newSelectNChannelWatcher(DefaultForeverOpt)
	testWithN(t, testN{
		count:        0,
		routineLimit: 10,
		consumerF: func(v channel.IData) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10000)))
		},
		channelWatcherF: func() ChannelWatcher {
			return w
		},
		receiverF:     nil,
		sleepInterval: -1,
	})
	fmt.Println("done")
}

func mockChannels(f func(v channel.IData), n int) ([]ChannelMember, *groupTestWp) {
	chs := make([]ChannelMember, 0)
	tests := make([]*TestWp, 0)
	for i := 0; i < n; i++ {
		values := make([]channel.IData, 0)
		for j := 0; j < 10; j++ {
			values = append(values, VString("_"+strconv.Itoa(0)+"_"+strconv.Itoa(i*10+j)))
		}
		test, c := newMockTest(f, nil, "test"+strconv.Itoa(i), values...)
		chs = append(chs, c)
		tests = append(tests, test)
	}
	r := newGroupTestWp(tests)
	return chs, r
}

func Test_WithPrepareChannels(t *testing.T) {
	// channels
	chs, r := mockChannels(nil, 100)
	w := newSelectNChannelWatcher(DefaultForeverOpt)
	cc := toChWp(chs)
	w.BStart(models.CtxStartOpt(context.WithValue(context.Background(), "channels", &cc)))
	now := time.Now()
	r.BlockWaitPanic()
	cost := time.Now().Sub(now)
	logplugin.Info(fmt.Sprintf("done,耗时:%f秒,%d毫秒", cost.Seconds(), cost.Milliseconds()))
}
func toChWp(chs []ChannelMember) []chWp {
	cc := make([]chWp, 0)
	for _, c := range chs {
		cc = append(cc, chWp{
			c:        c.c,
			consumer: c.consumer,
			name:     c.name,
			newM:     true,
		})
	}
	return cc
}

func Test_selectNChannelWatcher_rollBack(t *testing.T) {
	defaultTestSleepInterval = math.MaxInt32
	w := newSelectNChannelWatcher(DefaultForeverOpt)
	chs, wp := mockChannels(func(v channel.IData) {
		time.Sleep(time.Second * 1)
	}, 1)
	cc := toChWp(chs)
	w.BStart(models.CtxStartOpt(context.WithValue(context.Background(), "channels", &cc)))
	time.Sleep(time.Second)
	w.RollBack()
	wp.BlockWaitPanic()
}
