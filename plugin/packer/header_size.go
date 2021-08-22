/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 3:27 下午
# @File : header_size.go
# @Description :
# @Attention :
*/
package packer

import (
	"encoding/binary"
	"errors"
	"github.com/hyperledger/fabric-droplib/component/base"
)

var (
	_ StreamPacker   = (*FuncPackerUnPacker)(nil)
	_ StreamUnPacker = (*FuncPackerUnPacker)(nil)
)

type FuncPackerUnPacker struct {
	pack   func(bytes []byte) ([]byte, error)
	unPack func(stream base.IStream, lastData []byte) ([]byte, error)
}

func (f *FuncPackerUnPacker) UnPack(stream base.IStream, lastData []byte) ([]byte, error) {
	return f.unPack(stream, lastData)
}

func (f *FuncPackerUnPacker) Pack(bytes []byte) ([]byte, error) {
	return f.pack(bytes)
}

var HEAD_SIZE_PACK=func(bytes []byte) ([]byte, error){
	l := len(bytes)
	sizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuf, uint64(l))
	res := make([]byte, l+8)
	copy(res[:8], sizeBuf)
	copy(res[8:], bytes)
	return res, nil
}
var HEAD_SIZE_UNPACK=func(stream base.IStream, lastData []byte) ([]byte, error){
	sizeBuf := make([]byte, 8)
	if l, err := stream.Read(sizeBuf); nil != err {
		return nil, err
	} else if l != 8 {
		return nil, errors.New("wrong structural data ")
	}
	return sizeBuf, nil
}