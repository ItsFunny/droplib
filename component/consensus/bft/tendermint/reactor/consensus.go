/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/11 12:49 下午
# @File : consensus.go
# @Description :
# @Attention :
*/
package reactor

import "github.com/hyperledger/fabric-droplib/component/switch/services"

type IConsensusReactor interface {
	services.IReactor
}
