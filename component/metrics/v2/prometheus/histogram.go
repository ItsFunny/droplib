/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 2:24 下午
# @File : histogram.go
# @Description :
# @Attention :
*/
package prometheus

import (
	"github.com/hyperledger/fabric-droplib/component/metrics"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	_ metrics.Histogram = (*HistogramStator)(nil)
)

type HistogramStator struct {
	values []float64

	hv *prom.HistogramVec
	*baseStator
}

func (h *HistogramStator) With(labelValues ...string) metrics.Histogram {
	return &HistogramStator{
		hv:         h.hv,
		baseStator: h.baseStator.With(labelValues...),
	}
}

func (h *HistogramStator) WithSelf() metrics.Histogram {
	return h
}

func (h *HistogramStator) Observe(value float64) {
	panic("implement me")
}
