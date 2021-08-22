/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/1 2:50 下午
# @File : member_changed.go
# @Description :
# @Attention :
*/
package protocol

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/libp2p/go-libp2p-core/protocol"
)

const (
	MEMBERMANAGER_PROTOCOL protocol.ID = "/membermanager/1.0.0"
)

var (
	MEMBERMANAGER_PROTOCOL_WRAPPER services.IProtocolWrapper = &protos.StreamCommonProtocolWrapper{}
)

func main() {
}