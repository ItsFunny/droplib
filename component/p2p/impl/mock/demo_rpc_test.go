/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/1 1:28 下午
# @File : demo_rpc_test.go.go
# @Description :
# @Attention :
*/
package mock

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/base/debug"
	logplugin "github.com/hyperledger/fabric-droplib/base/log"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	p2p "github.com/hyperledger/fabric-droplib/component/p2p/cmd"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"testing"
	"time"
)

// 测试2个节点间的通信
func Test_Direct_Commutation(t *testing.T) {
	cfg := config.NewDefaultP2PConfiguration(config.DisablePingPong())
	p1, p2, p1Service, p2Service := makeCoupleP2P(cfg, 10001, 10002)
	fmt.Println(p1)
	fmt.Println(p2)
	fmt.Println(p1Service)
	fmt.Println(p2Service)

	// p1 发送信息给p2
	p1Service.SendTo(channelId1, p2.GetNodeID())

	time.Sleep(time.Second * 10)
	p2Service.SendTo(channelId2, p1.GetNodeID())

	debug.WaitFor()
}

// 测试2个节点之间的通信,但是use cache
func Test_2_nodes_with_cache(t *testing.T) {
	cfg := config.NewDefaultP2PConfiguration(config.DisablePingPong())
	p1, p2, p1Service, p2Service := makeCoupleP2P(cfg, 10001, 10002)
	fmt.Println(p1)
	fmt.Println(p2)
	fmt.Println(p1Service)
	fmt.Println(p2Service)

	for i := 0; i < 100000; i++ {
		// p1 发送信息给p2
		p1Service.SendTo(channelId1, p2.GetNodeID())
	}

	time.Sleep(time.Second * 10)
	p2Service.SendTo(channelId2, p1.GetNodeID())

	debug.WaitFor()
}

func makeOne(cfg *config.P2PConfiguration, port int, opts ...libp2p.Option) (*p2p.P2P, *DemoLogicServiceImpl) {
	p := makeRandP2P(port, cfg, opts...)
	p.Start(services.ASYNC_START)
	pService := NewDemoLogicServiceImpl()
	p.RegisterService(pService)
	return p, pService
}
func makeCoupleP2P(cfg *config.P2PConfiguration, port1, port2 int, opts ...libp2p.Option) (*p2p.P2P, *p2p.P2P, *DemoLogicServiceImpl, *DemoLogicServiceImpl) {
	p1, p1Service := makeOne(cfg, port1, opts...)
	p2, p2Service := makeOne(cfg, port2, opts...)

	addr := p2.GetExternalAddress()
	id := p2.GetRouter().GetNode().ID()
	logplugin.Info("手动添加的节点id为:" + id.Pretty())

	m := models.MemberChanged{
		PeerId:             id,
		ExternMultiAddress: addr,
		Type:               p2ptypes.MEMBER_CHANGED_ADD,
		Ttl:                peerstore.PermanentAddrTTL,
	}
	p1.GetRouter().OutMemberChanged(m)

	return p1, p2, p1Service, p2Service
}

func makeRandP2P(port int, cfg *config.P2PConfiguration, opts ...libp2p.Option) *p2p.P2P {
	cfg.IdentityProperty.Port = port
	p := p2p.NewP2P(cfg, opts...)
	p.AppendProtocol(DemoProtocolInfo)
	return p
}
