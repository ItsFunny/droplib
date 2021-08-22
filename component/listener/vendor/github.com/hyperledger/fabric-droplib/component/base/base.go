/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:23 上午
# @File : base.go
# @Description :
# @Attention :
*/
package base

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"strings"
)

type IComponent interface {
	services2.IBaseService
}

type BaseComponent struct {
	*impl.BaseServiceImpl
}

func NewBaseComponent(m modules.Module, i services2.IBaseService) *BaseComponent {
	r := &BaseComponent{
		BaseServiceImpl: nil,
	}
	name := strings.ToUpper(m.String())
	if !strings.Contains(name, "_COMPONENT") {
		name = name + "_COMPONENT"
		m = modules.NewModule(name, m.Index())
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, m, i)
	return r
}
