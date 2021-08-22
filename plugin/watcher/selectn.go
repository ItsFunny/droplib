/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/11 10:25 上午
# @File : selectn.go
# @Description :
# @Attention :
*/
package watcher

import (
	"context"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/services/models"
	"github.com/hyperledger/fabric-droplib/component/listener"
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	"github.com/hyperledger/fabric-droplib/libs/channel"
	"github.com/hyperledger/fabric-droplib/structure/maps/linkedhashmap"
	"github.com/sasha-s/go-deadlock"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	region_status_ok             = 1 << 0                  // 1
	region_status_deny_out       = 1 << 1                  // 2
	region_status_new            = 1<<2 | region_status_ok // 5 就绪态
	region_status_before_restart = 1 << 3
	region_status_running        = 1<<4 | region_status_ok            // 17
	region_status_with_blocking  = 1<<5 | region_status_deny_out      // 34
	region_status_pre_block      = 1<<6 | region_status_with_blocking // 98
	region_status_block          = 1<<7 | region_status_with_blocking // 162

	region_status_removing  = 1<<8 | region_status_deny_out
	region_status_to_remove = 1<<9 | region_status_removing  // 770
	region_status_remove    = 1<<10 | region_status_removing // 1282
)
const (
	selectn_channel_status_ok      uint32 = 1 << 0                                // 1
	selectn_channel_status_closed  uint32 = 1 << 1                                // 2
	selectn_channel_status_gone    uint32 = 1<<2 | selectn_channel_status_closed  // 6
	selectn_channel_status_removed uint32 = 1<<3 | selectn_channel_status_closed  // 10
	selectn_channel_status_final   uint32 = 1<<4 | selectn_channel_status_removed // 26
	// selectn_channel_status_reusing      uint32 = 1<<5 | selectn_channel_status_final   // 58
	selectn_channel_status_reusing      uint32 = 1 << 5                                // 32
	selectn_channel_status_before_reuse uint32 = 1<<6 | selectn_channel_status_reusing // 96
	selectn_channel_status_reuse               = 1<<7 | selectn_channel_status_ok      // 129

	selectn_channel_status_waitting uint32 = 1 << 31

	default_max_region_bit        = 5
	default_max_batch_size        = 4
	default_min_region_merge_size = 3
	default_merge_max_bit         = 3
)

var (
	default_specifial_func = func(name string) bool {
		return false
	}
)

const (
	region_group_status_new = 1 << 0
)

type selectNC struct {
	c            <-chan channel.IData
	status       uint32       // 4
	name         string       // 4
	consumer     DataConsumer // 4
	releaseWhere string
}

func (this *selectNC) cas(exce, newV uint32, f func()) {
	if atomic.CompareAndSwapUint32(&this.status, exce, newV) {
		f()
	}
}

func (this *selectNC) fromMember(c ChannelMember) {
	// attention:cant change the orderer !!!
	this.consumer = c.consumer
	this.name = c.name
	this.c = c.c
}
func (this *selectNC) closed() bool {
	return atomic.LoadUint32(&this.status)&selectn_channel_status_closed >= selectn_channel_status_closed
}

type routineGoneMsg struct {
	c        *selectNC
	regionId int
	name     string
}

func (r routineGoneMsg) ID() interface{} {
	return r.name + "_regId=" + strconv.Itoa(r.regionId)
}

type regionUpdateRemove struct {
	id []IRegion
}

func (r regionUpdateRemove) ID() interface{} {
	return "regionUpdateRemove"
}

type regionGroup struct {
	deadlock.Mutex
	stats   uint32
	regions *linkedhashmap.Map
}
type selectNChannelWatcher struct {
	*baseChannelWatcher

	ctx    context.Context
	cancel func()

	regions []*regionGroup

	routineSize int32

	regionId uint32

	lastGGoneRegionBits uint32
	lastGGoneRegion     uint32

	newMemberWg sync.WaitGroup

	// property
	maxRegionBit  uint8
	tryMergeCount uint8
	mergeMaxBit   uint8

	preMergeTimeout time.Duration
}

