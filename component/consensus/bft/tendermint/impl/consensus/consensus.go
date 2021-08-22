/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:27 上午
# @File : consensus_reactor.go
# @Description :
# @Attention :
*/
package consensus

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services2 "github.com/hyperledger/fabric-droplib/base/services"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/common/errors"
	models3 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	streamimpl2 "github.com/hyperledger/fabric-droplib/component/p2p/impl"
	p2ptypes "github.com/hyperledger/fabric-droplib/component/p2p/types"
	services4 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/consensus"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"strconv"
	"sync"
	"time"
)

var (
	_ services.IConsensusLogicService = (*consensusServiceImpl)(nil)
)


type consensusServiceImpl struct {
	*streamimpl2.BaseLogicServiceImpl

	consensusConfig *config.ConsensusConfiguration

	state    services.IStateComponent
	eventBus services4.ICommonEventBusComponent

	peerMtx  sync.RWMutex
	peers    map[types2.NodeID]*models.PeerState
	waitSync bool

	stateCh       *models3.Channel
	dataCh        *models3.Channel
	voteCh        *models3.Channel
	voteSetBitsCh *models3.Channel

	// peerUpdates   *models.PeerUpdates
	// 当有节点加入或者dead的时候,会收到event
	// peerUpdatesEvent <-chan interface{}
	// ?
	eventChan <-chan interface{}

	stateCloseCh chan struct{}
	closeCh      chan struct{}
}


func NewConsensusLogicServiceImpl(cfg *config.ConsensusConfiguration, state services.IStateComponent,
	bus services4.ICommonEventBusComponent, ) *consensusServiceImpl {
	r := &consensusServiceImpl{
		consensusConfig: cfg,
		state:           state,
		eventBus:        bus,
		peerMtx:         sync.RWMutex{},
		peers:           make(map[types2.NodeID]*models.PeerState),
		waitSync:        false,
		stateCloseCh:    make(chan struct{}),
		closeCh:         make(chan struct{}),
	}

	r.BaseLogicServiceImpl = streamimpl2.NewBaseLogicServiceImpl(modules.LOGICSERVICE_MODULE_CONSENSUS, r)

	return r
}

func (r *consensusServiceImpl) GetServiceId() types2.ServiceID {
	return LOGICSERVICE_CONSENSUS
}

func (r *consensusServiceImpl) ChoseInterestProtocolsOrTopics(pProtocol *models3.P2PProtocol) error {
	protoChannels := pProtocol.GetProtocolByProtocolID(CONSENSUS_PROTOCOL_ID)
	r.stateCh = protoChannels.GetChannelById(StateChannel)
	r.dataCh = protoChannels.GetChannelById(DataChannel)
	r.voteCh = protoChannels.GetChannelById(VoteChannel)
	r.voteSetBitsCh = protoChannels.GetChannelById(VoteSetBitsChannel)
	return nil
}

func (r *consensusServiceImpl) ChoseInterestEvents(bus base.IP2PEventBusComponent) {
	go func() {
		r.eventChan = bus.AcquireEventChanByNameSpace(types.TOPIC_BROADCAST_ALL)
	}()
}
func (r *consensusServiceImpl) OnStart() error {
	go r.peerStatusRoutine()
	go r.eventRoutine()
	go r.processStateRoutine()
	go r.processDataRoutine()
	go r.processVoteRoutine()
	go r.processVoteSetBitsRoutine()
	// go r.processPeerUpdatesRoutine()

	return nil
}
func (r *consensusServiceImpl) peerStatusRoutine() {
	for {
		if !r.IsRunning() {
			r.Logger.Info("stopping peerStatsRoutine")
			return
		}

		select {
		case msg := <-r.state.StateNotify():
			ps, ok := r.GetPeerState(msg.PeerID)
			if !ok || ps == nil {
				r.Logger.Debug("attempt to update stats for non-existent peer", "peer", msg.PeerID)
				continue
			}
			r.handlePeerStatusMsg(ps, msg)
		case <-r.closeCh:
			return
		}
	}
}

// GetPeerState returns PeerState for a given NodeID.
func (r *consensusServiceImpl) GetPeerState(peerID types2.NodeID) (*models.PeerState, bool) {
	r.peerMtx.RLock()
	defer r.peerMtx.RUnlock()

	ps, ok := r.peers[peerID]
	return ps, ok
}

func (r *consensusServiceImpl) handlePeerStatusMsg(ps *models.PeerState, msg models.MsgInfo) {
	switch msg.Msg.(type) {
	case *models.VoteMessage:
		if numVotes := ps.RecordVote(); numVotes%votesToContributeToBecomeGoodPeer == 0 { // nolint: staticcheck
			// TODO: Handle peer quality via the peer manager.
			// r.Switch.MarkPeerAsGood(peer)
		}
	case *models.BlockPartMessageWrapper:
		// 可以认为是同步当前的区块高度
		if numParts := ps.RecordBlockPart(); numParts%blocksToContributeToBecomeGoodPeer == 0 { // nolint: staticcheck
			// TODO: Handle peer quality via the peer manager.
			// r.Switch.MarkPeerAsGood(peer)
		}
	}
}

