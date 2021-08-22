/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 2:35 下午
# @File : canonical.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

func CanonicalizeBlockID(bid *types.BlockIDProto) *types.CanonicalBlockIDProto {
	rbid, err := BlockIDFromProto(bid)
	if err != nil {
		panic(err)
	}
	var cbid *types.CanonicalBlockIDProto
	if rbid == nil || rbid.IsZero() {
		cbid = nil
	} else {
		header := CanonicalizePartSetHeader(*bid.PartSetHeader)
		cbid = &types.CanonicalBlockIDProto{
			Hash:          bid.Hash,
			PartSetHeader: &header,
		}
	}

	return cbid
}
func BlockIDFromProto(bID *types.BlockIDProto) (*BlockID, error) {
	if bID == nil {
		return nil, errors.New("nil BlockID")
	}

	blockID := new(BlockID)
	ph, err := PartSetHeaderFromProto(bID.PartSetHeader)
	if err != nil {
		return nil, err
	}

	blockID.PartSetHeader = *ph
	blockID.Hash = bID.Hash

	return blockID, blockID.ValidateBasic()
}
func PartSetHeaderFromProto(ppsh *types.PartSetHeaderProto) (*PartSetHeader, error) {
	if ppsh == nil {
		return nil, errors.New("nil PartSetHeader")
	}
	psh := new(PartSetHeader)
	psh.Total = ppsh.Total
	psh.Hash = ppsh.Hash

	return psh, psh.ValidateBasic()
}
// CanonicalizeVote transforms the given PartSetHeader to a CanonicalPartSetHeader.
func CanonicalizePartSetHeader(psh types.PartSetHeaderProto) types.CanonicalPartSetHeaderProto {
	return types.CanonicalPartSetHeaderProto(psh)
}
