/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/12 2:53 下午
# @File : constants.go
# @Description :
# @Attention :
*/
package watcher

import "github.com/hyperledger/fabric-droplib/libs/channel"

const (
	mode_debug = iota+1
)
const (
	DEFAULT_STEP_ONE_LIMIT  = 16
	DEFAULT_STEP_TWO_LIMIT  = 256
	DEFAULT_WAIT_MILL_TIMES = 100
)

const (
	routinec_status_ok            = 1 << 0                             // 1
	routinec_status_unavailable   = 1 << 1                             // 2
	routinec_status_upgrade       = 1<<2 | routinec_status_unavailable // 6
	routinec_status_close         = 1<<3 | routinec_status_unavailable // 10
	routinec_status_release       = 1<<4 | routinec_status_close       // 26
	routinec_status_before_reuse  = 1<<5 | routinec_status_unavailable	// 34
	routinec_status_wait_listener = 1<<6 | routinec_status_upgrade	// 70
	routinec_status_running=1<<7 | routinec_status_ok
)

const (
	route_update_notify_chid channel.ChannelID = "updateNotifyC"

	upgradeRollbackNotifyC channel.ChannelID = "upgradeRollbackNotifyC"

	memberNotifyC channel.ChannelID = "memberNotifyC"

	reflect_notifyc channel.ChannelID = "reflect_notifyc"

	reuse_notifyc channel.ChannelID = "reuse_notifyc"

	selectn_notifyC       channel.ChannelID = "notifyC"
	selectn_region_notify channel.ChannelID = "selectn_region_notify"
)

var (
	config_delta_members = "config_delta_members"
)
var (
	listener_upgrade  = "upgrade"
	listener_rollback = "rollback"
)

const (
	default_routine_pool_size = 256
	default_channel_cap       = 20
)