func (r *consensusServiceImpl) eventRoutine() {
	r.Logger.Info("开始监听event")
	for {
		if !r.IsRunning() {
			r.Logger.Info("stopping peerStatsRoutine")
			return
		}

		select {
		case eve := <-r.eventChan:
			r.handleEvent(eve)
		}
	}
}

func (this *consensusServiceImpl) handleEvent(eve interface{}) {
	v := eve.(events.EventChanData)

	switch msg := v.Data.(type) {
	case events.EventPeerUpdate:
		this.processPeerUpdate(msg)
	case events.EventRoundState:
		this.stateCh.Out <- models3.ChannelEnvelope{
			Broadcast: types.TOPIC_BROADCAST_ALL,
			Message:   makeRoundStepMessage(msg),
		}
	case events.EventNewValidBlock:
		psHeader := msg.ProposalBlockParts.Header()
		this.stateCh.Out <- models3.ChannelEnvelope{
			Broadcast: types.TOPIC_BROADCAST_ALL,
			Message: &consensus.NewValidBlockProtoWrapper{
				Height:             msg.Height,
				Round:              msg.Round,
				BlockPartSetHeader: psHeader.ToProto(),
				BlockParts:         msg.ProposalBlockParts.BitArray().ToProto(),
				IsCommit:           msg.Step == types.RoundStepCommit,
			},
		}
	case events.EventVote:
		this.stateCh.Out <- models3.ChannelEnvelope{
			Broadcast: types.TOPIC_BROADCAST_ALL,
			Message: &consensus.HasVoteProtoWrapper{
				Height: msg.Height,
				Round:  msg.Round,
				Type:   msg.Type,
				Index:  msg.ValidatorIndex,
			},
		}
	}
}

func (r *consensusServiceImpl) processStateRoutine() {
	defer r.stateCh.Close()

	for {
		select {
		case envelope := <-r.stateCh.In:
			msg, logicMessage, err := r.UnwrapFromBytes(r.stateCh.ID, envelope.Message)
			if nil != err {
				r.Logger.Error("反序列化失败:" + err.Error())
				continue
			}
			if err := r.handleStateMsg(envelope, msg, logicMessage); nil != err {
				r.Logger.Error("处理state channel内部数据失败:chId为:" + strconv.Itoa(int(r.stateCh.ID)) + "err:" + err.Error())
				r.stateCh.ErrC <- errors.PeerError{
					NodeID: envelope.From,
					Err:    err,
				}
			}

		case <-r.stateCloseCh:
			r.Logger.Debug("stopped listening on StateChannel; closing...")
			return
		}
	}
}

func (this *consensusServiceImpl) handleStateMsg(envelope models3.Envelope, msg proto.Message, logicMessage services2.IMessage) error {
	ps, ok := this.GetPeerState(envelope.From)
	if !ok || ps == nil {
		this.Logger.Debug("节点在缓存中不存在,节点信息为:", envelope.From, "ch_id:", "StateChannel")
		return nil
	}
	switch m := msg.(type) {
	case *consensus.NewRoundStepProtoWrapper:
		initHeight := this.state.GetInitialHeight(true)
		if err := logicMessage.(*models.NewRoundStepMessage).ValidateHeight(initHeight); nil != err {
			this.Logger.Error("校验高度失败,高度为:" + strconv.Itoa(int(initHeight)))
			return err
		}
		ps.ApplyNewRoundStepMessage(logicMessage.(*models.NewRoundStepMessage))
	case *consensus.NewValidBlockProtoWrapper:
		ps.ApplyNewValidBlockMessage(logicMessage.(*models.NewValidBlockMessage))
	case *consensus.HasVoteProtoWrapper:
		ps.ApplyHasVoteMessage(logicMessage.(*models.HasVoteMessage))
	case *consensus.VoteSetMaj23ProtoWrapper:
		height, votes := this.state.GetCurrentHeightWithVote(true)
		if int64(height) != m.VoteSetMaj23.Height {
			this.Logger.Info("收到了不匹配的高度信息,")
			return nil
		}

		v23Msg := logicMessage.(*models.VoteSetMaj23Message)
		protoMsg := m.VoteSetMaj23
		err := votes.SetPeerMaj23(protoMsg.Round, protoMsg.Type, ps.NodeID, v23Msg.BlockID)
		if err != nil {
			return err
		}

		var ourVotes *libs.BitArray
		switch v23Msg.Type {
		case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
			ourVotes = votes.Prevotes(protoMsg.Round).BitArrayByBlockID(v23Msg.BlockID)
		case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
			ourVotes = votes.Precommits(protoMsg.Round).BitArrayByBlockID(v23Msg.BlockID)
		default:
			panic("bad VoteSetBitsMessage field type; forgot to add a check in ValidateBasic?")
		}

		respMsg := consensus.VoteSetBitsProtoWrapper{
			Height:  protoMsg.Height,
			Round:   protoMsg.Round,
			Type:    v23Msg.Type,
			BlockId: protoMsg.BlockId,
		}
		if votesProto := ourVotes.ToProto(); votesProto != nil {
			respMsg.Votes = votesProto
		}
		this.voteSetBitsCh.Out <- models3.ChannelEnvelope{
			To:         envelope.From,
			Message:    &respMsg,
			StreamFlag: protos.StreamFlag_CLOSE_IMMEDIATELY,
		}
	}

	return nil
}

