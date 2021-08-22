/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 11:26 下午
# @File : reactor.go
# @Description :
# @Attention :
*/
package mempool

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/switch/impl"
)

var (
	_ services.IMemPoolReactor = (*MemPoolReactor)(nil)
)

type MemPoolReactor struct {
	*impl.BaseReactor
}
