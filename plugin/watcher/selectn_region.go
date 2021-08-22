/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 11:29 上午
# @File : selectn_region.go
# @Description :
# @Attention :
*/
package watcher

import (
	"context"
	"fmt"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/component/listener"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	notifyc_block = iota
	notifyc_reuse
)

var emptyObj channel.IData

type chWp struct {
	c        <-chan channel.IData
	consumer DataConsumer
	name     string
	newM     bool
}
type reportGcGoneInfo struct {
	id   int
	bits byte
	c    *selectNC
	reg  IRegion
}
type RegionSort []IRegion

func (r RegionSort) Len() int {
	return len(r)
}

func (r RegionSort) Less(i, j int) bool {
	vi, vj := r[i], r[j]
	return vi.Bits() < vj.Bits()
}

func (r RegionSort) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type IRegion interface {
	fmt.Stringer
	Id() int
	Bits() byte // byte*8
	Chs() []*selectNC
	Start(ctx *RegionStartCTX)
	onStartRegion(ctx context.Context)

	Size() int
	PreBlock(ctx context.Context) bool
	Shutdown()
	Status() uint32
	CasStatus(exce int, newV int) bool
	Reuse(ctx context.Context, c ChannelMember) (*selectNC, bool)
	RunningCount() byte

	reset()
	Dec(v int32) bool
	Inc(v int32)
}
type RegionStartCTX struct {
	PreStart  func()
	Ctx       context.Context
	PostStart func()

	ElseCas func()
}

type baseRegion struct {
	id             int // 4|8
	chs            []*selectNC
	reportF        func(info *reportGcGoneInfo)
	getListener    func() listener.IListenerComponent
	getRoutinePool func() services.IRoutineComponent

	notifyC chan interface{}

	running int32
	*baseStatus

	impl IRegion
}

func (this *baseRegion) Shutdown() {
}
func (this *baseRegion) execute(index int, data channel.IData) {
	if this.chs[index].consumer.Async() {
		cop := this.chs[index].consumer
		this.getRoutinePool().AddJob(false, services.Job{
			Handler: func() error {
				cop.Handle(data)
				return nil
			},
		})
	} else {
		this.chs[index].consumer.Handle(data)
	}
}
func (this *baseRegion) String() string {
	sb := strings.Builder{}
	for _, ch := range this.chs {
		sb.WriteString(fmt.Sprintf("name=%s,status=%d,", ch.name, ch.status))
	}
	sb.WriteString(fmt.Sprintf(",regionStatus=%d,regId=%d", this.status, this.id))
	sb.WriteString(fmt.Sprintf(",running=%d", this.running))
	sb.WriteString(fmt.Sprintf(",bits=%d", this.impl.Bits()))
	return sb.String()
}
func (this *baseRegion) Start(ctx *RegionStartCTX) {
	if this.CasStatus(region_status_new, region_status_running) {
		if nil != ctx.PreStart {
			ctx.PreStart()
		}
		this.onStartRegion(ctx.Ctx)
		if nil != ctx.PostStart {
			ctx.PostStart()
		}
	} else {
		if nil != ctx.ElseCas {
			ctx.ElseCas()
		}
	}
}
func (this *baseRegion) Resume(ctx *RegionStartCTX, deleta uint32) {
}
func (this *baseRegion) onStartRegion(ctx context.Context) {
	this.impl.onStartRegion(ctx)
}
func (this *baseRegion) Size() int {
	return len(this.chs)
}
func (this *baseRegion) RunningCount() byte {
	return byte(this.running)
}
func (this *baseRegion) PreBlock(ctx context.Context) bool {
	if this.CasStatus(region_status_running, region_status_pre_block) {
		v := this.getListener().RegisterListener(strconv.Itoa(this.id))
		this.notifyC <- notifyc_block
		<-v
		// FIXME ,restart
		return true
	}
	return false
}
func (reg *baseRegion) Reuse(ctx context.Context, c ChannelMember) (*selectNC, bool) {
	st := reg.Status()
	if st&region_status_with_blocking >= region_status_with_blocking {
		return nil, false
	}
	if st == region_status_to_remove {
		chs := reg.impl.Chs()
		for index, ch := range chs {
			if ch.status != selectn_channel_status_final {
				continue
			}
			// FIXME ,just atomic
			if !reg.CasStatus(region_status_to_remove, region_status_new) {
				break
			}
			if atomic.CompareAndSwapUint32(&ch.status, selectn_channel_status_final, selectn_channel_status_before_reuse) {
				go reg.Start(&RegionStartCTX{
					PreStart: func() {
						logplugin.MInfo(selectnModule, fmt.Sprintf("region重启,regId=%d,bits=%d,routineName=%s", reg.impl.Id(), reg.impl.Bits(), c.name))
					},
					Ctx: ctx,
					ElseCas: func() {
						logplugin.MInfo(selectnModule, "该reg已经处于running 状态,", reg.Id())
					},
				})

				reg.notifyC <- channelRegionMember{
					index: index,
					m:     c,
				}
				return ch, true
			}
		}
	} else if st == region_status_running {
		chs := reg.impl.Chs()
		for index, ch := range chs {
			if ch.status != selectn_channel_status_final {
				continue
			}
			if atomic.CompareAndSwapUint32(&ch.status, selectn_channel_status_final, selectn_channel_status_before_reuse) {
				atomic.StoreUint32(&ch.status, selectn_channel_status_reuse)
				reg.notifyC <- channelRegionMember{
					index: index,
					m:     c,
				}
				return ch, true
			} else {
				// stop continue
				return nil, false
			}
		}
	}
	return nil, false
}
func newRegion(arg *newBaseArgument) IRegion {
	l := len(*arg.chsW)
	switch l {
	case 32:
		return newSelect32Region(arg)
	case 16:
		return newSelect16Region(arg)
	case 8:
		return newSelect8Region(arg)
	case 4:
		return newSelect4Region(arg)
	case 2:
		return newSelect2Region(arg)
	case 1:
		return newSelect1Region(arg)
	default:
		panic("wrong:" + strconv.Itoa(l))
	}
}

