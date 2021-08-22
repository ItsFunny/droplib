/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 10:06 上午
# @File : blockchain.go
# @Description :
# @Attention :
*/
package services

import (
	"context"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/blockchain/models"
)

type IBlockChainLogicComponent interface {
	base.ILogicComponent
}
type IBlockChainLogicService interface {
	base.ILogicService
	Propose(ctx context.Context, data []byte) error

	Commit() chan<- models.Entry
}
