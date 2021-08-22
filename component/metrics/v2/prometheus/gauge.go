/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 1:32 下午
# @File : gauge.go
# @Description :
# @Attention :
*/
package prometheus

import (
	"github.com/hyperledger/fabric-droplib/component/metrics"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	_ metrics.Gauge = (*GaugeStator)(nil)
)

type GaugeStator struct {
	gv *prom.GaugeVec
	*baseStator
}

func (g *GaugeStator) With(labelValues ...string) metrics.Gauge {
	return &GaugeStator{
		gv:         g.gv,
		baseStator: g.baseStator.With(labelValues...),
	}
}

func (g *GaugeStator) WithSelf() metrics.Gauge {
	return g
}

func (g *GaugeStator) Add(delta float64) {
	g.gv.With(makeLabels(g.lvs...)).Add(delta)
}

func (g *GaugeStator) Set(value float64) {
	g.gv.With(makeLabels(g.lvs...)).Set(value)
}

func (g *GaugeStator) OnDescribe(descs chan<- *prom.Desc) {
	g.gv.Describe(descs)
}

func makeLabels(labelValues ...string) prom.Labels {
	labels := prom.Labels{}
	for i := 0; i < len(labelValues); i += 2 {
		labels[labelValues[i]] = labelValues[i+1]
	}
	return labels
}

func NewGaugeStator(o metrics.GaugeOpts) *GaugeStator {
	r := &GaugeStator{
		gv: prom.NewGaugeVec(prom.GaugeOpts{
			Namespace: o.Namespace,
			Subsystem: o.Subsystem,
			Name:      o.Name,
			Help:      o.Help,
		},
			o.LabelNames),
	}
	return r
}
