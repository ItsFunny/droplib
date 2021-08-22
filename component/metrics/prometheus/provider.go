/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package prometheus

import (
	kitmetrics "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/hyperledger/fabric-droplib/component/metrics"
	prom "github.com/prometheus/client_golang/prometheus"
)

type Provider struct{}

func (p *Provider) NewCounter(o metrics.CounterOpts) metrics.Counter {
	return &Counter{
		Counter: prometheus.NewCounterFrom(
			prom.CounterOpts{
				Namespace: o.Namespace,
				Subsystem: o.Subsystem,
				Name:      o.Name,
				Help:      o.Help,
			},
			o.LabelNames,
		),
		BaseMetrics: BaseMetrics{selfLableCouple: o.SelfLableCouple},
	}
}

func (p *Provider) NewGauge(o metrics.GaugeOpts) metrics.Gauge {
	return &Gauge{
		Gauge: prometheus.NewGaugeFrom(
			prom.GaugeOpts{
				Namespace: o.Namespace,
				Subsystem: o.Subsystem,
				Name:      o.Name,
				Help:      o.Help,
			},
			o.LabelNames,
		),
		BaseMetrics: BaseMetrics{selfLableCouple: o.SelfLableCouple},
	}
}

func (p *Provider) NewHistogram(o metrics.HistogramOpts) metrics.Histogram {
	return &Histogram{
		Histogram: prometheus.NewHistogramFrom(
			prom.HistogramOpts{
				Namespace: o.Namespace,
				Subsystem: o.Subsystem,
				Name:      o.Name,
				Help:      o.Help,
				Buckets:   o.Buckets,
			},
			o.LabelNames,
		),
		BaseMetrics: BaseMetrics{selfLableCouple: o.SelfLableCouple},
	}
}

type BaseMetrics struct {
	selfLableCouple [][2]string

	logicType metrics.MetricLogicType
}

func (this BaseMetrics) GetLogicType() metrics.MetricLogicType {
	return this.logicType
}

func (this BaseMetrics) collectToStringArrays() []string {
	strings := make([]string, 0)
	for _, arr := range this.selfLableCouple {
		for _, v := range arr {
			strings = append(strings, v)
		}
	}
	return strings
}

type Counter struct {
	kitmetrics.Counter
	BaseMetrics
}

func (c *Counter) WithSelf() metrics.Counter {
	return c.With(c.BaseMetrics.collectToStringArrays()...)
}

func (c *Counter) AddWithSelfLabels(v float64) {
	if len(c.selfLableCouple) > 0 {
		c.With(c.BaseMetrics.collectToStringArrays()...).Add(v)
	}
}
func (c *Counter) With(labelValues ...string) metrics.Counter {
	return &Counter{Counter: c.Counter.With(labelValues...)}
}

type Gauge struct {
	BaseMetrics
	kitmetrics.Gauge
}

func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	return &Gauge{Gauge: g.Gauge.With(labelValues...)}
}
func (g *Gauge) WithSelf() metrics.Gauge {
	return &Gauge{Gauge: g.Gauge.With(g.BaseMetrics.collectToStringArrays()...)}
}

type Histogram struct {
	BaseMetrics
	kitmetrics.Histogram
}

func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	return &Histogram{Histogram: h.Histogram.With(labelValues...)}
}

func (h *Histogram) WithSelf() metrics.Histogram {
	return &Histogram{Histogram: h.Histogram.With(h.BaseMetrics.collectToStringArrays()...)}
}