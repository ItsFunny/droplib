/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 20218/12 6:21 下午
# @File : common_test.go
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
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type MockVerify struct {
	m map[channel.IData]struct{}
}

var defaultReceiverF = func(id string, v channel.IData) {
	debugPrint(fmt.Sprintf("id=%s,收到:%v", id, v))
}
var defaultTestSleepInterval = 20

type mockConsumer struct {
	f                 func(v channel.IData)
	stopCallBackLimit int
	async             bool
	callBackQueue     chan channel.IData
}

func (m *mockConsumer) Async() bool {
	return m.async
}
func (m *mockConsumer) Handle(i channel.IData) {
	m.f(i)
	m.callBackQueue <- i
	m.stopCallBackLimit--
	if m.stopCallBackLimit == 0 {
		close(m.callBackQueue)
	}
}

var callBackQueueFuncWrapper = func(q chan channel.IData, f func(vv channel.IData)) func(channel.IData) {
	return func(i channel.IData) {
		f(i)
		q <- i
	}
}

func listenNew(name string, v ...channel.IData) (chan channel.IData, ChannelMember) {
	cc := asChan(v...)
	forTest := make(chan channel.IData)
	r := ChannelMember{
		name:     name,
		c:        cc,
		consumer: NewFuncConsumer(callBackQueueFuncWrapper(forTest, defaultEmptyFunc)),
	}
	return forTest, r
}

type VString string

func (V VString) ID() interface{} {
	return V
}

func exceptedWithPanic(f func(id string, v channel.IData), sleepSeoncds int, id string, excepted []channel.IData, r <-chan channel.IData) (chan int, map[channel.IData]struct{}) {
	n := make(chan int)
	m := make(map[channel.IData]struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(sleepSeoncds))
		defer cancel()

		for _, v := range excepted {
			m[v] = struct{}{}
		}
		for {
			select {
			case <-ctx.Done():
				logplugin.Error(fmt.Sprintf("超时没有等到全部结果,剩余个数%d,%v", len(m), m))
				n <- len(m)
				return
			case v, ok := <-r:
				if !ok {
					n <- len(m)
					return
				}
				f(id, v)
				vs := v.(VString)
				if strings.Contains(string(vs), ":") {
					v = VString(strings.Split(string(vs), ":")[1])
				}
				if _, exist := m[v]; !exist {
					panic(fmt.Sprintf("%v,%v", m, v))
				} else {
					delete(m, v)
					if len(m) == 0 {
						n <- 0
						return
					}
				}
			}
		}
	}()
	return n, m
}

type TestWp struct {
	notify   <-chan int
	m        map[channel.IData]struct{}
	name     string
	finished bool
}
type groupTestWp struct {
	tests []*TestWp
}

func newGroupTestWp(tests []*TestWp) *groupTestWp {
	return &groupTestWp{
		tests: tests,
	}
}

func (this *groupTestWp) BlockWaitPanic() {
	wg := sync.WaitGroup{}
	wg.Add(len(this.tests))
	ok := make(chan struct{})
	go func() {
		ti := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ok:
				return
			case <-ti.C:

			}
		}
	}()
	all := int32(len(this.tests))
	for index := range this.tests {
		go func(index int) {
			now := time.Now()
			this.tests[index].BlockWaitPanic()
			cost := time.Now().Sub(now)
			debugPrint("一个任务done", "name", this.tests[index].name, "耗时", cost.Seconds(), "剩余", atomic.AddInt32(&all, -1))
			wg.Done()
		}(index)
	}
	wg.Wait()

}
func (this *TestWp) BlockWaitPanic() {
	v := <-this.notify
	if v != 0 {
		panic(v)
		os.Exit(-1)
	}
	this.finished = true
}

var defaultEmptyFunc = func(v channel.IData) {}