func fromReflectChannelWatcher(ref *reflectChannelWatcher, opt Opt) (*selectNChannelWatcher, context.Context) {
	r := newSelectNChannelWatcher(opt)
	r.baseStatus = ref.baseStatus
	r.baseStatus.status = status_ok
	chMems := make([]chWp, 0)
	for i := 0; i < len(ref.chs); i++ {
		ch := ref.chs[i]
		chMems = append(chMems, chWp{
			name:     ch.name,
			c:        ch.ch,
			consumer: ch.consumer,
		})
	}
	ctx := context.WithValue(context.Background(), "channels", &chMems)
	return r, ctx
}
func newSelectNChannelWatcher(opt Opt) *selectNChannelWatcher {
	cancel, cancelFunc := context.WithCancel(context.Background())
	r := &selectNChannelWatcher{
		ctx:           cancel,
		cancel:        cancelFunc,
		tryMergeCount: default_min_region_merge_size,
		maxRegionBit:  default_max_region_bit,
		mergeMaxBit:   default_merge_max_bit,
	}
	// empty region
	r.regions = make([]*regionGroup, r.maxRegionBit+1)
	for i := byte(0); i <= r.maxRegionBit; i++ {
		r.regions[i] = &regionGroup{
			regions: linkedhashmap.New(),
		}
	}

	r.baseChannelWatcher = newBaseChannelWatcher("selectn", r, 0, opt.SelectNRollbackLimit, opt.SpinTimeMills, opt.RoutinePoolFactory, opt.Mode)

	return r
}
func newSelectNC(name string, cc <-chan channel.IData, f DataConsumer) *selectNC {
	r := &selectNC{
		c:        cc,
		status:   selectn_channel_status_ok,
		name:     name,
		consumer: f,
	}
	return r
}

func (b *selectNChannelWatcher) OnStart(c *models.StartCTX) error {
	go b.gc()
	go b.daemon()
	go b.opt()
	go b.region()
	v := c.Ctx.Value("channels")
	if nil != v {
		chmems := v.(*[]chWp)
		b.selectN(chmems)
	}
	return nil
}
func (this *selectNChannelWatcher) region() {
	defer func() { this.wg.Done() }()
	notifyC := this.internalChs[selectn_region_notify].ch.Ch
	for {
		select {
		case <-this.internalFastQuitC:
			return
		case v := <-notifyC:
			this.handleMsg(v)
		}
	}
}
func (this *selectNChannelWatcher) opt() {
	trick := false
	handleRoutineGone := func(data channel.IData) {
		atomic.AddInt32(&this.routineSize, -1)
	}
	handleRegionCreate := func(data channel.IData) {
		regCreate := data.(regionCreateOperation)
		toChs := *regCreate.toChs
		if trick {
			for _, ch := range toChs {
				this.addDelta(ChannelMember{
					name:     ch.name,
					c:        ch.c,
					consumer: ch.consumer,
				})
			}
			return
		}
		// we dont want to lock ,but we have to do it
		// we dont  care about the repeat id
		routineDelta := 0
		cs := toChs
		for _, ch := range cs {
			if ch.newM {
				routineDelta++
			}
		}
		regId := atomic.AddUint32(&this.regionId, 1)
		reg := newRegion(&newBaseArgument{
			id:          int(regId),
			chsW:        regCreate.toChs,
			reportCGone: this.reportCGone,
			listener: func() listener.IListenerComponent {
				return this.listener
			},
			routine: func() services.IRoutineComponent {
				return this.routinePool
			},
		})
		this.putNew(reg)
		size := atomic.AddInt32(&this.routineSize, int32(routineDelta))
		this.Logger.Debug("添加region", "regId", reg.Id(), "当前routineSize", size, "bits", reg.Bits())
		go reg.Start(&RegionStartCTX{
			PreStart: func() {
			},
			Ctx: this.ctx,
		})
	}
	handlRollBack := func(op operation) {
		if trick {
			panic(PROGRAMA_ERROR)
		}
		trick = true
		opt := op.data.(Opt)
		back, options := this.rollBack(opt)
		r := rollBackResp{
			wh:   back,
			opts: options,
		}

		this.listener.NotifyListener(r, listener_rollback)
	}
	handleNewMem := func(op operation) {
		msg := op.data.(ChannelMember)
		if !trick {
			this.newMemberWg.Add(1)
			go func() {
				// we dont care about the prepare start member
				this.listenNewMember(msg)
			}()
			return
		}
		this.addDelta(msg)
	}
	handleRegionRemove := func(op operation) {
		if trick {
			return
		}
		msg := op.data.(regionUpdateRemove)
		for index := range msg.id {
			reg := msg.id[index]
			g := this.getRegionGroup(reg.Bits())
			id := reg.Id()
			if uint32(id) == this.lastGGoneRegion {
				atomic.CompareAndSwapUint32(&this.lastGGoneRegion, uint32(id), 0)
			}
			if reg.CasStatus(region_status_to_remove, region_status_remove) || reg.Status() == region_status_block {
				g.Lock()
				_, b := g.regions.RemoveWithReturn(id)
				g.Unlock()
				if !b {
					panic(PROGRAMA_ERROR)
				}
				size := this.Size()
				this.Logger.Debug("删除region", "bits", reg.Bits(), "regId", id, "当前routineSize", size)
			}
		}
	}

	for {
		select {
		case msg, ok := <-this.operation:
			if !ok {
				return
			}
			this.dec(msg.data)
			switch msg.opType {
			case op_routine_gone:
				handleRoutineGone(msg.data)
			case op_rollback:
				handlRollBack(msg)
			case op_new_member:
				handleNewMem(msg)
			case op_region_remove:
				handleRegionRemove(msg)
			case op_region_create:
				handleRegionCreate(msg.data)
			}
		}
	}
}
func (this *selectNChannelWatcher) GetChannelShims(cap int) (map[channel.ChannelID]*ChannelWp, int) {
	r := make(map[channel.ChannelID]*ChannelWp)
	r[selectn_region_notify] = &ChannelWp{
		ch: &channel.Channel{
			Id: selectn_region_notify,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			this.handleMsg(v)
		},
	}
	r[upgradeRollbackNotifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: upgradeRollbackNotifyC,
			Ch: make(chan channel.IData),
		},
		flush: func(v channel.IData) {
			this.handleMsg(v)
		},
	}
	r[selectn_notifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: selectn_notifyC,
			Ch: make(chan channel.IData, cap),
		},
		flush: func(v channel.IData) {
			this.handleMsg(v)
		},
	}
	r[memberNotifyC] = &ChannelWp{
		ch: &channel.Channel{
			Id: memberNotifyC,
			Ch: make(chan channel.IData, 2),
		},
		flush: func(v channel.IData) {
			this.handleMsg(v)
		},
	}

	return r, 3
}
func (this *selectNChannelWatcher) rollBack(opt Opt) (ChannelWatcher, []models.StartOption) {
	this.newMemberWg.Wait()
	// FIXME BETTER WAY : COW
	r := newReflectChannelWatcher(opt)
	r.baseStatus = this.baseStatus
	r.baseStatus.status = status_ok
	for _, regG := range this.regions {
		it := regG.regions.Iterator()
		regG.Lock()
		copys := make([]IRegion, 0)
		for it.Next() {
			reg := it.Value().(IRegion)
			copys = append(copys, reg)
		}
		regG.Unlock()

		for _, reg := range copys {
			if reg.Status()&region_status_removing >= region_status_removing {
				continue
			}
			if !reg.PreBlock(context.Background()) {
				continue
			}
			chs := reg.Chs()
			// 补齐
			for _, ch := range chs {
				if ch.closed() {
					continue
				}
				r.chs = append(r.chs, newReflectC(ch.name, ch.c, ch.consumer))
			}
		}
	}

	return r, nil
}
func (this *selectNChannelWatcher) OnStop(ctx *models.StopCTX) {
	for _, reg := range this.regions {
		reg.regions.Clear()
	}
	this.regions = nil
}