type newBaseArgument struct {
	id          int
	chsW        *[]chWp
	reportCGone func(info *reportGcGoneInfo)
	listener    func() listener.IListenerComponent
	routine     func() services.IRoutineComponent
}

func newBaseRegion(arg *newBaseArgument, impl IRegion) *baseRegion {
	chs := *arg.chsW
	r := &baseRegion{
		id:  arg.id,
		chs: nil,
		baseStatus: &baseStatus{
			status:                        region_status_new,
			waitStatusIntervalMillSeconds: DEFAULT_WAIT_MILL_TIMES,
		},
		notifyC: make(chan interface{}),
	}
	r.reportF = arg.reportCGone
	r.chs = make([]*selectNC, len(chs))
	for i := 0; i < len(chs); i++ {
		r.chs[i] = newSelectNC(chs[i].name, chs[i].c, chs[i].consumer)
	}
	r.running = int32(impl.Bits())
	r.impl = impl
	r.getListener = arg.listener
	r.getRoutinePool = arg.routine
	return r
}
func releaseSelectNC(str string, c *selectNC) {
	c.consumer = nil
	c.c = nil
	c.releaseWhere = str
	atomic.StoreUint32(&c.status, selectn_channel_status_final)
}
func (this *baseRegion) reset() {
	this.running = int32(this.impl.Bits())
}

