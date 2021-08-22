/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/27 7:46 下午
# @File : peer_lifecycle.go
# @Description :
# @Attention :
*/
package models

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"sync"
	"sync/atomic"
	"time"
)

type PeerRecord struct {
	PeerId types2.NodeID
	ExternMultiAddress   string
}


func (p PeerRecord) String() string {
	return "PeerRecord:节点id为:" + p.PeerId.Pretty() + ",地址为:" + p.ExternMultiAddress
}

type PeerLifeCycle struct {
	sync.RWMutex
	PeerId   types2.NodeID
	ExternMultiAddress string
	LastSeen time.Time
	Status   uint32
}

func (this *PeerLifeCycle) Alive() bool {
	return atomic.LoadUint32(&this.Status) == p2ptypes.PEER_STATUS_ALIVE
}


type NewPeerInfo struct {
	ID types2.NodeID
	ExternMultiAddress string
}