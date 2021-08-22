/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/5 8:25 上午
# @File : internal.go
# @Description :
# @Attention :
*/
package services

import (
	"context"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/pubsub/models"
)

// IInternalPubSubComponent
type ICommonEventBusComponent interface {
	base.IComponent
	Subscribe(ctx context.Context, clientID string, query Query, outCapacity ...int) (Subscription, error)
	PublishWithEvents(ctx context.Context, msg interface{}, events map[string][]string) error
	SubscribeUnbuffered(ctx context.Context, clientID string, query Query) (*models.SubscriptionImpl, error)
	Unsubscribe(ctx context.Context, subscriber string, query Query) error
	UnsubscribeAll(ctx context.Context, subscriber string) error

	NumClients() int
	NumClientSubscriptions(clientID string) int
}
type Subscription interface {
	Out() <-chan models.PubSubMessage
	Canceled() <-chan struct{}
	Cancel(err error)
	Err() error
}
type Query interface {
	Matches(events map[string][]string) (bool, error)
	String() string
}
