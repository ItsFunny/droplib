/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/29 10:38 上午
# @File : stream.go
# @Description :
# @Attention :
*/
package protos

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
)

var (
	_ services.IProtocolWrapper = (*StreamCommonProtocolWrapper)(nil)
	_ services.IStreamObject    = (*StreamDataMessageWrapper)(nil)
	_ services.IDataWrapper     = (*StreamDataMessageWrapper)(nil)
)

func (s *StreamDataMessageWrapper) GetData() []byte {
	return s.Envelope
}
func (s *StreamDataMessageWrapper) SetData(data []byte) {
	s.Envelope = data
}

func (s *StreamDataMessageWrapper) GetChanelID() types.ChannelID {
	return types.ChannelID(s.ChannelID)
}
func (s *StreamDataMessageWrapper) SetChannelID(cid types.ChannelID) {
	s.ChannelID = uint32(cid)
}
func (s *StreamDataMessageWrapper) SetStreamFlagInfo(flag int32) {
	s.StreamFlag = StreamFlag(flag)
}
func (s *StreamDataMessageWrapper) GetStreamFlagInfo()int32 {
	return int32(s.GetStreamFlag())
}
func (s *StreamDataMessageWrapper) ToBytes() ([]byte, error) {
	return proto.Marshal(s)
}

type StreamCommonProtocolWrapper struct {
}

func (s *StreamCommonProtocolWrapper) Encode(data []byte, decorators ...services.Decorator) services.IStreamObject {
	r := &StreamDataMessageWrapper{}
	r.SetData(data)
	for _, f := range decorators {
		f(r)
	}
	return r
}

func (s *StreamCommonProtocolWrapper) Decode(bytes []byte) (services.IStreamObject, error) {
	r := new(StreamDataMessageWrapper)
	if err := proto.Unmarshal(bytes, r); nil != err {
		return nil, err
	}
	return r, nil
}

func (this *StreamDataMessageWrapper) GetInternalData() interface{} {
	return nil
}
