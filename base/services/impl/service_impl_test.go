/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/8 6:43 下午
# @File : service_impl_test.go.go
# @Description :
# @Attention :
*/
package impl

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/constants"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type A struct {
	*BaseServiceImpl
}

func (this *A) OnStop(ctx *models.StopCTX) {
	fmt.Println("stop")
}

func newA() *A {
	r := &A{}
	r.BaseServiceImpl = NewBaseService(nil, modules.NewModule("123", 1), r)
	return r
}

func Test_Async_Start_Stop(t *testing.T) {
	a := newA()
	a.BStart(models.AsyncStartWaitReadyOpt)
	time.Sleep(time.Second * 5)
	a.BReady(models.ReadyWaitStartOpt)
	// a.BStop(models.StopCTXWithForce)

	require.Equal(t, a.stopped, uint32(constants.STOP))
	v := a.Quit()
	require.NotNil(t, v)
}
