/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 12:46 下午
# @File : helper.go
# @Description :
# @Attention :
*/
package types

import (
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

// IsVoteTypeValid returns true if t is a valid vote type.
func IsVoteTypeValid(t types.SignedMsgType) bool {
	switch t {
	case types.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE, types.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
		return true
	default:
		return false
	}
}