func (r *consensusServiceImpl) processDataRoutine() {
	defer r.dataCh.Close()
	for {
		select {
		case envelope := <-r.dataCh.In:
			msg, logicMessage, err := r.UnwrapFromBytes(r.stateCh.ID, envelope.Message)
			if nil != err {
				r.Logger.Error("反序列化失败:" + err.Error())
				continue
			}
			if err := r.handleDataMsg(envelope, msg, logicMessage); err != nil {
				r.Logger.Error("failed to process message", "ch_id", r.dataCh.ID, "envelope", envelope, "err", err)
				r.dataCh.ErrC <- errors.PeerError{
					NodeID: envelope.From,
					Err:    err,
				}
			}

		case <-r.closeCh:
			r.Logger.Debug("stopped listening on DataChannel; closing...")
			return
		}
	}
}

func (this *consensusServiceImpl) handleDataMsg(envelope models3.Envelope, protoMsg proto.Message, msgI services2.IMessage) error {
	logger := this.Logger

	ps, ok := this.GetPeerState(envelope.From)
	if !ok || ps == nil {
		this.Logger.Debug("failed to find peer state")
		return nil
	}

	if this.WaitSync() {
		logger.Info("ignoring message received during sync", "msg", msgI)
		return nil
	}

	switch m := protoMsg.(type) {
	case *consensus.ProposalProtoWrapper:
		concreteMsg := msgI.(*models.ProposalMessage)
		ps.SetHasProposal(concreteMsg.Proposal)
		this.state.PeerMsgNotify() <- models.MsgInfo{Msg: concreteMsg, PeerID: envelope.From}
	case *consensus.ProposalPOLProtoWrapper:
		ps.ApplyProposalPOLMessage(msgI.(*models.ProposalPOLMessage))
	case *consensus.BlockPartProtoWrapper:
		bpMsg := msgI.(*models.BlockPartMessage)
		// 这一步只是标识收到了block而已,真正的处理者还是在go routing
		ps.SetHasProposalBlockPart(bpMsg.Height, bpMsg.Round, int(bpMsg.Part.Index))
		this.state.PeerMsgNotify() <- models.MsgInfo{Msg: bpMsg, PeerID: envelope.From}
		// this.Metrics.BlockParts.With("peer_id", string(envelope.From)).Add(1)
	default:
		return fmt.Errorf("received unknown message on DataChannel: %T", m)
	}

	return nil
}

// WaitSync returns whether the consensus reactor is waiting for state/fast sync.
func (r *consensusServiceImpl) WaitSync() bool {
	r.peerMtx.RLock()
	defer r.peerMtx.RUnlock()

	return r.waitSync
}

func (r *consensusServiceImpl) processVoteRoutine() {
	defer r.voteCh.Close()
	for {
		select {
		case envelope := <-r.voteCh.In:
			msg, logicMessage, err := r.UnwrapFromBytes(r.stateCh.ID, envelope.Message)
			if nil != err {
				r.Logger.Error("反序列化失败:" + err.Error())
				continue
			}
			if err := r.handleVoteMsg(envelope, msg, logicMessage); nil != err {
				r.Logger.Error("处理投票信息失败:" + err.Error())
				continue
			}
		}
	}

}