func (this *selectNChannelWatcher) OnUpgrade(opt Opt) (ChannelWatcher, []models.StartOption) {
	panic("not supported yet")
}

type rollBackResp struct {
	wh   ChannelWatcher
	opts []models.StartOption
}

func (this *selectNChannelWatcher) OnRollBack(opt Opt) (ChannelWatcher, []models.StartOption) {
	registerListener := this.listener.RegisterListener(listener_rollback)
	this.sendMsg(upgradeRollbackNotifyC, rollback{opt: Opt{}})
	res := <-registerListener
	v := res.(rollBackResp)
	return v.wh, v.opts
}
func (this *selectNChannelWatcher) gc() {
	defer func() { this.wg.Done() }()
	notifyC := this.internalChs[selectn_notifyC].ch.Ch
	for {
		select {
		case <-this.internalFastQuitC:
			return
		case v := <-notifyC:
			this.handleMsg(v)
		}
	}
}
func (this *selectNChannelWatcher) daemon() {
	defer func() { this.wg.Done() }()
	memberNotifyC := this.internalChs[memberNotifyC].ch.Ch
	upgradeRollbackNotifyC := this.internalChs[upgradeRollbackNotifyC].ch.Ch
	for {
		select {
		case m := <-upgradeRollbackNotifyC:
			this.handleMsg(m)
		case m := <-memberNotifyC:
			this.handleMsg(m)
		case <-this.internalFastQuitC:
			return
		}
	}
}
func (this *selectNChannelWatcher) registerAndStartRegion(chs []chWp) {
	this.addNewRegion(&chs)
}
func (this *selectNChannelWatcher) putNew(reg IRegion) {
	rg := this.getRegionGroup(reg.Bits())
	rg.Lock()
	defer rg.Unlock()
	rg.regions.Put(reg.Id(), reg)
}

