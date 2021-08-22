/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/20 9:29 上午
# @File : opt.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/component/routine/worker/gowp/workpoolv0"
)

type Opt func(r *RoutineComponent)

func Cap(cap int) Opt {
	return func(r *RoutineComponent) {
		r.pool = workpoolv0.New(cap)
	}
}
