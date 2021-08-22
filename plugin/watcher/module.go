/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/22 8:44 上午
# @File : module.go
# @Description :
# @Attention :
*/
package watcher

import "github.com/hyperledger/fabric-droplib/base/log/modules"

var (
	routineModule = modules.NewModule("routine", 1)
	reflectModule = modules.NewModule("reflect", 1)
	selectnModule = modules.NewModule("selectn", 1)
)
