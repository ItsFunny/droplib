/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 10:33 上午
# @File : event_manager_component_test.go.go
# @Description :
# @Attention :
*/
package v2

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	listener1 = "listener1"
	listener2 = "listener2"

	event_1 EventNameSpace = "event_1"
	event_2 EventNameSpace = "event_2"
)

// listener1 监听 event_1 和event_2
// listener2 监听 event_2
func Test_EventComponent(t *testing.T) {
	com := NewDefaultEventComponent()
	l1Count := 0
	l2Count := 0
	com.RegisterListenerForEvent(listener1, event_1, func(data EventData) {
		fmt.Println("listener:"+listener1+"event:"+string(event_1), data)
		l1Count++
	})
	com.RegisterListenerForEvent(listener1, event_2, func(data EventData) {
		fmt.Println("listener:"+listener1+"event:"+string(event_2), data)
		l1Count++
	})
	com.RegisterListenerForEvent(listener2, event_2, func(data EventData) {
		fmt.Println("listener:"+listener2+"event:"+string(event_2), data)
		l2Count++
	})

	com.Fire(DefaultEventWrapper{
		NameSpace: event_1,
		EventData: "data_event_1",
	})
	com.Fire(DefaultEventWrapper{
		NameSpace: event_2,
		EventData: "data_event_2",
	})

	require.Equal(t, l1Count, 2)
	require.Equal(t, l2Count, 1)
}
