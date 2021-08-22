/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/22 1:24 下午
# @File : LibP2PNode.go
# @Description :
# @Attention :
*/
package streamimpl

import (
	"context"
	"github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/event/impl"
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	services2 "github.com/hyperledger/fabric-droplib/component/p2p/services"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"strings"
)

var (
	_ services2.INode = (*LibP2PNode)(nil)
)

type Option func(*LibP2PNode)
type EnvelopeFilter func(n *LibP2PNode) p2ptypes.FilterEnum

// type NewLibP2PNodeArrive struct {
// 	Addr multiaddr.Multiaddr
// 	Ttl  time.Duration
// }

// type ProtocolInfo struct {
// 	ProtocolId  p2pProtocol.ID
// 	MessageType services.IProtocolWrapper
// 	Channels    []Chann
// }
type LibP2PNode struct {
	host.Host
	ctx context.Context

	Protocols []models.ProtocolInfo

	Addr string
}

func (n *LibP2PNode) NewP2PEventBusComponent() base.IP2PEventBusComponent {
	return impl.NewLibP2PEventBusComponent(n.EventBus())
}

func (n *LibP2PNode) AddProtocol(info models.ProtocolInfo) {
	// FIXME ? SYNC? wait ready?
	n.Protocols = append(n.Protocols, info)
}

func (n *LibP2PNode) NewPubSub() services2.IPubSubManager {
	return NewLibP2PPubSubManager(n.ctx, n.Host)
}

func (n *LibP2PNode) NewStreamRegistry() services2.IStreamRegistry {
	return NewLibP2PStreamRegistry(n.ctx, n.Host)
}

func (n *LibP2PNode) FromExternAddress(str string) types.NodeID {
	arrays := strings.Split(str, "/")
	id, _ := peer.Decode(arrays[len(arrays)-1])
	return types.NodeID(id)
}

func (n *LibP2PNode) SetStreamHandler(pid protocol.ID, serv services2.IProtocolService) {
	n.Host.SetStreamHandler(pid, serv.(services2.ILibP2PProtocolService).ReceiveLibP2PStream)
}

func (n *LibP2PNode) GetProtocols() []models.ProtocolInfo {
	return n.Protocols
}

// func (n *LibP2PNode) GetBus() services.IEventBus {
// 	return n.Host.EventBus()
// }

func (n *LibP2PNode) ID() types.NodeID {
	return types.NodeID(n.Host.ID())
}

func (n *LibP2PNode) GetExternAddress() string {
	addrs := n.Host.Addrs()
	addr := ""
	if len(addrs) == 1 {
		addr = addrs[0].String()
	} else {
		for _, ar := range addrs {
			if strings.Contains(ar.String(), "127.0.0.1") {
				continue
			} else {
				addr = ar.String()
			}
		}
	}
	if len(addr) == 0 {
		panic("无法获取到监听地址,panic")
	}
	return addr + "/p2p/" + n.Host.ID().Pretty()
}
func (n *LibP2PNode) String() string {
	sb := strings.Builder{}
	sb.WriteString("本节点id为:" + n.Host.ID().Pretty())
	adds := make([]string, 0)
	for _, addr := range n.Host.Addrs() {
		adds = append(adds, addr.String())
	}
	sb.WriteString(",监听的地址为:" + strings.Join(adds, ","))
	return sb.String()
}

func (n *LibP2PNode) AppendProtocol(info models.ProtocolInfo) {
	n.Protocols = append(n.Protocols, info)
}
