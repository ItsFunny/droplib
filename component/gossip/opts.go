/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 2:58 下午
# @File : opts.go
# @Description :
# @Attention :
*/
package gossip

type GossipOption struct {
	filters []MessageFilter
}
type GossipOpt func(g *GossipOption)

func GossipWithFilters(messageFilter ...MessageFilter) GossipOpt {
	return func(g *GossipOption) {
		g.filters = messageFilter
	}
}
