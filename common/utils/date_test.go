/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/13 10:47 上午
# @File : date_test.go.go
# @Description :
# @Attention :
*/
package commonutils

import (
	"log"
	"testing"
	"time"
)

func Test_DateOver(t *testing.T) {
	add := time.Now().Add(-time.Minute * 10).In(loc)
	over := DateOver(add.Unix())
	log.Println(over)
}

func Test_Faster(t *testing.T) {
	add := time.Now().Add(time.Minute * 5).In(loc)
	over := DateOver(add.Unix())
	log.Println(over)
}
