/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/30 9:32 上午
# @File : pingpong.go
# @Description :
# @Attention :
*/
package events

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"strings"
)

type LifeCyclePeerEvent struct {
	Event interface{}
}

type MemberInfoEvent struct {
	Members []MemberInfoNode
}

func (this MemberInfoEvent) String() string {
	sts := make([]string, 0)
	for _, mem := range this.Members {
		sts = append(sts, mem.ExternMultiAddress)
	}
	return strings.Join(sts, ",")
}

type MemberInfoNode struct {
	ExternMultiAddress string
}

type PingPongPeerEvent struct {
	PeerId types2.NodeID
}