func (this *consensusServiceImpl) handleVoteMsg(envelope models3.Envelope, protoMsg proto.Message, logicMsg services2.IMessage) error {
	logger := this.Logger

	ps, ok := this.GetPeerState(envelope.From)
	if !ok || ps == nil {
		this.Logger.Debug("failed to find peer state")
		return nil
	}

	if this.WaitSync() {
		logger.Info("ignoring message received during sync", "msg", protoMsg)
		return nil
	}

	switch msg := protoMsg.(type) {
	case *consensus.VoteProtoWrapper:
		height, size, lastCommitSize := this.state.GetSizeWithHeightWithLastCommitSize(true)
		concreteMsg := logicMsg.(*models.VoteMessage)
		ps.EnsureVoteBitArrays(height, size)
		ps.EnsureVoteBitArrays(height-1, lastCommitSize)
		// 更新哦投票信息
		ps.SetHasVote(concreteMsg.Vote)
		// 如果满足了2/3 则会通过msgQueue实现通知
		this.state.PeerMsgNotify() <- models.MsgInfo{Msg: concreteMsg, PeerID: envelope.From}
	default:
		return fmt.Errorf("received unknown message on VoteChannel: %T", msg)
	}

	return nil
}

func (r *consensusServiceImpl) processVoteSetBitsRoutine() {
	defer r.voteSetBitsCh.Close()
	for {
		select {
		case envelope := <-r.dataCh.In:
			msg, logicMessage, err := r.UnwrapFromBytes(r.stateCh.ID, envelope.Message)
			if nil != err {
				r.Logger.Error("反序列化失败:" + err.Error())
				continue
			}
			if err := r.handleVoteSetBitMsg(envelope, msg, logicMessage); err != nil {
				r.Logger.Error("failed to process message", "ch_id", r.dataCh.ID, "envelope", envelope, "err", err)
				r.voteSetBitsCh.ErrC <- errors.PeerError{
					NodeID: envelope.From,
					Err:    err,
				}
			}

		case <-r.closeCh:
			r.Logger.Debug("stopped listening on DataChannel; closing...")
			return
		}
	}
}

func (this *consensusServiceImpl) handleVoteSetBitMsg(envelope models3.Envelope, protoMsg proto.Message, logicMsg services2.IMessage) error {
	logger := this.Logger

	ps, ok := this.GetPeerState(envelope.From)
	if !ok || ps == nil {
		this.Logger.Debug("failed to find peer state")
		return nil
	}

	if this.WaitSync() {
		logger.Info("ignoring message received during sync", "msg", logicMsg)
		return nil
	}

	switch msg := protoMsg.(type) {
	case *consensus.VoteSetBitsProtoWrapper:
		height, votes := this.state.GetCurrentHeightWithVote(true)
		concreteMsg := logicMsg.(*models.VoteSetBitsMessage)
		if int64(height) == msg.Height {
			var outVotes *libs.BitArray
			switch msg.Type {
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE:
				// prevote 投票
				outVotes = votes.Prevotes(msg.Round).BitArrayByBlockID(concreteMsg.BlockID)
			case types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT:
				outVotes = votes.Precommits(msg.Round).BitArrayByBlockID(concreteMsg.BlockID)
			default:
				panic("未知的msg类型")
			}
			ps.ApplyVoteSetBitsMessage(concreteMsg, outVotes)
		} else {
			ps.ApplyVoteSetBitsMessage(concreteMsg, nil)
		}
	default:
		return fmt.Errorf("received unknown message on VoteSetBitsChannel: %T", msg)
	}

	return nil
}

func (r *consensusServiceImpl) processPeerUpdatesRoutine() {
	// logger := r.Logger
	// for {
	// 	select {
	// 	case e := <-r.peerUpdatesEvent:
	// 		if v, ok := e.(events.EventPeerUpdate); ok {
	// 			r.processPeerUpdate(v)
	// 		} else {
	// 			logger.Error(fmt.Sprintf("获取到event的类型不是 EventPeerUpdate,实际类型为:%T", e))
	// 		}
	//
	// 	}
	// }

}

func (r *consensusServiceImpl) processPeerUpdate(v events.EventPeerUpdate) {
	r.Logger.Info("收到了新的节点变更信息:" + v.String())

	r.peerMtx.Lock()
	defer r.peerMtx.Unlock()

	switch v.Type {
	case p2ptypes.MEMBER_CHANGED_ADD:
		if !r.IsRunning() {
			return
		}

		var (
			ps *models.PeerState
			ok bool
		)

		ps, ok = r.peers[v.PeerId]
		if !ok {
			ps = models.NewPeerState(r.Logger, v.PeerId)
			r.peers[v.PeerId] = ps
		}

		if !ps.IsRunning() {
			// Set the peer state's closer to signal to all spawned goroutines to exit
			// when the peer is removed. We also set the running state to ensure we
			// do not spawn multiple instances of the same goroutines and finally we
			// set the waitgroup counter so we know when all goroutines have exited.
			ps.BroadcastWG.Add(3)
			ps.SetRunning(true)

			// start goroutines for this peer
			go r.gossipDataRoutine(ps)
			go r.gossipVotesRoutine(ps)
			go r.queryMaj23Routine(ps)

			// Send our state to the peer. If we're fast-syncing, broadcast a
			// RoundStepMessage later upon SwitchToConsensus().
			if !r.waitSync {
				go r.sendNewRoundStepMessage(ps.NodeID)
			}
		}
	}
}

