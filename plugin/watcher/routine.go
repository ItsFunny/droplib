/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/10 8:49 上午
# @File : routine.go
# @Description :
# @Attention :
*/
package watcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"github.com/hyperledger/fabric-droplib/structure/maps/linkedhashmap"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var emptyF func()

type reuseRoutine struct {
	m      ChannelMember
	reuseC *c
}

func (r reuseRoutine) ID() interface{} {
	return fmt.Sprintf("%v_reuse", r.m.ID())
}

type routingChannelWatcher struct {
	*baseChannelWatcher

	young *linkedhashmap.Map
	size  int32
}

func newRoutineChannelWatcher(opt Opt) *routingChannelWatcher {
	r := &routingChannelWatcher{}
	r.baseChannelWatcher = newBaseChannelWatcher("router", r, opt.RoutineUpgradeLimit, 0, int32(opt.SpinTimeMills), opt.RoutinePoolFactory, opt.Mode)
	r.young = linkedhashmap.New()
	return r
}

func (this *routingChannelWatcher) OnStart(ctx *models.StartCTX) error {
	go this.daemon()
	go this.gc()
	go this.startWithCtx(ctx)
	go this.opt()

	return nil
}

func (this *routingChannelWatcher) GetChannelShims(cap int) (map[channel.ChannelID]*ChannelWp, int) {
	r := make(map[channel.ChannelID]*ChannelWp)
	r[route_update_notify_chid] = &ChannelWp{
		ch: &channel.Channel{
			Id: route_update_notify_chid,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			this.handleMsg(flushWrapper{v: v})
		},
	}
	r[upgradeRollbackNotifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: upgradeRollbackNotifyC,
			Ch: make(chan channel.IData, 10),
		},
		flush: func(v channel.IData) {
			this.handleMsg(v)
		},
	}
	r[reuse_notifyc] = &ChannelWp{
		ch: &channel.Channel{
			Id: reuse_notifyc,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			this.handleMsg(flushWrapper{v: v})
		},
	}
	r[memberNotifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: memberNotifyC,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			this.handleMsg(flushWrapper{v: v})
		},
	}

	return r, 2
}
func (this *routingChannelWatcher) startWithCtx(ctx *models.StartCTX) {
	startChs := this.getStartChannelFromCtx(ctx)
	if nil == startChs {
		return
	}
	chs := *startChs
	for _, ch := range chs {
		this.OnMemberChanged(ch)
	}
}
func (this *routingChannelWatcher) OnStop(ctx *models.StopCTX) {
}
func (this *routingChannelWatcher) OnUpgrade(opt Opt) (ChannelWatcher, []models.StartOption) {
	listener := this.listener.RegisterListener(listener_upgrade)
	this.sendMsg(upgradeRollbackNotifyC, newUpgrade(opt))
	res := <-listener
	v := res.(upgradeResp)
	return v.r, v.opts
}

func (this *routingChannelWatcher) OnRollBack(opt Opt) (ChannelWatcher, []models.StartOption) {
	panic("not supported")
}
func (this *routingChannelWatcher) gc() {
	defer this.wg.Done()
}

func (this *routingChannelWatcher) daemon() {
	updateNotifyC := this.internalChs[route_update_notify_chid].ch.Ch
	memberNotifyC := this.internalChs[memberNotifyC].ch.Ch
	upRooC := this.internalChs[upgradeRollbackNotifyC].ch.Ch
	reuse_notifyc := this.internalChs[reuse_notifyc].ch.Ch
	defer this.wg.Done()
	for {
		select {
		case <-this.internalFastQuitC:
			return
		case v, ok := <-upRooC:
			if !ok {
				return
			}
			this.handleMsg(v)
		case v, ok := <-updateNotifyC:
			if !ok {
				return
			}
			this.handleMsg(v)
		case v, ok := <-memberNotifyC:
			if !ok {
				return
			}
			this.handleMsg(v)
		case v, ok := <-reuse_notifyc:
			if !ok {
				return
			}
			this.handleMsg(v)
		}
	}
}

