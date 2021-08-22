/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/8 3:02 下午
# @File : fail.go
# @Description :
# @Attention :
*/
package libs



import (
	"fmt"
	"os"
	"strconv"
)

func envSet() int {
	callIndexToFailS := os.Getenv("FAIL_TEST_INDEX")

	if callIndexToFailS == "" {
		return -1
	}

	var err error
	callIndexToFail, err := strconv.Atoi(callIndexToFailS)
	if err != nil {
		return -1
	}

	return callIndexToFail
}

// Fail when FAIL_TEST_INDEX == callIndex
var callIndex int // indexes Fail calls

func Fail() {
	callIndexToFail := envSet()
	if callIndexToFail < 0 {
		return
	}

	if callIndex == callIndexToFail {
		fmt.Printf("*** fail-test %d ***\n", callIndex)
		os.Exit(1)
	}

	callIndex++
}