func (r *consensusServiceImpl) gossipDataRoutine(ps *models.PeerState) {
	logger := r.Logger

	defer ps.BroadcastWG.Done()
OUTER_LOOP:
	for {
		if !r.IsRunning() {
			return
		}

		select {
		case <-ps.Closer.Done():
			// The peer is marked for removal via a PeerUpdate as the doneCh was
			// explicitly closed to signal we should exit.
			return

		default:
		}

		// 获取当前节点的roundState
		rs := r.state.GetRoundState()
		// 获取其他节点的roundState
		prs := ps.GetRoundState()

		// Send proposal Block parts?
		// 如果在本节点中有这个节点的 proposal的部分头部,说明可能没有同步完整个区块
		if rs.ProposalBlockParts.HasHeader(prs.ProposalBlockPartSetHeader) {
			//  因此同步区块
			if index, ok := rs.ProposalBlockParts.BitArray().Sub(prs.ProposalBlockParts.Copy()).PickRandom(); ok {
				part := rs.ProposalBlockParts.GetPart(index)
				partProto, err := part.ToProto()
				if err != nil {
					logger.Error("failed to convert block part to proto", "err", err)
					return
				}

				logger.Debug("sending block part", "height", prs.Height, "round", prs.Round)
				r.dataCh.Out <- models3.ChannelEnvelope{
					To: ps.NodeID.ToNodeID(),
					Message: &consensus.BlockPartProtoWrapper{
						Height: int64(rs.Height),
						Round:  rs.Round,
						Part:   partProto,
					},
				}

				ps.SetHasProposalBlockPart(prs.Height, prs.Round, index)
				continue OUTER_LOOP
			}
		}
		//
		// if the peer is on a previous height that we have, help catch up
		blockStoreBase := r.state.GetBlockStore().Base()
		// 如果远端节点的高度小于本节点的state中的高度 并且远端节点的高度又是大于存储着的节点高度
		if blockStoreBase > 0 && 0 < prs.Height && prs.Height < rs.Height && uint64(prs.Height) >= blockStoreBase {
			// FIXME
			// heightLogger := logger.With("height", prs.Height)
			heightLogger := r.Logger

			// If we never received the commit message from the peer, the block parts
			// will not be initialized.
			if prs.ProposalBlockParts == nil {
				blockMeta := r.state.GetBlockStore().LoadBlockMeta(uint64(prs.Height))
				if blockMeta == nil {
					heightLogger.Error(
						"failed to load block meta",
						"blockstoreBase", blockStoreBase,
						"blockstoreHeight", r.state.GetBlockStore().Height(),
					)
					time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
				} else {
					// 		// FIXME ,既然只是一个拷贝的话,为什么还要初始化
					ps.InitProposalBlockParts(blockMeta.BlockID.PartSetHeader)
				}
				//
				// Continue the loop since prs is a copy and not effected by this
				// initialization.
				continue OUTER_LOOP
			}

			r.gossipDataForCatchup(rs, prs, ps)
			continue OUTER_LOOP
		}

		// if height and round don't match, sleep
		if (rs.Height != prs.Height) || (rs.Round != prs.Round) {
			time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
			continue OUTER_LOOP
		}

		// By here, height and round match.
		// Proposal block parts were already matched and sent if any were wanted.
		// (These can match on hash so the round doesn't matter)
		// Now consider sending other things, like the Proposal itself.

		// Send Proposal && ProposalPOL BitArray?

		if rs.Proposal != nil && !prs.Proposal {
			// Proposal: share the proposal metadata with peer.
			{
				propProto := rs.Proposal.ToProto()

				logger.Debug("sending proposal", "height", prs.Height, "round", prs.Round)
				r.dataCh.Out <- models3.ChannelEnvelope{
					To: ps.NodeID.ToNodeID(),
					Message: &consensus.ProposalProtoWrapper{
						Proposal: propProto,
					},
				}

				// NOTE: A peer might have received a different proposal message, so
				// this Proposal msg will be rejected!
				ps.SetHasProposal(rs.Proposal)
			}

			// ProposalPOL: lets peer know which POL votes we have so far. The peer
			// must receive ProposalMessage first. Note, rs.Proposal was validated,
			// so rs.Proposal.POLRound <= rs.Round, so we definitely have
			// rs.Votes.Prevotes(rs.Proposal.POLRound).
			if 0 <= rs.Proposal.POLRound {
				pPol := rs.Votes.Prevotes(rs.Proposal.POLRound).BitArray()
				pPolProto := pPol.ToProto()

				logger.Debug("sending POL", "height", prs.Height, "round", prs.Round)
				r.dataCh.Out <- models3.ChannelEnvelope{
					To: ps.NodeID.ToNodeID(),
					Message: &consensus.ProposalPOLProtoWrapper{
						Height:           int64(rs.Height),
						ProposalPolRound: rs.Round,
						ProposalPol:      pPolProto,
					},
				}
			}

			continue OUTER_LOOP
		}

		// nothing to do -- sleep
		time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
		continue OUTER_LOOP
	}
}