type regionCreate struct {
	toChs *[]chWp
}

func (r regionCreate) ID() interface{} {
	return "regionCreate"
}

type regionCreateOperation struct {
	toChs *[]chWp
}

func (r regionCreateOperation) ID() interface{} {
	return "regionCreateOperation"
}

func (this *selectNChannelWatcher) addNewRegion(toChs *[]chWp) {
	this.sendMsg(selectn_region_notify, regionCreate{toChs: toChs})
}

func (this *selectNChannelWatcher) getRegionGroup(bits byte) *regionGroup {
	return this.regions[getLeftShift(bits)]
}
func getLeftShift(bits byte) byte {
	r := 0
	for bits > 0 {
		r++
		bits = bits & (bits - 1)
	}
	return byte(r) - 1
}
func (this *selectNChannelWatcher) listenNewMember(c ChannelMember) {
	defer this.newMemberWg.Done()
	if this.escape(c) {
		return
	}

	var lastRegion []IRegion

	i := 0
	for index := uint8(0); index <= this.mergeMaxBit; index++ {
		reg := this.regions[index]
		reg.Lock()
		if reg.regions.Size() == 0 {
			reg.Unlock()
			continue
		}
		iterator := reg.regions.ReverseIterator()
		for iterator.Prev() && i < int(this.tryMergeCount) {
			lastRegion = append(lastRegion, iterator.Value().(IRegion))
			i++
		}
		reg.Unlock()
		if i >= int(this.tryMergeCount) {
			break
		}
		index++
	}

	for i := len(lastRegion) - 1; i >= 0; i-- {
		region := lastRegion[i]
		bits := region.Bits()
		if bits > 1<<this.mergeMaxBit {
			lastRegion = append(lastRegion[:i], lastRegion[i+1:]...)
		}
	}
	this.mergeRegion(c, lastRegion...)
}

