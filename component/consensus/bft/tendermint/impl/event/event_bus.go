/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 3:20 下午
# @File : event_bus.go
# @Description :
# @Attention :
*/
package event

import (
	"context"
	"fmt"
	dataEvents "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/component/pubsub/impl"
)

var (
	_ services.IConsensusEventBusComponent = (*defaultEventBus)(nil)
)

const (
	EventTypeKey = "tm.event"
	// see EventBus#PublishEventTx
	TxHashKey = "tx.hash"
	// TxHeightKey is a reserved key, used to specify transaction block's height.
	// see EventBus#PublishEventTx
	TxHeightKey = "tx.height"

	// BlockHeightKey is a reserved key used for indexing BeginBlock and Endblock
	// events.
	BlockHeightKey = "block.height"
)

type defaultEventBus struct {
	*impl.CommonEventBusComponentImpl
}

const defaultCapacity = 0

func NewDefaultEventBus() *defaultEventBus {
	return NewDefaultEventBusWithBufferCapacity(defaultCapacity)
}
func NewDefaultEventBusWithComponent(compo *impl.CommonEventBusComponentImpl) *defaultEventBus {
	r := &defaultEventBus{
		CommonEventBusComponentImpl: compo,
	}
	return r
}
func NewDefaultEventBusWithBufferCapacity(cap int) *defaultEventBus {
	// capacity could be exposed later if needed
	r := &defaultEventBus{}
	r.CommonEventBusComponentImpl = impl.NewCommonEventBusComponentImpl(impl.BufferCapacity(cap))
	return r
}

func (b *defaultEventBus) PublishEventNewBlock(data dataEvents.EventDataNewBlock) error {
	// no explicit deadline for publishing events

	// resultEvents := append(data.ResultBeginBlock.Events, data.ResultEndBlock.dataEvents...)
	// events := b.validateAndStringifyEvents(resultEvents, b.Logger.With("block", data.Block.StringShort()))

	// add predefined new block event
	ctx := context.Background()
	events := make(map[string][]string)
	events[EventTypeKey] = append(events[EventTypeKey], types.EventNewBlock)
	return b.PublishWithEvents(ctx, data, events)
}

func (d *defaultEventBus) PublishEventNewBlockHeader(header dataEvents.EventDataNewBlockHeader) error {
	ctx := context.Background()
	events := make(map[string][]string)
	events[EventTypeKey] = append(events[EventTypeKey], types.EventNewBlockHeader)
	return d.PublishWithEvents(ctx, header, events)
}
func (b *defaultEventBus) Publish(eventType string, eventData interface{}) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	return b.PublishWithEvents(ctx, eventData, map[string][]string{EventTypeKey: {eventType}})
}

func (d *defaultEventBus) PublishEventNewEvidence(evidence dataEvents.EventDataNewEvidence) error {
	return d.Publish(types.EventNewEvidence, evidence)
}

func (d *defaultEventBus) PublishEventTx(data dataEvents.EventDataTx) error {
	ctx := context.Background()
	events := make(map[string][]string)
	events[EventTypeKey] = append(events[EventTypeKey], types.EventTx)
	events[TxHashKey] = append(events[TxHashKey], fmt.Sprintf("%X", types.Tx(data.Tx).Hash()))
	events[TxHeightKey] = append(events[TxHeightKey], fmt.Sprintf("%d", data.Height))
	return d.PublishWithEvents(ctx, data, events)
}

func (d *defaultEventBus) PublishEventValidatorSetUpdates(updates dataEvents.EventDataValidatorSetUpdates) error {
	return d.Publish(types.EventValidatorSetUpdates, updates)
}

func (d *defaultEventBus) PublishEventTimeoutPropose(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventTimeoutPropose, state)
}

func (d *defaultEventBus) PublishEventTimeoutWait(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventTimeoutWait, state)
}

func (d *defaultEventBus) PublishEventNewRoundStep(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventNewRoundStep, state)
}

func (d *defaultEventBus) PublishEventNewRound(round dataEvents.EventDataNewRound) error {
	return d.Publish(types.EventNewRound, round)
}

func (d *defaultEventBus) PublishEventCompleteProposal(proposal dataEvents.EventDataCompleteProposal) error {
	return d.Publish(types.EventCompleteProposal, proposal)
}

func (d *defaultEventBus) PublishEventPolka(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventPolka, state)
}

func (d *defaultEventBus) PublishEventUnlock(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventUnlock, state)
}

func (d *defaultEventBus) PublishEventRelock(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventRelock, state)
}

func (d *defaultEventBus) PublishEventLock(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventLock, state)
}

func (d *defaultEventBus) PublishEventVote(vote dataEvents.EventDataVote) error {
	return d.Publish(types.EventVote, vote)
}

func (d *defaultEventBus) PublishEventValidBlock(state dataEvents.EventDataRoundState) error {
	return d.Publish(types.EventValidBlock, state)
}