// FIXME ,routine 没有减一的问题
func (this *baseRegion) reportCGone(info *reportGcGoneInfo) {
	c := info.c
	st := atomic.LoadUint32(&c.status)
	if st&selectn_channel_status_reusing >= selectn_channel_status_reusing {
		selectNWaitUntil(c.name, this.waitStatusIntervalMillSeconds, c, selectn_channel_status_ok)
	}

	if atomic.CompareAndSwapUint32(&c.status, selectn_channel_status_ok, selectn_channel_status_gone) {
		v := atomic.AddInt32(&this.running, -1)
		releaseSelectNC("reportCGone,status_ok", c)
		if v == 0 {
			if this.Status() == region_status_to_remove {
				return
			}
			if atomic.CompareAndSwapUint32(&this.status, region_status_running, region_status_to_remove) {
				info.reg = this.impl
			}
		}
	} else {
		if !atomic.CompareAndSwapUint32(&c.status, selectn_channel_status_reuse, selectn_channel_status_gone) {
			return
		}
		v := atomic.AddInt32(&this.running, -1)
		releaseSelectNC("reportCGOne,stasas", c)
		if v == 0 {
			if this.Status()&region_status_removing >= region_status_removing {
				return
			}
			if atomic.CompareAndSwapUint32(&this.status, region_status_running, region_status_to_remove) {
				// report
				info.reg = this.impl
			}
		}
	}
	this.reportF(info)
}

func newSelect1Region(arg *newBaseArgument) *select1Region {
	r := &select1Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}
func newSelect2Region(arg *newBaseArgument) *select2Region {
	r := &select2Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}
func newSelect4Region(arg *newBaseArgument) *select4Region {
	r := &select4Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}
func newSelect8Region(arg *newBaseArgument) *select8Region {
	r := &select8Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}
func newSelect16Region(arg *newBaseArgument) *select16Region {
	r := &select16Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}
func newSelect32Region(arg *newBaseArgument) *select32Region {
	r := &select32Region{}
	r.baseRegion = newBaseRegion(arg, r)
	return r
}

type channelRegionMember struct {
	index int
	m     ChannelMember
}

func (this *baseRegion) Id() int {
	return this.id
}
func (this *baseRegion) Dec(v int32) bool {
	return atomic.AddInt32(&this.running, v*-1) == 0
}
func (this *baseRegion) Inc(v int32) {
	atomic.AddInt32(&this.running, v*1)
}
func (this *baseRegion) Chs() []*selectNC {
	return this.chs
}
func (this *baseRegion) wrapV(v channel.IData) channel.IData {
	return v
}

type select1Region struct {
	*baseRegion
}

func (s select1Region) Bits() byte {
	return 1 << 0
}
func (this *select1Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}
			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case <-c.Done():
				return
			}
		}
	}
}

type select2Region struct {
	*baseRegion
}

func (s select2Region) Bits() byte {
	return 1 << 1
}
func (this *select2Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}

			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case v, ok = <-this.chs[1].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[1], bits: this.Bits()})
					continue
				}
				this.execute(1, v)
			case <-c.Done():
				logplugin.Debug("ctxDone", "info", this.String())
				return
			}
		}
	}
}

type select4Region struct {
	//  1,2,4,8,16,32
	*baseRegion
}

func (s *select4Region) Bits() byte {
	return 1 << 2
}
func (this *select4Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}
			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case v, ok = <-this.chs[1].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[1], bits: this.Bits()})
					continue
				}
				this.execute(1, v)
			case v, ok = <-this.chs[2].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[2], bits: this.Bits()})
					continue
				}
				this.execute(2, v)
			case v, ok = <-this.chs[3].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[3], bits: this.Bits()})
					continue
				}
				this.execute(3, v)
			case <-c.Done():
				logplugin.Debug("ctxDone", "info", this.String())
				return
			}
		}
	}
}

type select8Region struct {
	//  1,2,4,8,16,32
	*baseRegion
}

func (s select8Region) Bits() byte {
	return 1 << 3
}
func (this *select8Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}
			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case v, ok = <-this.chs[1].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[1], bits: this.Bits()})
					continue
				}
				this.execute(1, v)
			case v, ok = <-this.chs[2].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[2], bits: this.Bits()})
					continue
				}
				this.execute(2, v)
			case v, ok = <-this.chs[3].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[3], bits: this.Bits()})
					continue
				}
				this.execute(3, v)
			case v, ok = <-this.chs[4].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[4], bits: this.Bits()})
					continue
				}
				this.execute(4, v)
			case v, ok = <-this.chs[5].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[5], bits: this.Bits()})
					continue
				}
				this.execute(5, v)
			case v, ok = <-this.chs[6].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[6], bits: this.Bits()})
					continue
				}
				this.execute(6, v)
			case v, ok = <-this.chs[7].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[7], bits: this.Bits()})
					continue
				}
				this.execute(7, v)
			case <-c.Done():
				logplugin.Debug("ctxDone", "info", this.String())
				return
			}
		}
	}
}

