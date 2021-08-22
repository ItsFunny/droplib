/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 1:46 下午
# @File : peer.go
# @Description :
# @Attention :
*/
package events

import (
	"github.com/hyperledger/fabric-droplib/base/types"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
)

type EventPeerUpdate struct {
	PeerId             types.NodeID
	ExternMultiAddress string
	Type               p2ptypes.MEMBER_CHANGED
}

func (e EventPeerUpdate) String() string {
	return ""
}
