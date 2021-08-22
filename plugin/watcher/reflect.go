/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/10 8:49 上午
# @File : reflect.go
# @Description :
# @Attention :
*/
package watcher

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"reflect"
	"time"
)

type reflectC struct {
	ch       <-chan channel.IData
	consumer DataConsumer
	name     string
}

func newReflectC(name string, cc <-chan channel.IData, f DataConsumer) *reflectC {
	r := &reflectC{
		ch:       cc,
		consumer: f,
		name:     name,
	}
	return r
}

type reflectChannelWatcher struct {
	*baseChannelWatcher
	chs []*reflectC
}

func newReflectChannelWatcher(opt Opt) *reflectChannelWatcher {
	r := &reflectChannelWatcher{
	}
	r.baseChannelWatcher = newBaseChannelWatcher("reflect", r, opt.ReflectUpgradeLimit, opt.ReflectRollbackLimit, int32(opt.SpinTimeMills), opt.RoutinePoolFactory, opt.Mode)

	return r
}

func (this *reflectChannelWatcher) PrintSelf() {

}
func fromRoutineChannelWatcher(rout *routingChannelWatcher, opt Opt) *reflectChannelWatcher {
	r := newReflectChannelWatcher(opt)
	r.baseStatus = rout.baseStatus
	r.baseStatus.status = status_ok

	chs := make([]*reflectC, 0)
	rout.mtx.Lock()
	iterat := rout.young.Iterator()
	for iterat.Next() {
		ch := iterat.Value().(*c)
		if ch.closed() {
			continue
		}
		if !ch.stopped() {
			PanicWithMsg(nil, fmt.Sprintf("cant do it,%d,%s,time:%s", ch.status, ch.name, ch.startt.String()))
		}
		chs = append(chs, &reflectC{
			ch:       ch.c,
			consumer: ch.consumer,
			name:     ch.name,
		})
	}
	rout.mtx.Unlock()
	r.chs = chs
	return r
}

func (this *reflectChannelWatcher) OnUpgrade(opt Opt) (ChannelWatcher, []models.StartOption) {
	r, ctx := fromReflectChannelWatcher(this, opt)
	return r, []models.StartOption{models.SyncStartOpt, models.CtxStartOpt(ctx)}
}
func (this *reflectChannelWatcher) GetChannelShims(cap int) (map[channel.ChannelID]*ChannelWp, int) {
	r := make(map[channel.ChannelID]*ChannelWp)
	r[memberNotifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: memberNotifyC,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			member := v.(ChannelMember)
			this.dec(v)
			this.addDelta(member)
		},
	}
	return r, 1
}
func (this *reflectChannelWatcher) OnRollBack(opt Opt) (ChannelWatcher, []models.StartOption) {
	r := newRoutineChannelWatcher(opt)
	r.baseStatus = this.baseStatus
	r.baseStatus.status = status_ok
	var chM []ChannelMember
	for _, ch := range this.chs {
		chM = append(chM, ChannelMember{
			name:     ch.name,
			c:        ch.ch,
			consumer: ch.consumer,
		})
	}
	return r, []models.StartOption{models.CtxStartOpt(r.wrapNewCtxWithMember(chM))}
}
func (this *reflectChannelWatcher) OnStart(ctx *models.StartCTX) error {
	this.Logger.Info("chs的初始长度", "size", len(this.chs))
	go this.reflect()

	return nil
}
func (this *reflectChannelWatcher) reflect() {
	cases := make([]reflect.SelectCase, len(this.chs))
	for i, ch := range this.chs {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.ch)}
	}
	memChan := this.internalChs[memberNotifyC].ch.Ch
	add := func(data channel.IData) {
		this.dec(data)
		msg := data.(ChannelMember)
		this.chs = append(this.chs, newReflectC(msg.name, msg.c, msg.consumer))
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(msg.c)})
	}
	defer this.wg.Done()
	var index int
	var value reflect.Value
	var ok bool
	var msg ChannelMember
	for {
		select {
		case <-this.internalFastQuitC:
			return
		case m := <-memChan:
			msg = m.(ChannelMember)
			add(msg)
		default:
			if len(this.chs) == 0 {
				begin := time.Now()
				this.Logger.Info("sleeping", "currentSize", len(this.chs), "deltaSize", this.deleta.Size(), "memChanLen", len(memChan))
				select {
				case m := <-memChan:
					msg = m.(ChannelMember)
					add(msg)
				case <-this.internalFastQuitC:
					return
				}
				cost := time.Now().Sub(begin)
				this.Logger.Info("退出sleep", "耗时", cost.Seconds())
			}
		}

		index, value, ok = reflect.Select(cases)
		if !ok {
			this.Logger.Debug("routine退出", "name", this.chs[index].name, "size", len(this.chs)-1)
			cases = append(cases[:index], cases[index+1:]...)
			this.chs = append(this.chs[:index], this.chs[index+1:]...)
			continue
		}
		msg := value.Interface().(channel.IData)
		v := this.wrapV(msg)
		cc := this.chs[index].consumer
		this.execute(cc.Async(), func() {
			cc.Handle(v)
		})
	}
}
func (this *reflectChannelWatcher) OnStop(ctx *models.StopCTX) {
	this.chs = nil
}
func (this *reflectChannelWatcher) wrapV(v channel.IData) channel.IData {
	return v
}
func (this *reflectChannelWatcher) HandleMsg(v channel.IData) {
	switch v.(type) {

	}
}

func (this *reflectChannelWatcher) Size() int {
	return len(this.chs)
}