func (this *routingChannelWatcher) handleMsg(v interface{}) {
	switch msg := v.(type) {
	case upgrade:
		this.sendOperation(op_upgrade, msg)
	case ChannelMember:
		this.sendOperation(op_new_member, routineAddMemberWrapper{m: msg})
	case routineCClose:
		this.sendOperation(op_routine_gone, msg)
	case reuseRoutine:
		this.sendOperation(op_new_member, routineAddMemberWrapper{
			m:      msg.m,
			reuseC: msg.reuseC,
		})
	case flushWrapper:
		switch data := msg.v.(type) {
		case routineCClose:
			this.sendOperation(op_routine_gone, msg.v)
		case ChannelMember:
			this.sendOperation(op_new_member, routineAddMemberWrapper{m: data})
		case reuseRoutine:
			this.sendOperation(op_new_member, routineAddMemberWrapper{
				m:      data.m,
				reuseC: data.reuseC,
			})
		}
	default:
		panic(fmt.Sprintf("未知的数据类型,data=%v,type=%v", v, reflect.TypeOf(v)))
	}
}
func (this *routingChannelWatcher) handleRoutineGone(reuse bool, v interface{}) {
	msg := v.(routineCClose)

	this.mtx.Lock()
	get, found := this.young.Get(msg.key)
	this.mtx.Unlock()
	if !found {
		PanicWithMsg(PROGRAMA_ERROR, fmt.Sprintf("name=%s", msg.key))
		return
	}
	cc := get.(*c)
	if !cc.closed() {
		if !cc.stopped() {
			PanicWithMsg(PROGRAMA_ERROR, "cant happen2:"+cc.name+",status:"+strconv.Itoa(int(cc.status))+",id:"+strconv.Itoa(cc.id))
		}
		this.listener.NotifyListener(nil, cc.name)
		return
	}
	cc.c = nil
	if reuse {
		memberNotifyC := this.internalChs[memberNotifyC].ch.Ch
		var m ChannelMember
		select {
		case msg, ok := <-memberNotifyC:
			if ok {
				// we steal a one
				m = msg.(ChannelMember)
				this.dec(m)
			}
		default:
		}
		if m.c != nil {
			cc.panicCAS(routinec_status_close, routinec_status_before_reuse)
			err := this.reuseC(m, cc)
			if nil == err {
				return
			}
			PanicWithMsg(err, "添加member失败")
		}
	}
	name := cc.name
	releaseC(cc)
	this.routinePool.AddJob(false, services.Job{
		Handler: func() error {
			this.listener.NotifyListener(nil, name)
			return nil
		},
	})

	this.mtx.Lock()
	this.young.Remove(msg.key)
	this.mtx.Unlock()
}
func (this *routingChannelWatcher) reuseC(m ChannelMember, existC *c) error {
	existC.fromMember(m, this.listener)
	this.sendMsg(reuse_notifyc, reuseRoutine{
		m:      m,
		reuseC: existC,
	})
	return nil
}

func (this *routingChannelWatcher) opt() {
	trick := false
	handleNewMember := func(v channel.IData) {
		msg := v.(routineAddMemberWrapper)
		if trick {
			this.addDelta(msg.m)
			atomic.AddInt32(&this.size, 1)
			if msg.reuseC != nil {
				this.listener.NotifyListener(msg.m.name)
			}
			return
		}
		member, err := this.addMember(msg)
		if nil != err {
			this.Logger.Error("添加member失败:" + err.Error())
			return
		}
		this.routineListenNew(member)
	}
	handleRoutineGone := func(v interface{}) {
		atomic.AddInt32(&this.size, -1)
		switch v.(type) {
		case flushWrapper:
			this.handleRoutineGone(false, v.(flushWrapper).v)
		default:
			if trick {
				this.handleRoutineGone(false, v)
			} else {
				this.handleRoutineGone(true, v)
			}
		}
	}
	handleUpgrade := func(v channel.IData) {
		if trick {
			panic(PROGRAMA_ERROR)
		}
		trick = true
		go func() {
			opt := v.(upgrade).opt
			this.Logger.Warn("开始stw", "now", time.Now())
			begin := time.Now()
			this.stw()
			cost := time.Now().Sub(begin)
			r := fromRoutineChannelWatcher(this, opt)
			this.listener.NotifyListener(upgradeResp{opts: nil, r: r}, listener_upgrade)
			this.Logger.Warn("结束stw", "耗时", cost.Seconds())
		}()
	}
	for {
		select {
		case msg, ok := <-this.operation:
			if !ok {
				this.Logger.Debug("operation 退出")
				return
			}
			this.dec(msg.data)
			switch msg.opType {
			case op_new_member:
				handleNewMember(msg.data)
			case op_routine_gone:
				handleRoutineGone(msg.data)
			case op_upgrade:
				handleUpgrade(msg.data)
			default:
				panic("未知")
			}
		}
	}
}

