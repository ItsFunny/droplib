/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 10:06 上午
# @File : blockchain.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/blockchain/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	services2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	p2p "github.com/hyperledger/fabric-droplib/component/p2p/cmd"
)

var (
	_ services.IBlockChainLogicComponent = (*blockChainComponent)(nil)
)

type blockChainComponent struct {
	*base.BaseLogicComponent
	cfg                   *config.GlobalConfiguration
	p2p                   *p2p.P2P
	consensusLogicService services2.IConsensusLogicService
}

func NewBlockChainComponent(cfg *config.GlobalConfiguration) *blockChainComponent {
	r := &blockChainComponent{}
	r.BaseLogicComponent = base.NewBaseLogicCompoennt(modules.COMPONENT_BLOCKCHAIN, r)
	r.p2p=p2p.NewP2P(cfg.P2PConfiguration)

	return r
}

func (b *blockChainComponent) GetBoundServices() []base.ILogicService {
	panic("implement me")
}