func (this *selectNChannelWatcher) escape(c ChannelMember) bool {
	if this.lastGGoneRegion == 0 || atomic.LoadUint32(&this.status)&status_deny_memchanged >= status_deny_memchanged {
		return false
	}
	regG := this.regions[this.lastGGoneRegionBits]
	regG.Lock()
	v, exist := regG.regions.Get(this.lastGGoneRegion)
	regG.Unlock()
	if !exist {
		return false
	}
	reg := v.(IRegion)
	_, reuse := reg.Reuse(this.ctx, c)
	if reuse {
		atomic.AddInt32(&this.routineSize, 1)
	}
	return reuse
}
func selectNWaitUntil(key string, interval int, c *selectNC, e uint32) {
	f := func() bool {
		if atomic.LoadUint32(&c.status) == uint32(e) {
			return true
		}
		return false
	}

	for {
		if f() {
			return
		}
		logplugin.MDebug(selectnModule, "spinning", key, c.status, e)
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}
func (this *selectNChannelWatcher) mergeRegion(m ChannelMember, regions ...IRegion) {
	regionSt := RegionSort(regions)
	sort.Sort(regionSt)
	delId := make([]IRegion, 0)
	newChMembers := make([]chWp, 0)
	newChMembers = append(newChMembers, chWp{
		c:        m.c,
		consumer: m.consumer,
		name:     m.name,
		newM:     true,
	})
	var t byte

	for _, reg := range regionSt {
		bits := reg.Bits()
		t = bits + byte(len(newChMembers))
		if t >= bits<<1 && t&(t-1) == 0 && t <= 1<<this.maxRegionBit {
			if !reg.PreBlock(nil) {
				continue
			}
			delId = append(delId, reg)
			chs := reg.Chs()
			for _, c := range chs {
				if c.closed() {
					continue
				}
				if atomic.LoadUint32(&c.status) == selectn_channel_status_before_reuse {
					selectNWaitUntil(c.name, this.waitStatusIntervalMillSeconds, c, selectn_channel_status_ok)
				}
				newChMembers = append(newChMembers, chWp{
					c:        c.c,
					name:     c.name,
					consumer: c.consumer,
				})
			}
		}
	}
	this.selectN(&newChMembers)
	if len(delId) > 0 {
		this.sendMsg(selectn_notifyC, regionUpdateRemove{
			id: delId,
		})
	}
}

func (this *selectNChannelWatcher) selectN(chsPo *[]chWp) {
	i := 0
	chs := *chsPo
	l := len(chs)
	max := 1 << this.maxRegionBit
	for i < len(chs) {
		l = len(chs) - i
		switch {
		case l > 31 && max >= 32:
			this.registerAndStartRegion(chs[i : i+32])
			i += 32
		case l > 15 && max >= 16:
			this.registerAndStartRegion(chs[i : i+16])
			i += 16
		case l > 7 && max >= 8:
			this.registerAndStartRegion(chs[i : i+8])
			i += 8
		case l > 3 && max >= 4:
			this.registerAndStartRegion(chs[i : i+4])
			i += 4
		case l > 1 && max >= 2:
			this.registerAndStartRegion(chs[i : i+2])
			i += 2
		case l > 0:
			this.registerAndStartRegion([]chWp{chs[i]})
			i += 1
		}
	}
}

func (this *selectNChannelWatcher) reportCGone(info *reportGcGoneInfo) {
	id := info.id
	atomic.StoreUint32(&this.lastGGoneRegion, uint32(id))
	atomic.StoreUint32(&this.lastGGoneRegionBits, uint32(getLeftShift(info.bits)))
	if info.reg != nil {
		this.sendMsg(selectn_notifyC, regionUpdateRemove{id: []IRegion{info.reg}})
	} else {
	}
	if nil != info.c {
		this.sendMsg(selectn_notifyC, routineGoneMsg{
			c:        info.c,
			regionId: id,
			name:     info.c.name,
		})
	}
}

// not safe
func (this *selectNChannelWatcher) getSelectNByName(name string) *selectNC {
	for _, reg := range this.regions {
		it := reg.regions.Iterator()
		for it.Next() {
			chs := it.Value().(IRegion).Chs()
			for _, ch := range chs {
				if ch.name == name {
					return ch
				}
			}
		}
	}
	return nil
}

// not thread safe
func (this *selectNChannelWatcher) Size() int {
	return int(atomic.LoadInt32(&this.routineSize))
}

// unsafe
func (this *selectNChannelWatcher) GetRegionSize() int {
	r := 0
	for _, reg := range this.regions {
		r += reg.regions.Size()
	}
	return r
}
func (this *selectNChannelWatcher) printRegionDebug() string {
	// it := this.regions.Iterator()
	// r := ""
	// for it.Next() {
	// 	v := it.Value().(IRegion)
	// 	r += strconv.Itoa(v.Id()) + ","
	// }
	// return r
	return ""
}

func (this *selectNChannelWatcher) internalGetLastRegion() IRegion {
	// r, v := this.regions.GetLastInsert()
	// if !v {
	// 	return nil
	// }
	// return r.(IRegion)
	return nil
}

func (this *selectNChannelWatcher) handleMsg(v interface{}) {
	switch msg := v.(type) {
	case rollback:
		this.sendOperation(op_rollback, msg.opt)
	case ChannelMember:
		this.sendOperation(op_new_member, msg)
	case routineGoneMsg:
		this.sendOperation(op_routine_gone, msg)
	case regionUpdateRemove:
		this.sendOperation(op_region_remove, msg)
	case regionCreate:
		this.sendOperation(op_region_create, regionCreateOperation{toChs: msg.toChs})
	}
}
func (this *selectNChannelWatcher) handleRegionUpdateRemove(msg regionUpdateRemove) {
	type regG struct {
		g       *regionGroup
		regions []IRegion
	}
	m := make(map[byte]*regG)
	regs := make([]*regG, 0)
	for _, id := range msg.id {
		v, exist := m[id.Bits()]
		if !exist {
			v = &regG{
				g: this.getRegionGroup(id.Bits()),
			}
			m[id.Bits()] = v
			regs = append(regs, v)
		}
		v.regions = append(v.regions, id)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(regs))
	for index := range regs {
		go func(index int) {
			defer wg.Done()
			regG := regs[index]
			g := regG.g
			for _, reg := range regG.regions {
				id := reg.Id()
				if uint32(id) == this.lastGGoneRegion {
					atomic.CompareAndSwapUint32(&this.lastGGoneRegion, uint32(id), 0)
				}
				if reg.CasStatus(region_status_to_remove, region_status_remove) || reg.Status() == region_status_block {
					g.Lock()
					_, b := g.regions.RemoveWithReturn(id)
					g.Unlock()

					if !b {
						panic(PROGRAMA_ERROR)
					}
					size := this.Size()
					this.Logger.Debug("删除region", "bits", reg.Bits(), "id", id, "当前routineSize", size)
				}
			}
		}(index)
	}
	wg.Wait()
}
