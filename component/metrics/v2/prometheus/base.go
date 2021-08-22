/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 1:26 下午
# @File : base.go
# @Description :
# @Attention :
*/
package prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

type baseCollector struct {
	child ICollector
}

func (b *baseCollector) Describe(descs chan<- *prom.Desc) {
	b.child.OnDescribe(descs)
}

func (b *baseCollector) Collect(metrics chan<- prom.Metric) {
	panic("implement me")
}

type baseStator struct {
	lvs LabelValues

	routine bool
	self    bool

	mask MetricUnion

	// 指标的名称
	maskNames map[MetricsMask]string
}


func (this *baseStator) With(labelValues ...string) *baseStator {
	r := &baseStator{
		lvs:     this.lvs.With(labelValues...),
		routine: this.routine,
		self:    this.self,
		mask:    this.mask,
	}
	return r
}
