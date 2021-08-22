/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/29 7:13 下午
# @File : support.go
# @Description :
# @Attention :
*/
package packer

import (
	"context"
	"github.com/hyperledger/fabric-droplib/component/base"
)

type StreamDataSupport interface {
	Pack([]byte) ([]byte, error)
	UnPack(ctx context.Context, stream base.IStream, lastData []byte) ([]byte, error)
}

type StreamPacker interface {
	Pack([]byte) ([]byte, error)
}

type StreamUnPacker interface {
	UnPack(stream base.IStream, lastData []byte) ([]byte, error)
}

