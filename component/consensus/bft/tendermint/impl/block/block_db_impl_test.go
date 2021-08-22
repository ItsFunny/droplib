/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:45 上午
# @File : block_db_impl_test.go.go
# @Description :
# @Attention :
*/
package block

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDBWrite(t *testing.T) {
	db, err := NewGoLevelDB("test", ".")
	require.NoError(t, err)
	k := []byte("123")
	v := []byte("456")
	err = db.Set(k, v)
	require.NoError(t, err)
}

func Test_Set_Get(t *testing.T) {
	db, err := NewGoLevelDB("test", ".")
	require.NoError(t, err)
	k := []byte("456")
	v := []byte("12331231")
	err = db.Set(k, v)
	get, err := db.Get(k)
	require.NoError(t, err)
	require.Equal(t, get, v)
}
