/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/31 12:12 下午
# @File : date.go
# @Description :
# @Attention :
*/
package commonutils

import (
	"fmt"
	"time"
)

const (
	JAVA_TIME = 1000000000000
)

var loc *time.Location

func init() {
	loc = time.FixedZone("CST", 8*3600)
}

func DateOver(v int64) *timeBy {
	if v >= JAVA_TIME {
		v = v / 1000
	}
	// CST
	r := timeBy{}
	r.sendTime = time.Unix(v, 0).In(loc)
	r.Now = time.Now().In(loc)
	r.sendTimestamp = v
	return &r
}

type timeBy struct {
	sendTime      time.Time
	Now           time.Time
	sendTimestamp int64
	sub           time.Duration
}

func (this timeBy) String() string {
	this.ready()
	sendTStr := this.sendTime.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("发送时间:%s,发送时间戳:%d,当前时间:%s,时间间隔(s):%f,时间间隔(毫秒):%d", sendTStr, this.sendTimestamp, this.Now.String(), this.sub.Seconds(), this.sub.Milliseconds())
}
func (this *timeBy) ready() {
	if this.sub == 0 {
		if this.Now.Before(this.sendTime) {
			this.sub = this.sendTime.Sub(this.Now)
		} else {
			this.sub = this.Now.Sub(this.sendTime)
		}
	}
}
func (this timeBy) SecondsTimeBy() float64 {
	this.ready()
	return this.sub.Seconds()
}
func (this timeBy) MillisecondsTimeBy() int64 {
	this.ready()
	return this.sub.Milliseconds()
}
