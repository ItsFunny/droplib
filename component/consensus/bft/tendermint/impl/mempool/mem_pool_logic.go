/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 10:43 下午
# @File : mem_pool_logic.go
# @Description :
# @Attention :
*/
package mempool

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	streamimpl "github.com/hyperledger/fabric-droplib/component/p2p/impl"
)

var (
	_ services.IMemPoolLogicService = (*defaultMemPoolLogicServiceImpl)(nil)
)

// 分发交易
type defaultMemPoolLogicServiceImpl struct {
	*streamimpl.BaseLogicServiceImpl

	memPool services.IMemPool
}

func NewDefaultMemPoolLoigcService(mp services.IMemPool) *defaultMemPoolLogicServiceImpl {
	r := &defaultMemPoolLogicServiceImpl{}
	// r.BaseLogicServiceImpl=streamimpl.NewBaseLogicServiceImpl()
	r.BaseLogicServiceImpl = streamimpl.NewBaseLogicServiceImpl(modules.MODULE_MEM_POOL, r)
	r.memPool = mp

	return r
}

func (d *defaultMemPoolLogicServiceImpl) MemPool() {
}

func (d *defaultMemPoolLogicServiceImpl) GetServiceId() types.ServiceID {
	return MEMPOOL_LOGICSERVICE_ID
}

func (d *defaultMemPoolLogicServiceImpl) ChoseInterestProtocolsOrTopics(pProtocol *models2.P2PProtocol) error {
	return nil
}

func (d *defaultMemPoolLogicServiceImpl) ChoseInterestEvents(bus base.IP2PEventBusComponent) {
}
