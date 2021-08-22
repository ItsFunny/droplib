/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/26 1:28 下午
# @File : init.go
# @Description :
# @Attention :
*/
package structure

import (
	_ "github.com/hyperledger/fabric-droplib/structure/containers"
	_ "github.com/hyperledger/fabric-droplib/structure/lists"
	_ "github.com/hyperledger/fabric-droplib/structure/maps"
	"github.com/hyperledger/fabric-droplib/structure/stacks/linkedliststack"
	_ "github.com/hyperledger/fabric-droplib/structure/trees"
)

func init() {
	linkedliststack.New()
}
