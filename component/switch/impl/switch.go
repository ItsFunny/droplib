/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 11:52 下午
# @File : switch.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/switch/services"
	libsync "github.com/hyperledger/fabric-droplib/libs/sync"
)

var (
	_ services.ISwitch = (*Switch)(nil)
)

type Switch struct {
	*impl.BaseServiceImpl
	libsync.RWMutex

	reactors map[types.ReactorId]services.IReactor
}

func (s *Switch) SwitchTo(id types.ReactorId) services.IReactor {
	return s.reactors[id]
}

func (s *Switch) RegisterReactor(reactor services.IReactor) {
	s.Lock()
	defer s.Unlock()
	id := reactor.GetReactorId()
	if _, exist := s.reactors[id]; exist {
		s.Logger.Error("重复注册", "id为:", id)
		panic("重复注册reactor")
	}
	s.reactors[id] = reactor
}
