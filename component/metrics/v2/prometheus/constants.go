/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 1:48 下午
# @File : constants.go
# @Description :
# @Attention :
*/
package prometheus


type MetricUnion int
type MetricType MetricUnion
type MetricsMask MetricUnion


const (
	COUNT     MetricType = 1 << 0
	GAUGE     MetricType = 1 << 1
	HISTOGRAM MetricType = 1 << 2

	max_value MetricsMask = 1 << 3
	min_value MetricsMask = 1 << 4
	avg_value MetricsMask = 1 << 5

	COUNT_MAX_VALUE = MetricsMask(COUNT) | max_value
	COUNT_MIN_VALUE = MetricsMask(COUNT) | min_value
	COUNT_AVG_VALUE = MetricsMask(COUNT) | avg_value

	GAUGE_MAX_VALUE = MetricsMask(GAUGE) | max_value
	GAUGE_MIN_VALUE = MetricsMask(GAUGE) | min_value
	GAUGE_AVG_VALUE = MetricsMask(GAUGE) | avg_value

	HISTOGRAM_MAX_VALUE = MetricsMask(HISTOGRAM) | max_value
	HISTOGRAM_MIN_VALUE = MetricsMask(HISTOGRAM) | min_value
	HISTOGRAM_AVG_VALUE = MetricsMask(HISTOGRAM) | avg_value
)

