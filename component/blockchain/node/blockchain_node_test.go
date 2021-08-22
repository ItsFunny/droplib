/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/19 6:46 下午
# @File : blockchain_node_test.go.go
# @Description :
# @Attention :
*/
package node

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"testing"
)

func TestNewNode(t *testing.T) {
	cfg := config.NewDefaultGlobalConfiguration()
	node := makeRandNode(10000, cfg)
	node.Start(services.ASYNC_START)

	for {
		select {}
	}
}
func makeRandNode(port int, cfg *config.GlobalConfiguration) *Node {
	cfg.P2PConfiguration.IdentityProperty.Port = port
	node := NewNode(cfg)
	return node
}
