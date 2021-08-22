/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/29 7:13 下午
# @File : support.go
# @Description :
# @Attention :
*/
package services

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/services"
)

type StreamDataSupport interface {
	Pack([]byte) ([]byte, error)
	UnPack(ctx context.Context, stream IStream, lastData []byte) ([]byte, error)
}

type StreamPacker interface {
	Pack([]byte) ([]byte, error)
	services.ILinkedList
}

type StreamUnPacker interface {
	UnPack(stream IStream,lastData []byte) ([]byte, error)
	services.ILinkedList
}

