/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/11 4:16 下午
# @File : go
# @Description : FIXME NOT HERE
# @Attention :
*/
package commonutils

import (
	"strconv"
	"strings"
)

const (
	BASE64 = iota + 1
	RAW
)

const (
	BIDSUN = "bidsun"
)

const (
	BIDSUN_DATA_PREFIX_LEN = 7

	BIDSUN_BEGIN_INDEX    = 0
	BIDSUN_END_INDEX      = 6
	BIDSUN_KEY_FLAG_INDEX = 6

	BIDSUN_KEY_BEGIN_INDEX = 7
)

const (
	BUILDER_DATA_LENGTH_INDEX = 0
)

// 格式: bidsunKey的长度+bidsunKey+cid
type BidsunData []byte

// 格式: bidsun+flag+key
type BidsunKey string


func BuildBidsunKey(originK string) BidsunKey {
	if BidsunKey(originK).Valid() {
		return BidsunKey(originK)
	}
	r := BIDSUN + strconv.Itoa(RAW) + originK
	return BidsunKey(r)
}

func NewBidsun(bidsunKey BidsunKey, outData []byte) BidsunData {
	l := len(bidsunKey)
	cidBs := outData
	cap := 1 + l + len(cidBs)
	data := make([]byte, cap, cap)
	data[0] = byte(l)
	bidsunBegin := 1
	bidsunEnd := 1 + l
	copy(data[bidsunBegin:bidsunEnd], bidsunKey)
	copy(data[bidsunEnd:], cidBs)

	return data
}


func BidsunDataToBidsunKey(data []byte) (BidsunKey, []byte) {
	l := data[BUILDER_DATA_LENGTH_INDEX]
	bidsunKeyBegin := BUILDER_DATA_LENGTH_INDEX + 1
	bidsunKeyEnd := BUILDER_DATA_LENGTH_INDEX + 1 + l
	bidsunKey := data[bidsunKeyBegin:bidsunKeyEnd]
	lastRootCid := data[bidsunKeyEnd:]
	return BidsunKey(bidsunKey), lastRootCid
}
func (this BidsunData) GrapData() string {
	data := this
	ll := data[BUILDER_DATA_LENGTH_INDEX]
	bidsunKeyEnd := BUILDER_DATA_LENGTH_INDEX + 1 + ll
	return string(this[bidsunKeyEnd:])
}
func (this BidsunData) String() string {
	return string(this)
}
func (this BidsunData) GrapLenth() byte {
	return this[0]
}
func (this BidsunData) Valid() bool {
	data := this
	l := len(data)
	if l < BIDSUN_DATA_PREFIX_LEN {
		return false
	}
	ll := data[BUILDER_DATA_LENGTH_INDEX]
	bidsunKeyBegin := BUILDER_DATA_LENGTH_INDEX + 1
	bidsunKeyEnd := BUILDER_DATA_LENGTH_INDEX + 1 + ll
	bidsunKey := data[bidsunKeyBegin:bidsunKeyEnd]
	if strings.HasPrefix(string(bidsunKey), BIDSUN) {
		return true
	}
	return false
}

// make sure valid
func (this BidsunData) GrapBidsunKey() BidsunKey {
	data := this
	ll := data[BUILDER_DATA_LENGTH_INDEX]
	bidsunKeyBegin := BUILDER_DATA_LENGTH_INDEX + 1
	bidsunKeyEnd := BUILDER_DATA_LENGTH_INDEX + 1 + ll
	bidsunKey := data[bidsunKeyBegin:bidsunKeyEnd]
	return BidsunKey(bidsunKey)
}

// make sure valid
func (this BidsunKey) GrapOriginKey() (string, bool) {
	bidsun := this[BIDSUN_BEGIN_INDEX:BIDSUN_END_INDEX]
	if bidsun != BIDSUN {
		return "", false
	}
	r := this[BIDSUN_KEY_BEGIN_INDEX:]
	return string(r), true
}
func (this BidsunKey) Valid() bool {
	if len(this) < len(BIDSUN) {
		return false
	}
	bidsun := this[BIDSUN_BEGIN_INDEX:BIDSUN_END_INDEX]
	if bidsun != BIDSUN {
		return false
	}
	return true
}
