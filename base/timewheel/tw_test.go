/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 8:30 上午
# @File : tw_test.go.go
# @Description :
# @Attention :
*/
package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func Test_TW(t *testing.T) {
	tt := New(time.Second, 10, func(i interface{}) {
		fmt.Println("开始执行任务", i)
	})
	tt.Start()

	tt.AddTimer(time.Second, 1, 1)
	tt.AddTimer(time.Second, 2, 2)
	tt.AddTimer(time.Second, 3, 3)
	tt.AddTimer(time.Second, 4, 4)
	tt.AddTimer(time.Second, 5, 5)

	time.Sleep(time.Second * 30)
}