type select16Region struct {
	*baseRegion
}

func (s *select16Region) Bits() byte {
	return 1 << 4
}
func (this *select16Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}
			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case v, ok = <-this.chs[1].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[1], bits: this.Bits()})
					continue
				}
				this.execute(1, v)
			case v, ok = <-this.chs[2].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[2], bits: this.Bits()})
					continue
				}
				this.execute(2, v)
			case v, ok = <-this.chs[3].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[3], bits: this.Bits()})
					continue
				}
				this.execute(3, v)
			case v, ok = <-this.chs[4].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[4], bits: this.Bits()})
					continue
				}
				this.execute(4, v)
			case v, ok = <-this.chs[5].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[5], bits: this.Bits()})
					continue
				}
				this.execute(5, v)
			case v, ok = <-this.chs[6].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[6], bits: this.Bits()})
					continue
				}
				this.execute(6, v)
			case v, ok = <-this.chs[7].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[7], bits: this.Bits()})
					continue
				}
				this.execute(7, v)
			case v, ok = <-this.chs[8].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[8], bits: this.Bits()})
					continue
				}
				this.execute(8, v)
			case v, ok = <-this.chs[9].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[9], bits: this.Bits()})
					continue
				}
				this.execute(9, v)
			case v, ok = <-this.chs[10].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[10], bits: this.Bits()})
					continue
				}
				this.execute(10, v)
			case v, ok = <-this.chs[11].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[11], bits: this.Bits()})
					continue
				}
				this.execute(11, v)
			case v, ok = <-this.chs[12].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[12], bits: this.Bits()})
					continue
				}
				this.execute(12, v)
			case v, ok = <-this.chs[13].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[13], bits: this.Bits()})
					continue
				}
				this.execute(13, v)
			case v, ok = <-this.chs[14].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[14], bits: this.Bits()})
					continue
				}
				this.execute(14, v)
			case v, ok = <-this.chs[15].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[15], bits: this.Bits()})
					continue
				}
				this.execute(15, v)
			case <-c.Done():
				logplugin.Debug("ctxDone", "info", this.String())
				return
			}
		}
	}
}

type select32Region struct {
	*baseRegion
}