func newMockTest(f func(v channel.IData), receiverF func(id string, v channel.IData), name string, excepted ...channel.IData) (*TestWp, ChannelMember) {
	if f == nil {
		f = defaultEmptyFunc
	}
	if nil == receiverF {
		receiverF = defaultReceiverF
	}
	newE := make([]channel.IData, len(excepted))
	for i := 0; i < len(excepted); i++ {
		newE[i] = VString(name + "_" + string(excepted[i].(VString)))
	}
	excepted = newE
	cc := asChan(excepted...)
	callBack := make(chan channel.IData)
	con := &mockConsumer{
		f:                 f,
		stopCallBackLimit: len(excepted),
		async:             debug_async,
		callBackQueue:     callBack,
	}
	mem := ChannelMember{
		name:     name,
		c:        cc,
		consumer: con,
	}
	r, m := exceptedWithPanic(receiverF, defaultTestSleepInterval, name, excepted, callBack)
	testW := &TestWp{
		notify:   r,
		m:        m,
		name:     name,
		finished: false,
	}
	return testW, mem
}

type testN struct {
	async           bool
	count           int
	routineLimit    int
	consumerF       func(v channel.IData)
	channelWatcherF func() ChannelWatcher
	// 收到msg的时候的func
	receiverF     func(id string, v channel.IData)
	sleepInterval int
}

func asValues(n, count int) [][]channel.IData {
	values := make([][]channel.IData, n)
	for i := 0; i < n; i++ {
		values[i] = make([]channel.IData, 10)
		for j := 0; j < 10; j++ {
			values[i][j] = VString("_" + strconv.Itoa(count) + "_" + strconv.Itoa(i*10+j))
		}
	}
	return values
}
func testWithN(t *testing.T, req testN, opts ...models.StartOption) {
	if req.sleepInterval > 0 {
		defaultTestSleepInterval = req.sleepInterval
	} else if req.sleepInterval < 0 {
		defaultTestSleepInterval = math.MaxInt32
	}
	n := req.routineLimit
	count := req.count
	f := req.channelWatcherF
	handlef := req.consumerF
	startT := time.Now()
	values := asValues(n, count)
	watcher := f()
	watcher.BStart(opts...)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(index int) {
			test, member := newMockTest(handlef, req.receiverF, "test"+strconv.Itoa(index), values[index]...)
			watcher.WatchMemberChanged(member)
			nn := time.Now()
			test.BlockWaitPanic()
			cost := time.Now().Sub(nn)
			debugPrint(fmt.Sprintf("一个job结束,耗时:%f秒,%d毫秒", cost.Seconds(), cost.Milliseconds()))
			wg.Done()
		}(i)
	}
	wg.Wait()
	// time.Sleep(time.Second * 1)
	// if watcher.Size() != 0 {
	// 	watcher.PrintSelf()
	// 	time.Sleep(time.Second * 3)
	// 	watcher.PrintSelf()
	// }
	// require.Equal(t, 0, watcher.Size())
	cost := time.Now().Sub(startT)
	logplugin.Warn(fmt.Sprintf("任务结束,耗时:%f秒,%d毫秒", cost.Seconds(), cost.Milliseconds()))
}

func Test_LocalConst(t *testing.T) {

	fmt.Println(routinec_status_upgrade)
	fmt.Println(routinec_status_close)
	fmt.Println(routinec_status_release)
	fmt.Println(routinec_status_before_reuse)
	fmt.Println(routinec_status_wait_listener)
	fmt.Println("==================")
	fmt.Println(region_status_new)
	fmt.Println(region_status_running)
	fmt.Println(region_status_with_blocking)
	fmt.Println(region_status_pre_block)
	fmt.Println(region_status_block)
	fmt.Println(region_status_to_remove)
	fmt.Println(region_status_remove)
	fmt.Println("================")
	fmt.Println(selectn_channel_status_ok)
	fmt.Println(selectn_channel_status_closed)
	fmt.Println(selectn_channel_status_gone)
	fmt.Println(selectn_channel_status_removed)
	fmt.Println(selectn_channel_status_final)
	fmt.Println(selectn_channel_status_reusing)
	fmt.Println(selectn_channel_status_before_reuse)
	fmt.Println(selectn_channel_status_reuse)
	fmt.Println("================")
	fmt.Println(status_ok)
	fmt.Println(status_deny_memchanged)
	fmt.Println(status_close)
	fmt.Println(status_changing)
}

