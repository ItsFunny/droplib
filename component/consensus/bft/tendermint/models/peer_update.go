/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:39 上午
# @File : peer_update.go
# @Description :
# @Attention :
*/
package models

import (
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"sync"
)

// PeerStatus is a peer status.
//
// The peer manager has many more internal states for a peer (e.g. dialing,
// connected, evicting, and so on), which are tracked separately. PeerStatus is
// for external use outside of the peer manager.
type PeerStatus string

// PeerUpdate is a peer update event sent via PeerUpdates.
type PeerUpdate struct {
	NodeID types2.NodeID
	Status PeerStatus
}

// PeerUpdates is a peer update subscription with notifications about peer
// events (currently just status changes).
type PeerUpdates struct {
	updatesCh chan PeerUpdate
	closeCh   chan struct{}
	closeOnce sync.Once
}
