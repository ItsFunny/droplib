/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 2:01 下午
# @File : common_test.go
# @Description :
# @Attention :
*/
package state

import (
	"context"
	"encoding/base64"
	"fmt"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services3 "github.com/hyperledger/fabric-droplib/base/services"
	types2 "github.com/hyperledger/fabric-droplib/base/types"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/block"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/event"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/evidence"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/mempool"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/mock"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/statestore"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/component/event/impl"
	models2 "github.com/hyperledger/fabric-droplib/component/pubsub/models"
	services2 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	rand "github.com/hyperledger/fabric-droplib/libs/random"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sort"
	"testing"
	"time"
)

var (
	ensureTimeout = time.Millisecond * 500000
)
var (
	demoChain = "demo_chain"
)

var (
	globalCfg     *config.GlobalConfiguration
	UnknownPeerID uint16 = 0
)

const (
	testSubscriber = "test-client"
)

func init() {
	globalCfg = config.NewDefaultGlobalConfiguration()
}

func StartTx(t *testing.T, mp services.IMemPool, notifyC chan int) {
	go routineAppendTx(t, mp, notifyC)
}
func routineAppendTx(t *testing.T, mp services.IMemPool, countC <-chan int) {
	for {
		select {
		case c := <-countC:
			appendTx(t, c, mp, UnknownPeerID)
		}
	}
}
func appendTx(t *testing.T, count int, mp services.IMemPool, peerID uint16) types.Txs {
	txs := make(types.Txs, count)
	// txInfo := types.TxInfo{SenderID: peerID}
	for i := 0; i < count; i++ {
		txBytes := rand.Bytes(100)
		txs[i] = txBytes
	}
	info := models.TxInfo{
		SenderID:    uint16(count),
		SenderP2PID: types2.NodeID(rand.CRandHex(3)),
	}
	err := mp.CheckAndAppendTx(txs, info)
	if nil != err {
		panic(err)
	}
	return txs
}

func Subscribe(eventBus services2.ICommonEventBusComponent, q services2.Query) <-chan models2.PubSubMessage {
	sub, err := eventBus.Subscribe(context.Background(), testSubscriber, q)
	if err != nil {
		panic(fmt.Sprintf("failed to subscribe %s to %v", testSubscriber, q))
	}
	return sub.Out()
}
func RandFabricState(t *testing.T, count int) (*State, []*mock.MockValidator, services.IMemPool, services2.ICommonEventBusComponent) {
	state, privVals := RandGenesisState(count, false, 10)

	validators := make([]*mock.MockValidator, count)
	cs, mempool, bus := newState(state, privVals[0])
	for i := 0; i < count; i++ {
		validators[i] = mock.NewMockValidator(privVals[i], int32(i))
		validators[i].GetPublicKey().Address()
	}
	if err := mempool.Start(services3.ASYNC_START); nil != err {
		panic(err)
	}
	// since cs1 starts at 1
	// 这里不需要增加,因为fabric 从0开始
	// incrementHeight(validators[1:]...)
	return cs, validators, mempool, bus
}

func newState(state models.LatestState, pv cryptolibs.ISigner) (*State, services.IMemPool, services2.ICommonEventBusComponent) {
	// config := cfg.ResetTestRoot("consensus_state_test")
	return newStateWithConfig(globalCfg, state, pv)
}
func newStateWithConfig(
	thisConfig *config.GlobalConfiguration,
	state models.LatestState,
	pv cryptolibs.ISigner,
) (*State, services.IMemPool, services2.ICommonEventBusComponent) {
	// blockDB := dbm.NewMemDB()
	return newStateWithConfigAndBlockStore(thisConfig, state, pv)
}
func newStateWithConfigAndBlockStore(cfg *config.GlobalConfiguration, latestState models.LatestState, pv cryptolibs.ISigner, ) (*State, services.IMemPool, services2.ICommonEventBusComponent) {
	db, err := block.NewGoLevelDB("test", cfg.BaseConfiguration.RootDir)
	if nil != err {
		panic(err)
	}
	// bus
	bus := event.NewDefaultEventBus()
	if err := bus.Start(services3.ASYNC_START); nil != err {
		panic(err)
	}

	// event_
	eventComponent := impl.NewDefaultEventComponent()

	// blockStore
	blockStore := block.NewBlockStoreImpl(db)

	//  Mempool
	mempool := mempool.NewDefaultMemPool(0, cfg.MemPoolConfiguration, bus)
	if cfg.ConsensusConfiguration.WaitForTxs() {
		mempool.EnableTxsAvailable()
	}

	// stateStore
	stateStore := statestore.NewDefaultStateStore(db)
	// evidence_pool
	evidencePool := evidence.NewEmptyEvidencePool()

	// block_executor
	blockExecutor := block.NewDefaultBlockExecutor(mempool, evidencePool, stateStore, bus)

	//  state
	cs := NewState(cfg.ConsensusConfiguration,
		latestState,
		blockExecutor,
		pv,
		blockStore,
		mempool,
		evidencePool,
		eventComponent,
		bus)
	cs.SetLogger(logrusplugin.NewLogrusLogger(modules.MODULE_STATE))
	cs.SetPrivValidator(pv)

	return cs, mempool, bus
}
func RandGenesisState(numValidators int, randPower bool, minPower int64) (models.LatestState, []cryptolibs.ISigner) {
	genDoc, privValidators := randGenesisDoc(numValidators, randPower, minPower)
	s0 := models.MakeGenesisState(genDoc)
	return s0, privValidators
}

