/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/18 9:06 上午
# @File : routine_test.go.go
# @Description :
# @Attention :
*/
package impl

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"github.com/panjf2000/ants/v2"
	"testing"
	"time"
)

func TestRoutine(t *testing.T) {
	c := NewRoutineComponent()
	ch := make(chan struct{})
	c.AddJob(false, services.Job{
		Pre: nil,
		Handler: func() error {
			fmt.Println("work1")
			ch <- struct{}{}
			return nil
		},
		Post: nil,
	})

	<-ch
	fmt.Println("end")
}

func Test_Ants(t *testing.T) {
	pool, err := ants.NewPool(1)
	if nil != err {
		panic(err)
	}
	go func() {
		for i := 0; i < 1000; i++ {
			go func() {
				ti := time.NewTicker(time.Millisecond)
				for {
					select {
					case <-ti.C:
						if err = pool.Submit(func() {
							fmt.Println(1)
						}); nil != err {
							panic(err)
						}
					}
				}
			}()
		}

	}()
	for {
		select {}
	}

}