func (r *consensusServiceImpl) gossipDataForCatchup(rs *models.RoundState, prs *models.PeerRoundState, ps *models.PeerState) {
	logger := r.Logger
	if index, ok := prs.ProposalBlockParts.Not().PickRandom(); ok {
		// ensure that the peer's PartSetHeader is correct
		blockMeta := r.state.GetBlockStore().LoadBlockMeta(uint64(prs.Height))
		// 说明区块数据被篡改了,或者是因为上层传入的参数并没有严格的校验
		if blockMeta == nil {
			logger.Error(
				"failed to load block meta",
				"our_height", rs.Height,
				"blockstore_base", r.state.GetBlockStore().Base(),
				"blockstore_height", r.state.GetBlockStore().Height(),
			)

			time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
			return
		} else if !blockMeta.BlockID.PartSetHeader.Equals(prs.ProposalBlockPartSetHeader) {
			logger.Info(
				"peer ProposalBlockPartSetHeader mismatch; sleeping",
				"block_part_set_header", blockMeta.BlockID.PartSetHeader,
				"peer_block_part_set_header", prs.ProposalBlockPartSetHeader,
			)

			time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
			return
		}
		// 同上
		part := r.state.GetBlockStore().LoadBlockPart(uint64(prs.Height), index)
		if part == nil {
			logger.Error(
				"failed to load block part",
				"index", index,
				"block_part_set_header", blockMeta.BlockID.PartSetHeader,
				"peer_block_part_set_header", prs.ProposalBlockPartSetHeader,
			)

			time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
			return
		}

		partProto, err := part.ToProto()
		if err != nil {
			logger.Error("failed to convert block part to proto", "err", err)

			time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
			return
		}

		logger.Debug("sending block part for catchup", "round", prs.Round, "index", index)
		r.dataCh.Out <- models3.ChannelEnvelope{
			To: ps.NodeID.ToNodeID(),
			Message: &consensus.BlockPartProtoWrapper{
				Height: int64(prs.Height),
				Round:  prs.Round,
				Part:   partProto,
			},
		}

		return
	}

	time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
}

func (r *consensusServiceImpl) gossipVotesRoutine(ps *models.PeerState) {
	logger := r.Logger
	defer ps.BroadcastWG.Done()

	// XXX: simple hack to throttle logs upon sleep
	logThrottle := 0
OUTER_LOOP:
	for {
		if !r.IsRunning() {
			return
		}

		select {
		case <-ps.Closer.Done():
			// The peer is marked for removal via a PeerUpdate as the doneCh was
			// explicitly closed to signal we should exit.
			return

		default:
		}

		rs := r.state.GetRoundState()
		prs := ps.GetRoundState()

		switch logThrottle {
		case 1: // first sleep
			logThrottle = 2
		case 2: // no more sleep
			logThrottle = 0
		}

		// if height matches, then send LastCommit, Prevotes, and Precommits

		if rs.Height == prs.Height {
			if r.gossipVotesForHeight(rs, prs, ps) {
				continue OUTER_LOOP
			}
		}

		// special catchup logic -- if peer is lagging by height 1, send LastCommit
		if prs.Height != 0 && rs.Height == prs.Height+1 {
			if r.pickSendVote(ps, rs.LastCommit) {
				logger.Debug("picked rs.LastCommit to send", "height", prs.Height)
				continue OUTER_LOOP
			}
		}

		// catchup logic -- if peer is lagging by more than 1, send Commit
		blockStoreBase := r.state.GetBlockStore().Base()
		// 说明远端节点落后很多块,同时为了确保数据的有效性(既数据从磁盘中获取),所以又多了一步比本地的磁盘高度大
		if blockStoreBase > 0 && prs.Height != 0 && rs.Height >= prs.Height+2 && uint64(prs.Height) >= blockStoreBase {
			// Load the block commit for prs.Height, which contains precommit
			// signatures for prs.Height.
			if commit := r.state.GetBlockStore().LoadBlockCommit(uint64(prs.Height)); commit != nil {
				if r.pickSendVote(ps, commit) {
					logger.Debug("picked Catchup commit to send", "height", prs.Height)
					continue OUTER_LOOP
				}
			}
		}

		if logThrottle == 0 {
			// we sent nothing -- sleep
			logThrottle = 1
			logger.Debug(
				"no votes to send; sleeping",
				"rs.Height", rs.Height,
				"prs.Height", prs.Height,
				"localPV", rs.Votes.Prevotes(rs.Round).BitArray(), "peerPV", prs.Prevotes,
				"localPC", rs.Votes.Precommits(rs.Round).BitArray(), "peerPC", prs.Precommits,
			)
		} else if logThrottle == 2 {
			logThrottle = 1
		}

		time.Sleep(r.consensusConfig.PeerGossipSleepDuration)
		continue OUTER_LOOP
	}

}

