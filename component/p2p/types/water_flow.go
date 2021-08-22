/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/31 1:25 下午
# @File : water_flow.go
# @Description :
# @Attention :
*/
package p2ptypes

type WaterFlow uint8

const (
	WATER_FLOW_AS_USUAL WaterFlow= 0
	WATER_FLOW_STOP     WaterFlow = 1
)
