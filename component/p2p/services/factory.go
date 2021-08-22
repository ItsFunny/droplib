/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 1:41 下午
# @File : factory.go
# @Description :
# @Attention :
*/
package services

import (
	"context"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/libp2p/go-libp2p"
)

type INodeFactory interface {
	CreateNode(ctx context.Context,cfg *config.P2PConfiguration, opts ...libp2p.Option) INode
}