func (this *select32Region) onStartRegion(c context.Context) {
	{
		defer func() {
			if this.CasStatus(region_status_running, region_status_block) {
				this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
			}
		}()
		var v channel.IData
		var ok bool
		for {
			select {
			case v := <-this.notifyC:
				handlNewMember := func(m channelRegionMember) {
					atomic.AddInt32(&this.running, 1)
					this.chs[m.index].consumer = m.m.consumer
					this.chs[m.index].name = m.m.name
					this.chs[m.index].c = m.m.c
					atomic.StoreUint32(&this.chs[m.index].status, selectn_channel_status_ok)
				}
				switch v.(type) {
				case channelRegionMember:
					m := v.(channelRegionMember)
					handlNewMember(m)
				default:
					if this.CasStatus(region_status_pre_block, region_status_block) {
					HandleLeft:
						for {
							select {
							case vv := <-this.notifyC:
								if m, ok := vv.(channelRegionMember); ok {
									handlNewMember(m)
								}
							default:
								break HandleLeft
							}
						}
						this.getListener().NotifyListener(nil, strconv.Itoa(this.id))
					} else {
						if this.Status() != region_status_block {
							panic(PROGRAMA_ERROR)
						}
					}
					return
				}
			case v, ok = <-this.chs[0].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[0], bits: this.Bits()})
					continue
				}
				this.execute(0, v)
			case v, ok = <-this.chs[1].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[1], bits: this.Bits()})
					continue
				}
				this.execute(1, v)
			case v, ok = <-this.chs[2].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[2], bits: this.Bits()})
					continue
				}
				this.execute(2, v)
			case v, ok = <-this.chs[3].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[3], bits: this.Bits()})
					continue
				}
				this.execute(3, v)
			case v, ok = <-this.chs[4].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[4], bits: this.Bits()})
					continue
				}
				this.execute(4, v)
			case v, ok = <-this.chs[5].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[5], bits: this.Bits()})
					continue
				}
				this.execute(5, v)
			case v, ok = <-this.chs[6].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[6], bits: this.Bits()})
					continue
				}
				this.execute(6, v)
			case v, ok = <-this.chs[7].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[7], bits: this.Bits()})
					continue
				}
				this.execute(7, v)
			case v, ok = <-this.chs[8].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[8], bits: this.Bits()})
					continue
				}
				this.execute(8, v)
			case v, ok = <-this.chs[9].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[9], bits: this.Bits()})
					continue
				}
				this.execute(9, v)
			case v, ok = <-this.chs[10].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[10], bits: this.Bits()})
					continue
				}
				this.execute(10, v)
			case v, ok = <-this.chs[11].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[11], bits: this.Bits()})
					continue
				}
				this.execute(11, v)
			case v, ok = <-this.chs[12].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[12], bits: this.Bits()})
					continue
				}
				this.execute(12, v)
			case v, ok = <-this.chs[13].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[13], bits: this.Bits()})
					continue
				}
				this.execute(13, v)
			case v, ok = <-this.chs[14].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[14], bits: this.Bits()})
					continue
				}
				this.execute(14, v)
			case v, ok = <-this.chs[15].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[15], bits: this.Bits()})
					continue
				}
				this.execute(15, v)
			case v, ok = <-this.chs[16].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[16], bits: this.Bits()})
					continue
				}
				this.execute(16, v)
			case v, ok = <-this.chs[17].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[17], bits: this.Bits()})
					continue
				}
				this.execute(17, v)
			case v, ok = <-this.chs[18].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[18], bits: this.Bits()})
					continue
				}
				this.execute(18, v)
			case v, ok = <-this.chs[19].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[19], bits: this.Bits()})
					continue
				}
				this.execute(19, v)
			case v, ok = <-this.chs[20].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[20], bits: this.Bits()})
					continue
				}
				this.execute(20, v)
			case v, ok = <-this.chs[21].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[21], bits: this.Bits()})
					continue
				}
				this.execute(21, v)
			case v, ok = <-this.chs[22].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[22], bits: this.Bits()})
					continue
				}
				this.execute(22, v)
			case v, ok = <-this.chs[23].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[23], bits: this.Bits()})
					continue
				}
				this.execute(23, v)
			case v, ok = <-this.chs[24].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[24], bits: this.Bits()})
					continue
				}
				this.execute(24, v)
			case v, ok = <-this.chs[25].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[25], bits: this.Bits()})
					continue
				}
				this.execute(25, v)
			case v, ok = <-this.chs[26].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[26], bits: this.Bits()})
					continue
				}
				this.execute(26, v)
			case v, ok = <-this.chs[27].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[27], bits: this.Bits()})
					continue
				}
				this.execute(27, v)
			case v, ok = <-this.chs[28].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[28], bits: this.Bits()})
					continue
				}
				this.execute(28, v)
			case v, ok = <-this.chs[29].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[29], bits: this.Bits()})
					continue
				}
				this.execute(29, v)
			case v, ok = <-this.chs[30].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[30], bits: this.Bits()})
					continue
				}
				this.execute(30, v)
			case v, ok = <-this.chs[31].c:
				if !ok {
					this.reportCGone(&reportGcGoneInfo{id: this.id, c: this.chs[31], bits: this.Bits()})
					continue
				}
				this.execute(31, v)
			case <-c.Done():
				logplugin.Debug("ctxDone", "info", this.String())
				return
			}
		}
	}
}

func (s *select32Region) Bits() byte {
	return 1 << 5
}
