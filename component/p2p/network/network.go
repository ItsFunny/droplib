// /*
// # -*- coding: utf-8 -*-
// # @Author : joker
// # @Time : 2021/3/25 4:39 下午
// # @File : network.go
// # @Description :
// # @Attention :
// */
package network
//
// import (
// 	"github.com/hyperledger/fabric-droplib/base/services/impl"
// 	"github.com/hyperledger/fabric-droplib/terminal"
// 	"github.com/hyperledger/fabric-droplib/terminal/config"
// 	streamimpl "github.com/hyperledger/fabric-droplib/terminal/impl"
// 	"sync"
// 	"time"
// )
//
// type Network struct {
// 	*impl.BaseServiceImpl
// 	rwMtx sync.RWMutex
//
// 	cfg         *config.P2PConfiguration
// 	nodeFactory services.INodeFactory
//
// 	router *streamimpl.Router
// }
//
// func (this *Network) OnStart() error {
// 	node := this.nodeFactory.CreateNode(this.cfg)
// 	this.router = streamimpl.NewRouter(node)
//
// 	return nil
// }
// func (this *Network) RegisterProtocol(interestProtocols []*ProtocolProperty) {
// 	for this.IsRunning() {
// 		time.Sleep(time.Millisecond * 20)
// 	}
//
// }
