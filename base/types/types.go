/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 10:18 上午
# @File : types.go
# @Description :
# @Attention :
*/
package types

// type ReactorId uint16
// 
// // ChannelID is an arbitrary channel ID.
// type ChannelID uint16
//
// type PeerIDPretty string
//
// func FromBytes(data []byte) PeerIDPretty {
// 	id, _ := peer.IDFromBytes(data)
// 	return PeerIDPretty(id.Pretty())
// }
// func (this PeerIDPretty) ToBytes() []byte {
// 	marshal, _ := this.ToPeerID().Marshal()
// 	return marshal
// }
// func (this PeerIDPretty) ToPeerID() peer.ID {
// 	decode, _ := peer.Decode(string(this))
// 	return decode
// }
// func (this PeerIDPretty) ToNodeID() NodeID {
// 	return NodeID(this.ToPeerID())
// }
//
// type NodeID peer.ID
//
// func (this NodeID) ToPretty() PeerIDPretty {
// 	return PeerIDPretty(peer.ID(this).Pretty())
// }
// func (this NodeID) Size() int {
// 	return len(string(this))
// }
// func (this NodeID) Pretty() string {
// 	return peer.ID(this).Pretty()
// }
//
// type StreamID string
//
// type ServiceID string
