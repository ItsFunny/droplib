/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/10 1:09 下午
# @File : operation.go
# @Description :
# @Attention :
*/
package watcher

import (
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"sync/atomic"
)

const (
	op_new_member = iota
	op_routine_gone
	op_region_remove
	op_rollback
	op_upgrade
	op_reuse

	op_region_create
)

type IOperationData interface {
	channel.IData
}
type operationWrapper struct {
	id int32
	data channel.IData
}


type operation struct {
	id     int32
	opType byte
	data   IOperationData
}


var (
	opId int32
)

func acquireOperation(opType byte, data IOperationData) operation {
	return operation{
		id:     atomic.AddInt32(&opId, 1),
		opType: opType,
		data:   data,
	}
}



type routineCClose struct {
	key string
}

func (r routineCClose) ID() interface{} {
	return r.key
}



type flushWrapper struct {
	v channel.IData
}