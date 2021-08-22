/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:06 上午
# @File : common.go
# @Description :
# @Attention :
*/
package protobuflibs

import "github.com/golang/protobuf/proto"

// Marshal serializes a protobuf message.
func Marshal(pb proto.Message) ([]byte, error) {
	return proto.Marshal(pb)
}


