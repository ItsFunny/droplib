// /*
// # -*- coding: utf-8 -*-
// # @Author : joker
// # @Time : 2021/3/22 1:24 下午
// # @File : node.go
// # @Description :
// # @Attention :
// */
package models

import (
	"github.com/hyperledger/fabric-droplib/base/types"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	"time"
)

//
// import (
// 	"github.com/hyperledger/fabric-droplib/base/services"
// 	"github.com/hyperledger/fabric-droplib/base/types"
// 	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
// 	"github.com/libp2p/go-libp2p-core/host"
// 	"github.com/libp2p/go-libp2p-core/peer"
// 	p2pProtocol "github.com/libp2p/go-libp2p-core/protocol"
// 	"github.com/multiformats/go-multiaddr"
// 	"strings"
// 	"time"
// )
//
// type Option func(*Node)
// type EnvelopeFilter func(n *Node) p2ptypes.FilterEnum
//
type MemberChanged struct {
	PeerId             types.NodeID
	ExternMultiAddress string
	Type               p2ptypes.MEMBER_CHANGED
	Ttl                time.Duration

	Callback func()
}
// type NewNodeArrive struct {
// 	Addr multiaddr.Multiaddr
// 	Ttl  time.Duration
// }
// type ChannelInfo struct {
// 	ChannelId   types.ChannelID
// 	ChannelSize int
// 	MessageType services.IChannelMessageWrapper
// 	Topics []TopicInfo
// }
//
// type ProtocolInfo struct {
// 	ProtocolId  p2pProtocol.ID
// 	MessageType services.IProtocolWrapper
// 	Channels    []ChannelInfo
// }
// type Node struct {
// 	host.Host
//
// 	Protocols []ProtocolInfo
//
// 	Id   types.NodeID
// 	Addr string
// }
//
// func (n *Node) GetExternAddress() string {
// 	addrs := n.Host.Addrs()
// 	addr := ""
// 	if len(addrs) == 1 {
// 		addr = addrs[0].String()
// 	} else {
// 		for _, ar := range addrs {
// 			if strings.Contains(ar.String(), "127.0.0.1") {
// 				continue
// 			} else {
// 				addr = ar.String()
// 			}
// 		}
// 	}
// 	if len(addr) == 0 {
// 		panic("无法获取到监听地址,panic")
// 	}
// 	return addr + "/p2p/" + n.ID().Pretty()
// }
// func (n *Node) String() string {
// 	sb := strings.Builder{}
// 	sb.WriteString("本节点id为:" + n.Host.ID().Pretty())
// 	adds := make([]string, 0)
// 	for _, addr := range n.Host.Addrs() {
// 		adds = append(adds, addr.String())
// 	}
// 	sb.WriteString(",监听的地址为:" + strings.Join(adds, ","))
// 	return sb.String()
// }
//
// func (n *Node) AppendProtocol(info ProtocolInfo) {
// 	n.Protocols = append(n.Protocols, info)
// }
//
// type TopicInfo struct {
// 	Topic string
// }
