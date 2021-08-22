/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 10:20 上午
# @File : p2p.go
# @Description :
# @Attention :
*/
package reactor

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	p2p "github.com/hyperledger/fabric-droplib/component/p2p/cmd"
	"github.com/hyperledger/fabric-droplib/component/switch/impl"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
	"github.com/libp2p/go-libp2p"
)

var (
	_ IP2PReactor = (*P2PReactor)(nil)

	p2pId = types.ReactorId(0x01)
)

type P2PReactor struct {
	*impl.BaseReactor
	p2p *p2p.P2P
}

func NewP2PReactor(s services.ISwitch, cfg *config.P2PConfiguration, opts ...libp2p.Option) services.IReactor {
	r := &P2PReactor{}
	r.BaseReactor = impl.NewBaseReactor(modules.P2P_REACTOR_MODULE, p2pId, s)
	r.p2p = p2p.NewP2P(cfg, opts...)
	return r
}
