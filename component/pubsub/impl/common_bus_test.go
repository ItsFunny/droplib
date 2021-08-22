/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 9:38 上午
# @File : internal_pubsub_test.go.go
# @Description :
# @Attention :
*/
package impl

import (
	"context"
	"encoding/json"
	"fmt"
	models2 "github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/pubsub/models"
	"github.com/hyperledger/fabric-droplib/component/pubsub/services"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var EventTypeKey = "demo.event"
var (
	event_type_demo_1 = "demo_1"
	event_type_demo_2 = "demo_2"

	query_1 = QueryForEvent(event_type_demo_1)
	query_2 = QueryForEvent(event_type_demo_2)
)

func QueryForEvent(eventType string) services.Query {
	return models.MustParse(fmt.Sprintf("%s='%s'", EventTypeKey, eventType))
}

var (
	subscriber1 = "subscriber1"
	subscriber2 = "subscriber2"
)

func Test_CommonBus(t *testing.T) {
	bus := NewCommonEventBusComponentImpl(BufferCapacity(10))
	bus.BStart(models2.AsyncStartOpt)

	ctx := context.Background()

	q11C := 0
	s11, err := bus.Subscribe(ctx, subscriber1, query_1, 10)
	require.NoError(t, err)
	go func() {
		for {
			select {
			case msg := <-s11.Out():
				v, _ := json.Marshal(msg.Events())
				q11C++
				println(subscriber1 + ",s11,events:" + string(v) + ",data:" + fmt.Sprintf("%v", msg.Data()))
			}
		}
	}()

	q12C := 0
	s12, err := bus.Subscribe(ctx, subscriber1, query_2, 10)
	require.NoError(t, err)
	go func() {
		for {
			select {
			case msg := <-s12.Out():
				v, _ := json.Marshal(msg.Events())
				q12C++
				println(subscriber1 + ",s12,events:" + string(v) + ",data:" + fmt.Sprintf("%v", msg.Data()))
			}
		}
	}()

	q21C := 0
	s22, err := bus.Subscribe(ctx, subscriber2, query_2, 10)
	require.NoError(t, err)
	go func() {
		for {
			select {
			case msg := <-s22.Out():
				q21C++
				v, _ := json.Marshal(msg.Events())
				println(subscriber2 + ",s21,events:" + string(v) + ",data:" + fmt.Sprintf("%v", msg.Data()))
			}
		}
	}()

	msg := "one_event"
	err = bus.PublishWithEvents(ctx, msg, map[string][]string{EventTypeKey: []string{event_type_demo_1}})
	require.NoError(t, err)
	msg = "all_event"
	err = bus.PublishWithEvents(ctx, msg, map[string][]string{EventTypeKey: []string{event_type_demo_1, event_type_demo_2}})

	time.Sleep(time.Second * 1)
	require.Equal(t, 2, q11C)
	require.Equal(t, 1, q12C)
	require.Equal(t, 1, q21C)
}
