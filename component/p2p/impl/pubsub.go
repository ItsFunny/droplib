/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 11:59 上午
# @File : pubsub.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	services2 "github.com/hyperledger/fabric-droplib/component/p2p/services"
	"github.com/hyperledger/fabric-droplib/component/pubsub/impl"
	"github.com/hyperledger/fabric-droplib/component/pubsub/services"
	"github.com/libp2p/go-libp2p-core/host"
)

var (
	_ services2.IPubSubManager = (*pubsubManagerImpl)(nil)
)

type pubsubManagerImpl struct {
	services.IPubSubComponent
}

func NewLibP2PPubSubManager(ctx context.Context, host host.Host) *pubsubManagerImpl {
	r := &pubsubManagerImpl{}
	r.IPubSubComponent = impl.NewLibP2PPubSubComponentImpl(ctx, host)
	return r
}
