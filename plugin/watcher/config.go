/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/8/13 4:55 下午
# @File : config.go
# @Description :
# @Attention :
*/
package watcher

import (
	"github.com/hyperledger/fabric-droplib/component/routine/services"
	v2 "github.com/hyperledger/fabric-droplib/component/routine/services/v2"
)

var (
	config_channel_cap = "channel_cap"
)

type Option = func(opt *Opt)

type Opt struct {
	RoutineUpgradeLimit int32

	ReflectRollbackLimit int32
	ReflectUpgradeLimit  int32

	SelectNRollbackLimit int32
	// SelectNMaxBatchSize       uint8
	SelectNMinRegionMergeSize uint8

	SpinTimeMills int32

	ChannelCap int

	RoutinePoolFactory func() services.IRoutineComponent
	Mode               byte
	SpecialFunc        func(name string) bool
}

func (this Opt) ID() interface{} {
	return "opt"
}

func (this Opt) ToOptions() []Option {
	r := make([]Option, 0)
	r = append(r,
		RoutineUpgradeLimitOption(int(this.RoutineUpgradeLimit)),
		ReflectRollBackOption(int(this.ReflectRollbackLimit)),
		ReflectUpgradeLimitOption(int(this.ReflectUpgradeLimit)),
		SelectNRollBackOption(int(this.SelectNRollbackLimit)),
		SelectNMinRegionMergeSizeOption(this.SelectNMinRegionMergeSize),
		SpinTimeMillsOption(int(this.SpinTimeMills)),
		ChannelCapOption(this.ChannelCap),
		RoutinePoolFactoryOption(this.RoutinePoolFactory),
		SpecialFuncOption(this.SpecialFunc),
	)
	return r
}

var SpecialFuncOption = func(f func(name string) bool) Option {
	return func(opt *Opt) {
		opt.SpecialFunc = f
	}
}
var ChannelCapOption = func(cap int) Option {
	return func(opt *Opt) {
		opt.ChannelCap = cap
	}
}

var RoutinePoolFactoryOption = func(f func() services.IRoutineComponent) Option {
	return func(opt *Opt) {
		opt.RoutinePoolFactory = f
	}
}
var SpinTimeMillsOption = func(slp int) Option {
	return func(opt *Opt) {
		opt.SpinTimeMills = int32(slp)
	}
}
var RoutineUpgradeLimitOption = func(count int) Option {
	return func(opt *Opt) {
		opt.RoutineUpgradeLimit = int32(count)
	}
}
var ReflectRollBackOption = func(count int) Option {
	return func(opt *Opt) {
		opt.ReflectRollbackLimit = int32(count)
	}
}
var ReflectUpgradeLimitOption = func(count int) Option {
	return func(opt *Opt) {
		opt.ReflectUpgradeLimit = int32(count)
	}
}
var SelectNRollBackOption = func(count int) Option {
	return func(opt *Opt) {
		opt.SelectNRollbackLimit = int32(count)
	}
}
var SelectNMinRegionMergeSizeOption = func(size uint8) Option {
	return func(opt *Opt) {
		opt.SelectNMinRegionMergeSize = size
	}
}

func GetOption(ops ...Option) Opt {
	r := DefaultOpt
	rr := &r
	for _, o := range ops {
		o(rr)
	}
	if r.ChannelCap <= 0 {
		r.ChannelCap = default_channel_cap
	}
	if r.SpecialFunc == nil {
		r.SpecialFunc = func(name string) bool {
			return false
		}
	}
	return r
}

var DefaultOpt = Opt{
	RoutineUpgradeLimit:  DEFAULT_STEP_ONE_LIMIT,
	ReflectRollbackLimit: DEFAULT_STEP_ONE_LIMIT,
	ReflectUpgradeLimit:  DEFAULT_STEP_TWO_LIMIT,
	SelectNRollbackLimit: DEFAULT_STEP_TWO_LIMIT,
	// SelectNMaxBatchSize:       default_max_batch_size,
	SelectNMinRegionMergeSize: default_min_region_merge_size,
	SpinTimeMills:             1000,
	ChannelCap:                default_channel_cap,
	RoutinePoolFactory: func() services.IRoutineComponent {
		return v2.NewV2RoutinePoolExecutorComponent(v2.WithSize(default_routine_pool_size))
	},
	SpecialFunc: default_specifial_func,
}

var DefaultForeverOpt = Opt{
	RoutineUpgradeLimit:  0,
	ReflectRollbackLimit: 0,
	ReflectUpgradeLimit:  0,
	SelectNRollbackLimit: 0,
	// SelectNMaxBatchSize:       0,
	SelectNMinRegionMergeSize: 0,
	SpinTimeMills:             1000,
	ChannelCap:                default_channel_cap,
	RoutinePoolFactory: func() services.IRoutineComponent {
		return v2.NewV2RoutinePoolExecutorComponent(v2.WithSize(default_routine_pool_size))
	},
}

func foreverOptions() []Option {
	r := make([]Option, 0)
	r = append(r, RoutineUpgradeLimitOption(0),
		ReflectRollBackOption(0),
		ReflectUpgradeLimitOption(0),
		SelectNRollBackOption(0),
	)
	return r
}
func commonOptions() []Option {
	r := make([]Option, 0)
	r = append(r, RoutineUpgradeLimitOption(DEFAULT_STEP_ONE_LIMIT),
		ReflectRollBackOption(DEFAULT_STEP_ONE_LIMIT),
		ReflectUpgradeLimitOption(DEFAULT_STEP_TWO_LIMIT),
		SelectNRollBackOption(DEFAULT_STEP_TWO_LIMIT),
		SelectNMinRegionMergeSizeOption(default_min_region_merge_size),
		ChannelCapOption(default_channel_cap),
		RoutinePoolFactoryOption(func() services.IRoutineComponent {
			return v2.NewV2RoutinePoolExecutorComponent(v2.WithSize(default_routine_pool_size))
		}),
		SpecialFuncOption(default_specifial_func),
	)
	return r
}

var ReflectNoRollbackOption = func() Option {
	return func(opt *Opt) {
		opt.ReflectRollbackLimit = 0
	}
}

var SelectNNoRollBackOption = func() Option {
	return func(opt *Opt) {
		opt.SelectNRollbackLimit = 0
	}
}
