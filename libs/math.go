/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/2 3:51 下午
# @File : math.go
# @Description :
# @Attention :
*/
package libs

import (
	"errors"
	"math"
)

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//-----------------------------------------------------------------------------

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
var ErrOverflowInt32 = errors.New("int32 overflow")
var ErrOverflowUint8 = errors.New("uint8 overflow")
var ErrOverflowInt8 = errors.New("int8 overflow")
// SafeConvertUint8 takes an int64 and checks if it overflows
// If there is an overflow it returns an error
func SafeConvertUint8(a int64) (uint8, error) {
	if a > math.MaxUint8 {
		return 0, ErrOverflowUint8
	} else if a < 0 {
		return 0, ErrOverflowUint8
	}
	return uint8(a), nil
}

// SafeSubInt32 subtracts two int32 integers
// If there is an overflow this will panic
func SafeSubInt32(a, b int32) int32 {
	if b > 0 && (a < math.MinInt32+b) {
		panic(ErrOverflowInt32)
	} else if b < 0 && (a > math.MaxInt32+b) {
		panic(ErrOverflowInt32)
	}
	return a - b
}
func SafeAddInt32(a, b int32) int32 {
	if b > 0 && (a > math.MaxInt32-b) {
		panic(ErrOverflowInt32)
	} else if b < 0 && (a < math.MinInt32-b) {
		panic(ErrOverflowInt32)
	}
	return a + b
}
func SafeConvertInt32(a int64) int32 {
	if a > math.MaxInt32 {
		panic(ErrOverflowInt32)
	} else if a < math.MinInt32 {
		panic(ErrOverflowInt32)
	}
	return int32(a)
}