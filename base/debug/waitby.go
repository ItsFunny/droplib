/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/1 1:43 下午
# @File : waitby.go
# @Description :
# @Attention :
*/
package debug

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func WaitFor() {
	fmt.Println("exit")
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGKILL, syscall.SIGINT)
	<-sigCh
}

