/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 7:23 下午
# @File : helper.go
# @Description :
# @Attention :
*/
package block

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/orderedcode"
)


//---------------------------------- KEY ENCODING -----------------------------------------

// key prefixes
const (
	// prefixes are unique across all tm db's
	prefixBlockMeta   = int64(0)
	prefixBlockPart   = int64(1)
	prefixBlockCommit = int64(2)
	prefixSeenCommit  = int64(3)
	prefixBlockHash   = int64(4)
)


func decodeBlockMetaKey(key []byte) (height int64, err error) {
	var prefix int64
	remaining, err := orderedcode.Parse(string(key), &prefix, &height)
	if err != nil {
		return
	}
	if len(remaining) != 0 {
		return -1, fmt.Errorf("expected complete key but got remainder: %s", remaining)
	}
	if prefix != prefixBlockMeta {
		return -1, fmt.Errorf("incorrect prefix. Expected %v, got %v", prefixBlockMeta, prefix)
	}
	return
}

func blockPartKey(height uint64, partIndex int) []byte {
	key, err := orderedcode.Append(nil, prefixBlockPart, height, int64(partIndex))
	if err != nil {
		panic(err)
	}
	return key
}

func blockCommitKey(height uint64) []byte {
	key, err := orderedcode.Append(nil, prefixBlockCommit, height)
	if err != nil {
		panic(err)
	}
	return key
}
func blockMetaKey(height uint64) []byte {
	key, err := orderedcode.Append(nil, prefixBlockMeta, height)
	if err != nil {
		panic(err)
	}
	return key
}
func seenCommitKey(height uint64) []byte {
	key, err := orderedcode.Append(nil, prefixSeenCommit, height)
	if err != nil {
		panic(err)
	}
	return key
}

func blockHashKey(hash []byte) []byte {
	key, err := orderedcode.Append(nil, prefixBlockHash, string(hash))
	if err != nil {
		panic(err)
	}
	return key
}

//-----------------------------------------------------------------------------

// mustEncode proto encodes a proto.message and panics if fails
func mustEncode(pb proto.Message) []byte {
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(fmt.Errorf("unable to marshal: %w", err))
	}
	return bz
}
