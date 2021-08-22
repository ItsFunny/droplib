/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 6:04 下午
# @File : task_queue_pb_test.go.go
# @Description :
# @Attention :
*/
package v2

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/libs/queue/concurrentpriority"
	"math/rand"
	"testing"
	"time"
)

func Test_Priority(t *testing.T) {
	queue := concurrentpriority.NewPriorityQueue(func(a, b interface{}) int {
		a1 := a.(IPriorityTask)
		a2 := a.(IPriorityTask)
		return a1.GetPriority() - a2.GetPriority()
	})
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 1))
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 2))
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 3))
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 4))
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 5))
	queue.Push(NewDefaultPriorityTask(func() {
		fmt.Println(1)
	}, 6))
	for queue.Len() > 0 {
		v := queue.Pop()
		fmt.Println(v.(IPriorityTask).GetPriority())
	}
}

func Test_One_Consumer_multi_Producer(t *testing.T) {
	queue := NewPriorityTaskQueue()
	consumer := func() {
		for {
			task := queue.Take()
			task.Execute()
		}
	}
	producer := func(number int) {
		tt := time.NewTicker(time.Millisecond * 2)
		for {
			select {
			case <-tt.C:
				if _, err := queue.Push(NewDefaultPriorityTask(func() {
					fmt.Println("number:", number)
				}, number)); nil != err {
					panic(err)
				}
			}
		}
	}

	go consumer()

	for i := 0; i < 100; i++ {
		go producer(i)
	}
	time.Sleep(time.Second * 2)
}

func Test_Single_PriorityPool(t *testing.T) {
	pool, _ := NewPool(WithSize(1), WithPriorityQueue())

	producer := func() {
		p := rand.Intn(100)
		if err := pool.SubmitTask(NewDefaultPriorityTask(func() {
			time.Sleep(time.Second * 3)
			fmt.Println("执行了", p)
		}, p)); nil != err {
			fmt.Println("错误", err.Error())
		}
	}

	for i := 0; i < 100; i++ {
		go producer()
	}
	for {
		select {}
	}
}