func (r *consensusServiceImpl) gossipVotesForHeight(rs *models.RoundState, prs *models.PeerRoundState, ps *models.PeerState) bool {
	// logger := r.Logger.With("height", prs.Height).With("peer", ps.peerID)
	logger := r.Logger

	// if there are lastCommits to send..
	//  可能远端节点因为网络延误,导致消息还未发出
	if prs.Step == types.RoundStepNewHeight {
		// 如果远端节点,我们这个节点还未发送出去,则发送,并且推出
		if r.pickSendVote(ps, rs.LastCommit) {
			logger.Debug("picked rs.LastCommit to send")
			return true
		}
	}

	// if there are POL prevotes to send...
	// 说明远端节点处于lock 状态
	if prs.Step <= types.RoundStepPropose && prs.Round != -1 && prs.Round <= rs.Round && prs.ProposalPOLRound != -1 {
		// 本节点获取对于该pol 的round是否已经有了投票
		if polPrevotes := rs.Votes.Prevotes(prs.ProposalPOLRound); polPrevotes != nil {
			// 如果之前没有投票的话,对这个节点发送自己的投票信息
			if r.pickSendVote(ps, polPrevotes) {
				logger.Debug("picked rs.Prevotes(prs.ProposalPOLRound) to send", "round", prs.ProposalPOLRound)
				return true
			}
		}
	}

	// if there are prevotes to send...
	// 则当前的远端节点处于provote阶段,当发送完了prevote之后就会进入prevoteWait等待阶段
	if prs.Step <= types.RoundStepPrevoteWait && prs.Round != -1 && prs.Round <= rs.Round {
		if r.pickSendVote(ps, rs.Votes.Prevotes(prs.Round)) {
			logger.Debug("picked rs.Prevotes(prs.Round) to send", "round", prs.Round)
			return true
		}
	}

	// if there are precommits to send...
	if prs.Step <= types.RoundStepPrecommitWait && prs.Round != -1 && prs.Round <= rs.Round {
		if r.pickSendVote(ps, rs.Votes.Precommits(prs.Round)) {
			logger.Debug("picked rs.Precommits(prs.Round) to send", "round", prs.Round)
			return true
		}
	}

	// if there are prevotes to send...(which are needed because of validBlock mechanism)
	if prs.Round != -1 && prs.Round <= rs.Round {
		if r.pickSendVote(ps, rs.Votes.Prevotes(prs.Round)) {
			logger.Debug("picked rs.Prevotes(prs.Round) to send", "round", prs.Round)
			return true
		}
	}

	// if there are POLPrevotes to send...
	if prs.ProposalPOLRound != -1 {
		if polPrevotes := rs.Votes.Prevotes(prs.ProposalPOLRound); polPrevotes != nil {
			if r.pickSendVote(ps, polPrevotes) {
				logger.Debug("picked rs.Prevotes(prs.ProposalPOLRound) to send", "round", prs.ProposalPOLRound)
				return true
			}
		}
	}

	return false
}

func (r *consensusServiceImpl) pickSendVote(ps *models.PeerState, votes models.VoteSetReader) bool {
	if vote, ok := ps.PickVoteToSend(votes); ok {
		r.Logger.Debug("sending vote message", "ps", ps, "vote", vote)
		r.voteCh.Out <- models3.ChannelEnvelope{
			To: ps.NodeID.ToNodeID(),
			Message: &consensus.VoteProtoWrapper{
				Vote: vote.ToProto(),
			},
		}
		ps.SetHasVote(vote)
		return true
	}

	return false
}

