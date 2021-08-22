/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/6 6:30 下午
# @File : channel_test.go
# @Description :
# @Attention :
*/
package watcher

import (
	"fmt"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/common/models"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	"strconv"
	"testing"
	"time"
)

// one routine
// 测试 routine的形式,能否收到数据
func Test_Routine(t *testing.T) {
	c := NewMockConsumer()
	pro := NewDefaultChannelWatcherProperty()
	watcher := NewChannelWatcher(c, pro)
	c1 := make(chan models.Envelope, 20)
	cs := make([]<-chan models.Envelope, 0)
	cs = append(cs, c1)
	g1 := &ChannelGroup{
		Group:       1,
		Name:        "group_1",
		OutChannels: cs,
		OnClose: func() {
			fmt.Println("closing")
		},
	}
	go func() {
		watcher.WatchNewChannelGroup(g1)
	}()
	// produce
	go func() {
		for i := 0; i < 10; i++ {
			c1 <- models.Envelope{
				From:      "123",
				To:        "456",
				Broadcast: "",
				Message:   []byte("456"),
			}
			time.Sleep(time.Second * 1)
		}
	}()

	for {
		select {}
	}
}

// two producer
func Test_TwoProducer(t *testing.T) {
	c := NewMockConsumer()
	pro := NewDefaultChannelWatcherProperty()
	watcher := NewChannelWatcher(c, pro)
	c1 := make(chan models.Envelope, 20)
	cs := make([]<-chan models.Envelope, 0)
	cs = append(cs, c1)
	g1 := &ChannelGroup{
		Group:       1,
		Name:        "group_1",
		OutChannels: cs,
		OnClose: func() {
			fmt.Println("closing")
		},
	}
	go func() {
		watcher.WatchNewChannelGroup(g1)
	}()

	// produce
	for i := 0; i < 2; i++ {
		go func(index int) {
			for i := 0; i < 10; i++ {
				c1 <- models.Envelope{
					From:      types2.NodeID(strconv.Itoa(index)),
					To:        types2.NodeID("TO_" + strconv.Itoa(index)),
					Broadcast: "",
					Message:   []byte(strconv.Itoa(index)),
				}
				time.Sleep(time.Second * 1)
			}
		}(i)
	}

	for {
		select {}
	}
}

// MULTI CHANNELS WITH MORE PRODUCER
func Test_More(t *testing.T) {
	c := NewMockConsumer()
	pro := NewDefaultChannelWatcherProperty()
	watcher := NewChannelWatcher(c, pro)
	cs := make([]<-chan models.Envelope, 0)
	allCs := make([]chan models.Envelope, 0)
	for i := 0; i < 1000; i++ {
		c1 := make(chan models.Envelope, 5)
		cs = append(cs, c1)
		allCs = append(allCs, c1)
	}

	g1 := &ChannelGroup{
		Group:       1,
		Name:        "group_1",
		OutChannels: cs,
		OnClose: func() {
			fmt.Println("closing")
		},
	}
	go func() {
		watcher.WatchNewChannelGroup(g1)
	}()

	// produce
	for i := 0; i < 10; i++ {
		go func(index int) {
			for i := 0; i < 10; i++ {
				intn := rand.Intn(1000)
				c1 := allCs[intn]
				c1 <- models.Envelope{
					From:      types2.NodeID(strconv.Itoa(index)),
					To:        types2.NodeID("TO_" + strconv.Itoa(index)),
					Broadcast: "",
					Message:   []byte(strconv.Itoa(index)),
				}
				time.Sleep(time.Second * 1)
			}
		}(i)
	}

	for {
		select {}
	}
}
