/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/1 4:58 上午
# @File : channel.go
# @Description :
  1. 动态化监听channel: 当小于 n值时,一个channel一个routine,当 n<=v<=m 时,用反射,当 v>m 时,用selectN
  2. (这里是一个policy)
	2.1 channel 提供分组机制,既同一个group内的channel收到数据之后 同一个routine处理
	(因为如果不是这样的话,很可能10000个channel共用同一个消费routine的话,则会卡死(producer大于consumer))
	2.2 或者是统一的通过workerpool处理 (默认为worker pool)
  3.
# @Attention :
*/
package watcher

import (
	"errors"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/libs/worker/gowp/workpool"
	"reflect"
	"sync"
)

type IDataConsumer interface {
	Consume(data models.Envelope) error
}
type ChannelWatcherProperty struct {
	MaxWorker int

	// 超过则
	MaxMsgInFlight int
	// 第一阶段
	StepOneLimit int16
	StepTwoLimit int16

	// 对于特定的channel采用一个chanel 一个routine的形式
	SpecialName []string
}

func NewDefaultChannelWatcherProperty() *ChannelWatcherProperty {
	r := &ChannelWatcherProperty{
		MaxWorker:      10,
		MaxMsgInFlight: 100,
		StepOneLimit:   64,
		StepTwoLimit:   1024,
		SpecialName:    nil,
	}
	return r
}

type IChannelWrapper interface {
	ID() int
}

// 如果是反射的做法,则当取到数据之后,会有一个对应的index,通过index找到对应的group
// channel 也希望分group ,类似于
type ChannelGroup struct {
	// 所属的组,默认情况下,相同组的channel是同一个routine在跑
	Group       int8
	Name        string
	OutChannels []<-chan models.Envelope

	OnClose func()
}

func (c ChannelGroup) ValidateBasic() error {
	if c.OnClose == nil {
		return errors.New("onclose cant be nil")
	}
	if len(c.OutChannels) == 0 {
		return errors.New("channels cant be nil")
	}
	if len(c.Name) == 0 {
		return errors.New("name cant be nil")
	}
	return nil
}

type ChannelWatcher struct {
	*impl.BaseServiceImpl

	mtx           sync.RWMutex
	channelCounts int
	status        int8

	MaxBatchSize int

	groups map[int8]*OnRunningChannelGroup

	newGroupC chan *ChannelGroup

	inFlightData uint32

	dataConsumer IDataConsumer

	workerPool *workpool.WorkPool

	watcherProperty *ChannelWatcherProperty
}

func NewChannelWatcher(dataConsumer IDataConsumer,
	proper *ChannelWatcherProperty) *ChannelWatcher {
	r := &ChannelWatcher{
		BaseServiceImpl: nil,
		mtx:             sync.RWMutex{},
		groups:          make(map[int8]*OnRunningChannelGroup),
		newGroupC:       make(chan *ChannelGroup),
		dataConsumer:    dataConsumer,
		workerPool:      workpool.New(proper.MaxWorker),
		watcherProperty: proper,
	}
	r.BaseServiceImpl = impl.NewBaseService(nil, modules.NewModule("CHANNEL_WATCHER", 1), r)
	return r
}

func (this *ChannelWatcher) OnStop() {
	for _, v := range this.groups {
		v.OnClose()
	}
}

// 监听新的一组channel
func (this *ChannelWatcher) WatchNewChannelGroup(group *ChannelGroup) error {
	if err := group.ValidateBasic(); nil != err {
		return err
	}

	this.mtx.Lock()
	defer this.mtx.Unlock()

	name := group.Name
	notify := make(chan struct{})
	special := false
	for _, n := range this.watcherProperty.SpecialName {
		if name == n {
			this.routineChannel(group, notify)
			special = true
			break
		}
	}

	if !special {
		l := len(group.OutChannels)
		// 否则的话通过当前的channel数决定是否全部添加
		if this.channelCounts <= int(this.watcherProperty.StepOneLimit) && l <= int(this.watcherProperty.StepOneLimit) {
			// 一个channel 一个routine
			go func() {
				this.routineChannel(group, notify)
			}()
		} else if this.channelCounts <= int(this.watcherProperty.StepTwoLimit) && l <= int(this.watcherProperty.StepTwoLimit) {
			// 第二阶段,通过反射的形式监听
			go func() {
				this.reflectWatchChannel(group, notify)
			}()
		} else {
			// 说明数目已经很大了,则直接通过selectN
			// FIXME 但是拿全部的数据来匹配是错误的,因为除非能够将新加入的channel 也添加到之间的 channel中
			// FIXME   当分阶段监听channel的时候,既 当 A channelGroup 之后添加B channelGroup ,是否需要将B中的所有channel添加到A中
		}
	}
	<-notify
	this.groups[group.Group] = ConvGroupT2OnRunning(group)
	this.channelCounts += len(group.OutChannels)
	if this.channelCounts < int(this.watcherProperty.StepOneLimit) {
		return nil
	}
	if this.channelCounts > int(this.watcherProperty.StepOneLimit) && this.channelCounts < int(this.watcherProperty.StepTwoLimit) {
		this.status = WATCHER_STATUS_REFLECT
	} else {
		this.status = WATCHER_STATUS_SELECTN
	}

	return nil
}

// 一个channel 一个routine
func (this *ChannelWatcher) routineChannel(group *ChannelGroup, notify chan struct{}) {
	channes := group.OutChannels
	for index, ccc := range channes {
		go func(i int, ch <-chan models.Envelope) {
			for {
				select {
				case envelope, ok := <-ch:
					if !ok {
						return
					}
					// always same routine
					if err := this.dataConsumer.Consume(envelope); nil != err {
						this.Logger.Error("consume失败", "err", err.Error())
					}
				}
			}
		}(index, ccc)
	}
	notify <- struct{}{}
}

func (this *ChannelWatcher) reflectWatchChannel(group *ChannelGroup, notify chan struct{}) {
	// 说明原先处于小批量的状态
	if this.status == WATCHER_STATUS_STEP_ONE {
		this.firstReflectWatchChannel(group, notify)
	} else {
		this.newGroupC <- group
	}
}
func (this *ChannelWatcher) firstReflectWatchChannel(group *ChannelGroup, notify chan struct{}) {
	chans := group.OutChannels
	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	go func() {
		for {
			select {
			case g := <-this.newGroupC:
				for _, ch := range g.OutChannels {
					cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
				}
			default:
			}
			index, value, ok := reflect.Select(cases)
			if !ok {
				// Channel cases[i] has been closed, remove it
				// from our slice of cases and update our ids
				// mapping as well.
				cases = append(cases[:index], cases[index+1:]...)
				// ids = append(ids[:i], ids[i+1:]...)
				continue
			}
			this.inFlightData++
			// ok will be true if the channel has not been closed.
			// ch := chans[chosen]
			msg := value.Interface()
			this.Logger.Info("accquire msg", msg, "index", index)
			this.workerPool.Do(func() error {
				// sync ,FIXME: this should  be routine or work pool
				if err := this.dataConsumer.Consume(msg.(models.Envelope)); nil != err {
					this.Logger.Error("consume失败:" + err.Error())
				}
				return nil
			})
		}
	}()
	notify <- struct{}{}
}
