/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/17 2:08 下午
# @File : demo.go
# @Description :
# @Attention :
*/
package main

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/metrics"
	"github.com/hyperledger/fabric-droplib/component/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type DemoMetrics struct {
	Height            metrics.Gauge
	CommitDuration    metrics.Histogram
	PayloadBufferSize metrics.Gauge
}

var (
	HeightOpts = metrics.GaugeOpts{
		Namespace:    "gossip",
		Subsystem:    "state",
		Name:         "height",
		Help:         "Current ledger height",
		LabelNames:   []string{"channel"},
		StatsdFormat: "%{#fqname}.%{channel}",
	}
	CommitDurationOpts = metrics.HistogramOpts{
		Namespace:    "gossip",
		Subsystem:    "state",
		Name:         "commit_duration",
		Help:         "Time it takes to commit a block in seconds",
		// LabelNames:   []string{"channel"},
		StatsdFormat: "%{#fqname}.%{channel}",
	}
	PayloadBufferSizeOpts = metrics.GaugeOpts{
		Namespace:    "gossip",
		Subsystem:    "payload_buffer",
		Name:         "size",
		Help:         "Size of the payload buffer",
		LabelNames:   []string{"channel"},
		StatsdFormat: "%{#fqname}.%{channel}",
	}
)

func NewDemoMetrics(p metrics.Provider) *DemoMetrics {
	r := &DemoMetrics{
		Height:            p.NewGauge(HeightOpts),
		CommitDuration:    p.NewHistogram(CommitDurationOpts),
		PayloadBufferSize: p.NewGauge(PayloadBufferSizeOpts),
	}
	return r
}

func main() {
	p := &prometheus.Provider{}
	demoMetrics := NewDemoMetrics(p)
	go func() {
		tt := time.NewTicker(time.Second * 1)
		for {
			select {
			case <-tt.C:
				fmt.Println("add 1")
				// demoMetrics.Height.With("channel","demo").Add(1)
				demoMetrics.CommitDuration.Observe(1)
			}
		}
	}()
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9001", server)
	for {
		select {}
	}
}
