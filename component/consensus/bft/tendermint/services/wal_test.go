/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 9:13 上午
# @File : wal_test.go.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	auto "github.com/hyperledger/fabric-droplib/libs/autofiles"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Wal(t *testing.T) {
	walFile := "wal"
	w, err := NewWAL(walFile, func(group *auto.Group) {
		group.BaseServiceImpl = impl.NewBaseService(nil, modules.NewModule("TEST_WAL", 1), group)
	})
	require.NoError(t, err)
	err = w.Start()
	require.NoError(t, err)
	s := struct {
		name string `json:"name"`
		age  int    `json:"age"`
	}{}
	s.name = "joker"
	s.age = 23
	msg := WALMessage(s)
	err = w.Write(msg)
	require.NoError(t, err)
}
