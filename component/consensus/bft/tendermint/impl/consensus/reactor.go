/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 11:02 上午
# @File : reactor.go
# @Description :
# @Attention :
*/
package consensus

import (
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/switch/impl"
)

var (
	consensusReactorId                            = types.ReactorId(0x02)
	_                  services.IConsensusReactor = (*ConsensusReactor)(nil)
)

type ConsensusReactor struct {
	*impl.BaseReactor
}
