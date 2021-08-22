/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/6/20 8:35 上午
# @File : helper.go
# @Description :
# @Attention :
*/
package listener

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/pubsub/models"
	"github.com/hyperledger/fabric-droplib/component/pubsub/services"
)
const (
	EventTypeKey="event.listsner"
)
func QueryForEvent(eventType string) services.Query {
	return models.MustParse(fmt.Sprintf("%s='%s'", EventTypeKey, eventType))
}
