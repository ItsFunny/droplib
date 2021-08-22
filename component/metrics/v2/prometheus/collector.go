/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 1:25 下午
# @File : collector.go
# @Description :
# @Attention :
*/
package prometheus

import prom "github.com/prometheus/client_golang/prometheus"

type ICollector interface {
	OnDescribe(descs chan<- *prom.Desc)
}
