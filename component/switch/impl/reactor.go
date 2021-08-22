/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 11:52 下午
# @File : reactor.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
)

var (
	_ services.IReactor = (*BaseReactor)(nil)
	_ services.IReactor = (*NoneSwitchReactor)(nil)
)

type BaseReactor struct {
	*impl.BaseServiceImpl
	Id     types.ReactorId
	Switch services.ISwitch
}

func NewBaseReactor(m modules.Module, id types.ReactorId, p services.ISwitch) *BaseReactor {
	r := &BaseReactor{
		Id:     id,
		Switch: p,
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, m, r)

	return r
}

func (b *BaseReactor) GetReactorId() types.ReactorId {
	return b.Id
}

func (b *BaseReactor) SetSwitch(iSwitch services.ISwitch) {
	b.Switch = iSwitch
}

type NoneSwitchReactor struct {
	*impl.BaseServiceImpl
	Id types.ReactorId
}

func NewNonSwitchReactor(m modules.Module,rr services2.IBaseService) *NoneSwitchReactor {
	r := &NoneSwitchReactor{
		Id: types.ReactorId(m.Index()),
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, m, rr)

	return r
}

func (n *NoneSwitchReactor) GetReactorId() types.ReactorId {
	return n.Id
}

func (n *NoneSwitchReactor) SetSwitch(iSwitch services.ISwitch) {
}
