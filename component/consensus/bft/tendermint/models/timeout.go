/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 4:27 下午
# @File : timeout.go
# @Description :
# @Attention :
*/
package models

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"time"
)

// internally generated messages which may update the state
type TimeoutInfo struct {
	Duration time.Duration       `json:"duration"`
	Height   uint64               `json:"height"`
	Round    int32               `json:"round"`
	Step     types.RoundStepType `json:"step"`
}

