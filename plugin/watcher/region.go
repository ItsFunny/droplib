/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/10 7:07 上午
# @File : region.go
# @Description :
# @Attention :
*/
package watcher

const (
	status_ok              = 1 << 0                        // 1
	status_deny_memchanged = 1 << 1                        // 2
	status_close           = 1<<2 | status_deny_memchanged // 6
	// status_ready_to_upgrade = 1<<3 | status_deny_memchanged
	status_changing = 1<<4 | status_deny_memchanged // 18

	status_on_update = 1 << 3
	status_on_gc     = 1 << 4

	halt_upgrade         = 1 << 0
	upgrade_finish       = 1 << 1
	halt_upgrade_reflect = 1<<2 | halt_upgrade
)