func (r *consensusServiceImpl) queryMaj23Routine(ps *models.PeerState) {
	defer ps.BroadcastWG.Done()

OUTER_LOOP:
	for {
		if !r.IsRunning() {
			return
		}

		select {
		case <-ps.Closer.Done():
			// The peer is marked for removal via a PeerUpdate as the doneCh was
			// explicitly closed to signal we should exit.
			return

		default:
		}

		// maybe send Height/Round/Prevotes
		{
			rs := r.state.GetRoundState()
			prs := ps.GetRoundState()

			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Prevotes(prs.Round).TwoThirdsMajority(); ok {
					r.stateCh.Out <- models3.ChannelEnvelope{
						To: ps.NodeID.ToNodeID(),
						Message: &consensus.VoteSetMaj23ProtoWrapper{
							VoteSetMaj23: &types3.VoteSetMaj23Proto{
								Height:  int64(prs.Height),
								Round:   prs.Round,
								Type:    types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE,
								BlockId: maj23.ToProto(),
							},
						},
					}
					time.Sleep(r.consensusConfig.PeerQueryMaj23SleepDuration)
				}
			}
		}

		// maybe send Height/Round/Precommits
		{
			rs := r.state.GetRoundState()
			prs := ps.GetRoundState()

			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Precommits(prs.Round).TwoThirdsMajority(); ok {
					r.stateCh.Out <- models3.ChannelEnvelope{
						To: ps.NodeID.ToNodeID(),
						Message: &consensus.VoteSetMaj23ProtoWrapper{
							VoteSetMaj23: &types3.VoteSetMaj23Proto{
								Height:  int64(prs.Height),
								Round:   prs.Round,
								Type:    types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT,
								BlockId: maj23.ToProto(),
							},
						},
					}

					time.Sleep(r.consensusConfig.PeerQueryMaj23SleepDuration)
				}
			}
		}

		// maybe send Height/Round/ProposalPOL
		{
			rs := r.state.GetRoundState()
			prs := ps.GetRoundState()

			// 可能那会高度并不一致,因为全是异步,所以此处再判断一次
			if rs.Height == prs.Height && prs.ProposalPOLRound >= 0 {
				if maj23, ok := rs.Votes.Prevotes(prs.ProposalPOLRound).TwoThirdsMajority(); ok {
					r.stateCh.Out <- models3.ChannelEnvelope{
						To: ps.NodeID.ToNodeID(),
						Message: &consensus.VoteSetMaj23ProtoWrapper{
							VoteSetMaj23: &types3.VoteSetMaj23Proto{
								Height:  int64(prs.Height),
								Round:   prs.Round,
								Type:    types3.SignedMsgType_SIGNED_MSG_TYPE_PREVOTE,
								BlockId: maj23.ToProto(),
							},
						},
					}
					time.Sleep(r.consensusConfig.PeerQueryMaj23SleepDuration)
				}
			}
		}

		// Little point sending LastCommitRound/LastCommit, these are fleeting and
		// non-blocking.

		// maybe send Height/CatchupCommitRound/CatchupCommit
		{
			prs := ps.GetRoundState()
			// 说明远端节点,高度不一致,但是也并不处于锁块状态,
			// FIXME 则可能处于???
			// FIXME 为什么不从本地的height判断
			// 本地有刚提交过commit,并且远端节点的高度小于本地文件数据库的高度,又大于初次加载时候的区块(说明中间有交易发生,但是
			// 可能因为某种原因导致远端节点没有commit)
			blockStore := r.state.GetBlockStore()
			if prs.CatchupCommitRound != -1 && prs.Height > 0 && uint64(prs.Height) <= blockStore.Height() &&
				uint64(prs.Height) >= blockStore.Base() {
				if commit := r.state.LoadCommit(prs.Height); commit != nil {
					r.stateCh.Out <- models3.ChannelEnvelope{
						To: ps.NodeID.ToNodeID(),
						Message: &consensus.VoteSetMaj23ProtoWrapper{
							VoteSetMaj23: &types3.VoteSetMaj23Proto{
								Height:  int64(prs.Height),
								Round:   commit.Round,
								Type:    types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT,
								BlockId: commit.BlockID.ToProto(),
							},
						},
					}
					time.Sleep(r.consensusConfig.PeerQueryMaj23SleepDuration)
				}
			}
		}

		time.Sleep(r.consensusConfig.PeerQueryMaj23SleepDuration)
		continue OUTER_LOOP
	}
}

func (r *consensusServiceImpl) sendNewRoundStepMessage(id types2.PeerIDPretty) {
	rs := r.state.GetRoundState()
	msg := &consensus.NewRoundStepProtoWrapper{
		NewRoundStep: &types3.NewRoundStepProto{
			Height:                int64(rs.Height),
			Round:                 rs.Round,
			Step:                  uint32(rs.Step),
			SecondsSinceStartTime: int64(time.Since(rs.StartTime).Seconds()),
			LastCommitRound:       rs.LastCommit.GetRound(),
		},
	}
	r.stateCh.Out <- models3.ChannelEnvelope{
		To:      id.ToNodeID(),
		Message: msg,
	}
}

func makeRoundStepMessage(rs events.EventRoundState) *consensus.NewRoundStepProtoWrapper {
	return &consensus.NewRoundStepProtoWrapper{
		NewRoundStep: &types3.NewRoundStepProto{
			Height:                rs.Height,
			Round:                 rs.Round,
			Step:                  uint32(rs.Step),
			SecondsSinceStartTime: int64(time.Since(rs.StartTime).Seconds()),
			LastCommitRound:       rs.LastCommit.GetRound(),
		},
	}
}
