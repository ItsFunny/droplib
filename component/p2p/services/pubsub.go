/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 11:06 上午
# @File : pubsub.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/component/pubsub/services"
)

type IPubSubManager interface {
	services.IPubSubComponent
}
