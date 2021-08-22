/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/26 9:18 上午
# @File : support_test.go.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/p2p/impl/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func PanicError(err error) {
	if nil != err {
		panic(err)
	}
}

func TestPack(t *testing.T) {
	helper := NewPackerUnPacker()
	data := []byte("123")
	pack, err := helper.Pack(data)
	PanicError(err)
	fmt.Println(string(pack))
	s := &mock.MockStream{}
	_, err = s.Write(pack)
	PanicError(err)
	unPack, err := helper.UnPack(context.Background(), s, nil)
	PanicError(err)
	fmt.Println(unPack)
	require.Equal(t, unPack, data)
}

func Test_Long_Bytes(t *testing.T) {
	fmt.Println(randString(100))
	helper := NewPackerUnPacker()
	data := []byte(randString(10000))
	check(t, helper, data)
}
func check(t *testing.T, helper services.StreamDataSupport, data []byte) {
	pack, err := helper.Pack(data)
	PanicError(err)
	fmt.Println(string(pack))
	s := &mock.MockStream{}
	_, err = s.Write(pack)
	PanicError(err)
	unPack, err := helper.UnPack(context.Background(), s, nil)
	PanicError(err)
	require.Equal(t, unPack, data)
}
func randString(l int) string {
	var bigSlice = make([]byte, 73437, 73437)
	bigSlice[0] = 67
	for j := 1; j < l; j *= 2 {
		copy(bigSlice[j:], bigSlice[:j])
	}
	return string(bigSlice)
}
