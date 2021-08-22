/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:19 上午
# @File : event_bus.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	services2 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
)

// BlockEventPublisher publishes all block related events
type IEventBusPublisher interface {
	PublishEventNewBlock(block events.EventDataNewBlock) error
	PublishEventNewBlockHeader(header events.EventDataNewBlockHeader) error
	PublishEventNewEvidence(evidence events.EventDataNewEvidence) error
	PublishEventTx(events.EventDataTx) error
	PublishEventValidatorSetUpdates(events.EventDataValidatorSetUpdates) error
	// PublishFabricEventTx(cb []*cb.Envelope, senderId uint16, senderPeerAddr string) error
}

type IConsensusEventBusComponent interface {
	services.IBaseService
	IEventBusPublisher
	services2.ICommonEventBusComponent

	PublishEventTimeoutPropose(state events.EventDataRoundState) error
	PublishEventTimeoutWait(state events.EventDataRoundState) error
	PublishEventNewRoundStep(state events.EventDataRoundState) error
	PublishEventNewRound(events.EventDataNewRound) error
	PublishEventCompleteProposal(proposal events.EventDataCompleteProposal)error
	PublishEventPolka(state events.EventDataRoundState) error
	PublishEventUnlock(state events.EventDataRoundState) error
	PublishEventRelock(state events.EventDataRoundState) error
	PublishEventLock(state events.EventDataRoundState) error
	PublishEventVote(vote events.EventDataVote)error
	PublishEventValidBlock(state events.EventDataRoundState)error
}