func randGenesisDoc(numValidators int, randPower bool, minPower int64) (*models.GenesisDoc, []cryptolibs.ISigner) {
	validators := make([]models.GenesisValidator, numValidators)
	privValidators := make([]cryptolibs.ISigner, numValidators)
	for i := 0; i < numValidators; i++ {
		val, privVal := models.RandValidator(randPower, minPower)
		validators[i] = models.GenesisValidator{
			PubKey: models.Base64PubKeyString(base64.StdEncoding.EncodeToString(val.PubKey.ToBytes())),
			Power:  val.VotingPower,
		}
		privValidators[i] = privVal
	}
	sort.Sort(cryptolibs.PrivValidatorsByAddress(privValidators))
	params := config.ConsensusParamsFromConfig(config.NewDefaultGlobalConfiguration())
	return &models.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         demoChain,
		InitialHeight:   0,
		ConsensusParams: params,
		Validators:      validators,
	}, privValidators
}

func EnsureNewRound(roundCh <-chan models2.PubSubMessage, height uint64, round int32) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewRound event")
	case msg := <-roundCh:
		newRoundEvent, ok := msg.Data().(events.EventDataNewRound)
		if !ok {
			panic(fmt.Sprintf("expected a EventDataNewRound, got %T. Wrong subscription channel?",
				msg.Data()))
		}
		if newRoundEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, newRoundEvent.Height))
		}
		if newRoundEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, newRoundEvent.Round))
		}
	}
}

func EnsureNewProposal(proposalCh <-chan models2.PubSubMessage, height int64, round int32) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewProposal event")
	case msg := <-proposalCh:
		proposalEvent, ok := msg.Data().(events.EventDataCompleteProposal)
		if !ok {
			panic(fmt.Sprintf("expected a EventDataCompleteProposal, got %T. Wrong subscription channel?",
				msg.Data()))
		}
		if int64(proposalEvent.Height) != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, proposalEvent.Height))
		}
		if proposalEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, proposalEvent.Round))
		}
	}
}

func signAddVotes(to *State, voteType types3.SignedMsgType, hash []byte, header models.PartSetHeader, vss ...*mock.MockValidator, ) {
	votes := signVotes(voteType, hash, header, vss...)
	// 伪造2/3 投票
	addVotes(to, votes...)
}
func signVotes(voteType types3.SignedMsgType, hash []byte, header models.PartSetHeader, vss ...*mock.MockValidator) []*models.Vote {
	votes := make([]*models.Vote, len(vss))
	for i, vs := range vss {
		votes[i] = signVote(vs, voteType, hash, header)
	}
	return votes
}

func signVote(vs *mock.MockValidator, voteType types3.SignedMsgType, hash []byte, header models.PartSetHeader) *models.Vote {
	v, err := vs.MockSignVote(globalCfg, voteType, hash, header)
	if err != nil {
		panic(fmt.Errorf("failed to sign vote: %v", err))
	}
	return v
}

func addVotes(to *State, votes ...*models.Vote) {
	for _, vote := range votes {
		to.peerMsgQueue <- models.MsgInfo{Msg: &models.VoteMessage{vote}}
	}
}

func ensureNewTimeout(timeoutCh <-chan models2.PubSubMessage, height uint64, round int32, timeout int64) {
	timeoutDuration := time.Duration(timeout*10) * time.Nanosecond
	ensureNewEvent(timeoutCh, height, round, timeoutDuration,
		"Timeout expired while waiting for NewTimeout event")
}

func ensureNewEvent(ch <-chan models2.PubSubMessage, height uint64, round int32, timeout time.Duration, errorMessage string) {
	select {
	case <-time.After(timeout):
		panic(errorMessage)
	case msg := <-ch:
		roundStateEvent, ok := msg.Data().(events.EventDataRoundState)
		if !ok {
			panic(fmt.Sprintf("expected a EventDataRoundState, got %T. Wrong subscription channel?",
				msg.Data()))
		}
		if roundStateEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, roundStateEvent.Height))
		}
		if roundStateEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, roundStateEvent.Round))
		}
		// TODO: We could check also for a step at this point!
	}
}
