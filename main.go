/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/24 2:15 下午
# @File : main.go
# @Description :
# @Attention :
*/
package fabric_droplib

import (
	_ "github.com/hyperledger/fabric-droplib/component/base"
	_ "github.com/hyperledger/fabric-droplib/component/event/impl"
	_ "github.com/hyperledger/fabric-droplib/component/p2p/cmd"

	_ "github.com/hyperledger/fabric-droplib/base/services"
	_ "github.com/hyperledger/fabric-droplib/base/services/impl"
	_ "github.com/hyperledger/fabric-droplib/base/utils"
)
