/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 9:05 上午
# @File : stream.go
# @Description :
# @Attention :
*/
package p2ptypes

import (
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/libp2p/go-libp2p-core/protocol"
)

// //////////// stream
const (
	STREAM_STATUS_ALIVE = 1 << 0

	STREAM_STATUS_CLOSE_READ  = 1 << 1
	STREAM_STATUS_CLOSE_WRITE = 1 << 2
	STREAM_STATUS_RESET       = 1 << 3
	STREAM_STATUS_CLOSE_ALL   = STREAM_STATUS_CLOSE_READ | STREAM_STATUS_CLOSE_WRITE

	STREAM_STATUS_SWEEP = 1 << 6

	STREAM_STATUS_NOT_HEALTHY = STREAM_STATUS_CLOSE_ALL | STREAM_STATUS_SWEEP | STREAM_STATUS_RESET
)

// ////////////// protocol
const (
	PROTOCOL_STATUS_AVALIABLE = 1
)

// ////////////// peer status
const (
	PEER_STATUS_ALIVE = 1
)

// //////////// router status
const (
// ROUTER_STATUS_FINISH_SETUP = 1
)

// /////////// logic_service
const (
	LOGIC_PING_PONG types.ServiceID = "ping_pong"
)

const (
	// topic
	TOPIC_LIFE_CYCLE = "PING_PONG"

	// protocol
	PROTOCOL_PING_PONG protocol.ID = "/ping/1.0.0"

	// channel
	CHANNEL_ID_LIFE_CYCLE types.ChannelID = 0x01
)

const (
	MANAGER_READY = 1
)

type MEMBER_CHANGED byte

// / member
const (
	MEMBER_CHANGED_ADD MEMBER_CHANGED = 0
)
