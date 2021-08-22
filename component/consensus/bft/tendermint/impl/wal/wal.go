/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 4:16 下午
# @File : wal.go
# @Description :
# @Attention :
*/
package wal

import (
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"io"
)

type NilWAL struct{}

var _ services.IWAL = NilWAL{}

func (NilWAL) Write(m services.WALMessage) error     { return nil }
func (NilWAL) WriteSync(m services.WALMessage) error { return nil }
func (NilWAL) FlushAndSync() error          { return nil }
func (NilWAL) SearchForEndHeight(height uint64, options *services.WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	return nil, false, nil
}
func (NilWAL) Start(flag services2.START_FLAG) error { return nil }
func (NilWAL) Stop() error  { return nil }
func (NilWAL) Wait()        {}
