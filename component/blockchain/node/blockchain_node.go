/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/19 3:45 下午
# @File : blockchain_node.go
# @Description :
# @Attention :
*/
package node

import (
	impl2 "github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/component/blockchain/impl"
	"github.com/hyperledger/fabric-droplib/component/blockchain/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	p2p "github.com/hyperledger/fabric-droplib/component/p2p/cmd"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
)

type Node struct {
	*impl2.BaseServiceImpl
	Signer cryptolibs.ISigner

	P2P *p2p.P2P

	blockChainService services.IBlockChainLogicService
}
func NewNode(cfg *config.GlobalConfiguration) *Node {
	r := &Node{}
	r.Signer = &cryptolibs.MockSigner{}
	r.P2P = p2p.NewP2P(cfg.P2PConfiguration)
	r.blockChainService = impl.NewBlockChainLogicServiceImpl(cfg, r.Signer)
	return r
}
