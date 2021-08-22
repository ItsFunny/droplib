/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/26 9:32 上午
# @File : stream_helper.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
)

var (
	_ services2.StreamDataSupport = (*PackerUnpacker)(nil)
	_ services2.StreamUnPacker    = (*HeadSizeStreamUnpackerImpl)(nil)
	_ services2.StreamPacker      = (*HeadSizePacker)(nil)
)

type PackerUnpacker struct {
	packer   services2.StreamPacker
	unpacker services2.StreamUnPacker
}

func (this *PackerUnpacker) GetStreamPacker() services2.StreamPacker {
	return this.packer
}

func (this *PackerUnpacker) Pack(data []byte) ([]byte, error) {
	return Pack(this.packer, data)
}
func (this *PackerUnpacker) UnPack(ctx context.Context, stream services2.IStream, lastData []byte) ([]byte, error) {
	res := make(chan interface{})
	go func() {
		unpack, err := Unpack(this.unpacker, stream)
		if nil != err {
			res <- err
		} else {
			res <- unpack
		}
	}()
	select {
	case v := <-res:
		if err, ok := v.(error); ok {
			return nil, err
		}
		return v.([]byte), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func NewPackerUnPacker() services2.StreamDataSupport {
	r := &PackerUnpacker{
		packer:   DefaultNewStreamPacker(),
		unpacker: DefaultNewStreamUnPacker(),
	}
	return r
}

func DefaultNewStreamPacker() services2.StreamPacker {
	_1 := &NoOpPacker{BaseLinkedList: impl.NewBaseLinkedList()}
	_2 := &HeadSizePacker{BaseLinkedList: impl.NewBaseLinkedList()}
	return NewStreamPacker(_1, _2)
}
func DefaultNewStreamUnPacker() services2.StreamUnPacker{
	_1 := &HeadSizeStreamUnpackerImpl{BaseLinkedList: impl.NewBaseLinkedList()}
	_2 := &DataStreamUnPackerImpl{BaseLinkedList: impl.NewBaseLinkedList()}
	return NewStreamUnpacker(_1, _2)
}
func NewStreamPacker(packers ...services2.StreamPacker) services2.StreamPacker {
	l := len(packers)
	_1 := packers[0]
	for i := 1; i < l; i++ {
		_1 = services2.LinkLast(_1, packers[i]).(services2.StreamPacker)
	}
	return _1
}
func NewStreamUnpacker(unpackers ...services2.StreamUnPacker) services2.StreamUnPacker {
	l := len(unpackers)
	_1 := unpackers[0]
	for i := 1; i < l; i++ {
		_1 = services2.LinkLast(_1, unpackers[i]).(services2.StreamUnPacker)
	}
	return _1
}

func Pack(packer services2.StreamPacker, bytes []byte) ([]byte, error) {
	var res []byte
	var err error
	err = services2.Iterator(packer, func(node services2.ILinkedList) error {
		v := node.(services2.StreamPacker)
		if res, err = v.Pack(bytes); nil != err {
			return err
		}
		return nil
	})
	return res, err
}

func Unpack(unpacker services2.StreamUnPacker, stream services2.IStream) ([]byte, error) {
	var lastData []byte
	var err error
	err = services2.Iterator(unpacker, func(node services2.ILinkedList) error {
		n := node.(services2.StreamUnPacker)
		lastData, err = n.UnPack(stream, lastData)
		if nil != err {
			return err
		}
		return nil
	})
	return lastData, err
}

// //////////////////// packer
type HeadSizePacker struct {
	*impl.BaseLinkedList
}

func (h *HeadSizePacker) Pack(bytes []byte) ([]byte, error) {
	l := len(bytes)
	sizeBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuf, uint64(l))
	res := make([]byte, l+8)
	copy(res[:8], sizeBuf)
	copy(res[8:], bytes)
	return res, nil
}

type NoOpPacker struct {
	*impl.BaseLinkedList
}

func (h *NoOpPacker) Pack(bytes []byte) ([]byte, error) {
	return bytes, nil
}

// ////////////// unpacker
type HeadSizeStreamUnpackerImpl struct {
	*impl.BaseLinkedList
}



type DataStreamUnPackerImpl struct {
	*impl.BaseLinkedList
}

func (s *DataStreamUnPackerImpl) UnPack(stream services2.IStream, lastData []byte) ([]byte, error) {
	size := binary.BigEndian.Uint64(lastData)
	reader := bufio.NewReader(stream)
	dataBuf := make([]byte, size)
	if l, err := reader.Read(dataBuf); nil != err {
		return nil, err
	} else if l != int(size) {
		return nil, errors.New("stream data not correct")
	}
	return dataBuf, nil
}