type routineAddMemberWrapper struct {
	m      ChannelMember
	reuseC *c
}

func (r routineAddMemberWrapper) ID() interface{} {
	return r.m.ID()
}

func (this *routingChannelWatcher) addMember(wp routineAddMemberWrapper) (*c, error) {
	m := wp.m
	_, exist := this.young.Get(m.name)
	if exist {
		return nil, errors.New("重复的routine," + "name:" + m.name)
	}
	cc := wp.reuseC
	if cc == nil {
		cc = newC(m.name, m.c, m.consumer, func() <-chan interface{} {
			return this.listener.RegisterListener(m.name)
		})
	}
	this.young.Put(m.name, cc)
	atomic.AddInt32(&this.size, 1)
	return cc, nil
}

// FIXME stack over flow
func (this *routingChannelWatcher) routineListenNew(c *c) {
	if !atomic.CompareAndSwapUint32(&c.status, routinec_status_ok, routinec_status_running) {
		panic(PROGRAMA_ERROR)
		return
	}
	this.Logger.Debug("启动routine", "name", c.name)
	go func() {
		for {
			select {
			case _, ok := <-c.notifyC:
				if !ok {
					if atomic.CompareAndSwapUint32(&c.status, routinec_status_wait_listener, routinec_status_upgrade) {
						this.sendMsg(route_update_notify_chid, routineCClose{key: c.name})
						return
					} else {
						panic("4444 cant")
					}
				}
			case value, ok := <-c.c:
				if !ok {
					if atomic.CompareAndSwapUint32(&c.status, routinec_status_running, routinec_status_close) {
						this.sendMsg(route_update_notify_chid, routineCClose{key: c.name})
						return
					} else if atomic.LoadUint32(&c.status) == routinec_status_wait_listener {
						atomic.StoreUint32(&c.status, routinec_status_upgrade)
						this.sendMsg(route_update_notify_chid, routineCClose{key: c.name})
						return
					} else {
						panic(fmt.Sprintf("cant be ,status=%d,name=%s", c.status, c.name))
					}
				}
				c.consumer.Handle(value)
			}
		}
	}()
}
func (this *routingChannelWatcher) stw() {
	this.parallelSTW()
}
func (this *routingChannelWatcher) parallelSTW() {
	this.mtx.Lock()
	iterator := this.young.Iterator()
	rc := make(chan *c, this.young.Size())
	for iterator.Next() {
		rc <- iterator.Value().(*c)
	}
	close(rc)
	this.mtx.Unlock()

	wg := sync.WaitGroup{}
	ctx := context.Background()
	if this.mode == mode_debug {
		c, cancel := context.WithTimeout(ctx, time.Second*180)
		defer cancel()
		ctx = c
	}
	wgCount := runtime.NumCPU()>>1 + 1
	wgCount = 1
	wg.Add(wgCount)
	go func(rc chan *c) {
		for i := 0; i < wgCount; i++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case v, ok := <-rc:
						if !ok {
							return
						}
						l := v.close(this.listener)
						select {
						case <-l:
						case <-ctx.Done():
							panic(PROGRAMA_ERROR)
						}
					}
				}
			}()
		}
	}(rc)
	wg.Wait()
}
func (this *routingChannelWatcher) Size() int {
	return int(atomic.LoadInt32(&this.size))
}
