/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 7:32 下午
# @File : block_meta.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

func BlockMetaFromProto(pb *types.BlockMetaProto) (*BlockMeta, error) {
	if pb == nil {
		return nil, errors.New("blockmeta is empty")
	}

	bm := new(BlockMeta)

	bi, err := BlockIDFromProto(pb.BlockId)
	if err != nil {
		return nil, err
	}

	h, err := HeaderFromProto(pb.Header)
	if err != nil {
		return nil, err
	}

	bm.BlockID = *bi
	bm.BlockSize = int(pb.BlockSize)
	bm.Header = h
	bm.NumTxs = int(pb.NumTxs)

	return bm, bm.ValidateBasic()
}
