/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 5:41 下午
# @File : tx.go
# @Description :
# @Attention :
*/
package models

import (
	"context"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
)

// TxInfo are parameters that get passed when attempting to add a tx to the
// mempool.
type TxInfo struct {
	// SenderID is the internal peer ID used in the mempool to identify the
	// sender, storing 2 bytes with each tx instead of 20 bytes for the p2p.ID.
	SenderID uint16
	// SenderP2PID is the actual p2p.ID of the sender, used e.g. for logging.
	SenderP2PID types2.NodeID
	// Context is the optional context to cancel CheckTx
	Context context.Context
}
