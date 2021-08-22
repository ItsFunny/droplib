/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/20 8:50 上午
# @File : listener_test.go.go
# @Description :
# @Attention :
*/
package listener

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"strconv"
	"testing"
	"time"
)

func Test_multi(t *testing.T) {
	l := DefaultNewListenerComponent()
	l.BStart(models.AsyncStartOpt)
	for i := 0; i < 100; i++ {
		go func(index int) {
			listener := l.RegisterListener(strconv.Itoa(index))
			fmt.Println(<-listener)
		}(i)
	}
	go func() {
		time.Sleep(time.Second * 4)
		l.NotifyListener(123, strconv.Itoa(1))
	}()
	time.Sleep(time.Second * 2000)

}