func mockUpgrade(n int, f func() ChannelWatcher, ops ...Option) ChannelWatcher {
	if len(ops) == 0 {
		ops = foreverOptions()
	}
	defaultTestSleepInterval = 300
	watcher := f()
	watcher.BStart()
	values := asValues(n, 1)
	wg := sync.WaitGroup{}
	wg.Add(n)
	need := int32(n)
	unjoined := int32(n)
	// set := make(map[string]struct{})
	for i := 0; i < n; i++ {
		go func(index int) {
			test, member := newMockTest(func(v channel.IData) {
				// time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(200, 2000)))
				time.Sleep(time.Millisecond * 20)
			}, nil, "test"+strconv.Itoa(index), values[index]...)
			for watcher.WatchMemberChanged(member) {
				time.Sleep(time.Millisecond * 100)
			}
			atomic.AddInt32(&unjoined, -1)

			test.BlockWaitPanic()
			wg.Done()
			atomic.AddInt32(&need, -1)
			debugPrint("done", "name", "test"+strconv.Itoa(index), "未收到结果的", need, "unjoined", unjoined)
		}(i)
	}
	res := make(chan ChannelWatcher)
	go func() {
		time.Sleep(time.Second * 2)
		r := watcher.Upgrade(ops...)
		res <- r
	}()
	wg.Wait()
	return <-res
}

func mockRollBack(n int, f func() ChannelWatcher, ops ...Option) ChannelWatcher {
	if len(ops) == 0 {
		ops = foreverOptions()
	}
	defaultTestSleepInterval = 300
	watcher := f()
	watcher.BStart()
	values := asValues(n, 1)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(index int) {
			test, member := newMockTest(func(v channel.IData) {
				time.Sleep(time.Millisecond * 1000)
			}, nil, "test"+strconv.Itoa(index), values[index]...)
			for watcher.WatchMemberChanged(member) {
				time.Sleep(time.Millisecond * 100)
			}

			test.BlockWaitPanic()
			wg.Done()
		}(i)
	}
	res := make(chan ChannelWatcher)
	go func() {
		time.Sleep(time.Second * 1)
		r := watcher.RollBack(ops...)
		res <- r
	}()
	wg.Wait()

	return <-res
}

// test
func newInternalTestReflectChannelWatcher() *reflectChannelWatcher {
	r := newReflectChannelWatcher(DefaultForeverOpt)
	r.rollBackLimit = 0
	return r
}

func mockBaseUse(t *testing.T, f func() ChannelWatcher) {
	values := asValues(1, 1)
	watcher := f()
	watcher.BStart()
	test, member := newMockTest(nil, nil, "test", values[0]...)
	watcher.WatchMemberChanged(member)
	test.BlockWaitPanic()
	time.Sleep(time.Millisecond * 300)
	require.Equal(t, 0, watcher.Size())
}

// not useful
func mockTestDeleta(f func(opt Opt) ChannelWatcher, rollBack bool) {
	opt := DefaultForeverOpt
	opt.SpinTimeMills = 3000
	watcher := f(opt)
	watcher.BStart()
	mems, wp := mockChannels(func(v channel.IData) {
		time.Sleep(time.Second)
	}, 1)
	watcher.WatchMemberChanged(mems[0])
	if rollBack {
		watcher.RollBack(commonOptions()...)
	} else {
		watcher.Upgrade(commonOptions()...)
	}
	wp.BlockWaitPanic()
}

func commonTestSmallJob(t *testing.T, f func() ChannelWatcher) {
	testWithN(t, testN{
		count:        1,
		routineLimit: 100,
		consumerF: func(v channel.IData) {
			time.Sleep(time.Millisecond * time.Duration(rand.RandInt32(100, 1500)))
		},
		channelWatcherF: f,
		receiverF:       nil,
		sleepInterval:   0,
	})
}

func commonTestNCounts(t *testing.T, count int, f func()) {
	for i := 0; i < count; i++ {
		t.Run("count_"+strconv.Itoa(i), func(t *testing.T) {
			f()
		})
	}
}
