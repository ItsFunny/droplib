/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 9:11 上午
# @File : constants.go
# @Description :
# @Attention :
*/
package consensus

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	LOGICSERVICE_CONSENSUS types2.ServiceID = "CONSENSUS"
)

var (
	CONSENSUS_PROTOCOL_ID protocol.ID = "/consensus/1.0.0"
)

const (
	StateChannel       = types2.ChannelID(0x20)
	DataChannel        = types2.ChannelID(0x21)
	VoteChannel        = types2.ChannelID(0x22)
	VoteSetBitsChannel = types2.ChannelID(0x23)
)

const (
	maxMsgSize = 1048576 // 1MB; NOTE: keep in sync with types.PartSet sizes.

	blocksToContributeToBecomeGoodPeer = 10000
	votesToContributeToBecomeGoodPeer  = 10000

	listenerIDConsensus = "consensus-reactor"
)
