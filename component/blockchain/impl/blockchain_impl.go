/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 10:32 上午
# @File : blockchain_impl.go
# @Description :
# @Attention :
*/
package impl

import (
	"context"
	"errors"
	"fmt"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	services3 "github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/types"
	models2 "github.com/hyperledger/fabric-droplib/common/models"
	"github.com/hyperledger/fabric-droplib/component/base"
	models3 "github.com/hyperledger/fabric-droplib/component/blockchain/models"
	"github.com/hyperledger/fabric-droplib/component/blockchain/services"
	"github.com/hyperledger/fabric-droplib/component/common/blockchain/config"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/block"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/consensus"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/event"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/evidence"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/mempool"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/state"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/statestore"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	services2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	impl2 "github.com/hyperledger/fabric-droplib/component/event/impl"
	streamimpl "github.com/hyperledger/fabric-droplib/component/p2p/impl"
	"github.com/hyperledger/fabric-droplib/component/pubsub/impl"
	services4 "github.com/hyperledger/fabric-droplib/component/pubsub/services"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	"github.com/hyperledger/fabric-droplib/libs/json"
)

var (
	_ services.IBlockChainLogicService = (*blockChainLogicServiceImpl)(nil)
)

type blockChainLogicServiceImpl struct {
	*streamimpl.BaseLogicServiceImpl

	consensusLogicService services2.IConsensusLogicService

	// memPool
	memPool services2.IMemPoolLogicService

	bus services4.ICommonEventBusComponent
}

func (this *blockChainLogicServiceImpl) Propose(ctx context.Context, data []byte) error {
	panic("implement me")
	// 通知tx ,收到了数据 ,让其进行共识
}

func (this *blockChainLogicServiceImpl) Commit() chan<- models3.Entry {
	panic("implement me")
}

func NewBlockChainLogicServiceImpl(cfg *config.GlobalConfiguration, signer cryptolibs.ISigner) *blockChainLogicServiceImpl {
	r := &blockChainLogicServiceImpl{}
	r.BaseLogicServiceImpl = streamimpl.NewBaseLogicServiceImpl(modules.LOGICSERVICE_MODULE_BLOCKCHAIN, r)
	bus := impl.NewCommonEventBusComponentImpl()
	r.bus = bus
	stateBus := event.NewDefaultEventBusWithComponent(bus)
	eventCompoennt := impl2.NewDefaultEventComponent()
	//  Mempool
	// FIXME ,MemPool 需要有一个对应的reactor
	// 2021-04-19 add : 不需要
	mp := mempool.NewDefaultMemPool(0, cfg.MemPoolConfiguration, bus)
	r.memPool = mempool.NewDefaultMemPoolLoigcService(mp)

	stComponent, _ := buildState(cfg, signer, mp, stateBus, eventCompoennt)
	r.consensusLogicService = consensus.NewConsensusLogicServiceImpl(cfg.ConsensusConfiguration, stComponent, r.bus)
	return r
}
func buildState(cfg *config.GlobalConfiguration, pv cryptolibs.ISigner, mempool services2.IMemPool, bus services2.IConsensusEventBusComponent, eventComponent base.IEventComponent) (services2.IStateComponent, *models.GenesisDoc) {
	blockStore, stateDb, err := initDB(cfg, DefaultDBProvider)
	if nil != err {
		panic(err)
	}

	latestState, doc, err := LoadStateFromDBOrGenesisDocProvider(stateDb, DefaultGenesisDocProviderFunc(cfg))
	if nil != err {
		panic(err)
	}

	// stateStore
	stateStore := statestore.NewDefaultStateStore(stateDb)
	// evidence_pool
	evidencePool := evidence.NewEmptyEvidencePool()

	// block_executor
	blockExecutor := block.NewDefaultBlockExecutor(mempool, evidencePool, stateStore, bus)

	cs := state.NewState(cfg.ConsensusConfiguration,
		latestState,
		blockExecutor,
		pv,
		blockStore,
		mempool,
		evidencePool,
		eventComponent, bus)
	cs.SetLogger(logrusplugin.NewLogrusLogger(modules.MODULE_STATE))
	cs.SetPrivValidator(pv)

	return cs, doc
}

type GenesisDocProvider func() (*models.GenesisDoc, error)

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *config.GlobalConfiguration
}

func DefaultDBProvider(ctx *DBContext) (services2.DB, error) {
	return block.NewGoLevelDB(ctx.ID, ctx.Config.DBDir())
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (services2.DB, error)

func initDB(config *config.GlobalConfiguration, dbProvider DBProvider) (services2.IBlockStore, services2.DB, error) {
	var blockStoreDB services2.DB
	blockStoreDB, err := dbProvider(&DBContext{"blockstore", config})
	if err != nil {
		return nil, nil, err
	}

	// blockStore = store.NewBlockStore(blockStoreDB)
	blockStore := block.NewBlockStoreImpl(blockStoreDB)

	stateDB, err := dbProvider(&DBContext{"state", config})
	if err != nil {
		return nil, nil, err
	}

	return blockStore, stateDB, err
}
func DefaultGenesisDocProviderFunc(config *config.GlobalConfiguration) GenesisDocProvider {
	return func() (*models.GenesisDoc, error) {
		return models.GenesisDocFromFile(config.GenesisFile())
	}
}

func LoadStateFromDBOrGenesisDocProvider(stateDB services2.DB, genesisDocProvider GenesisDocProvider, ) (models.LatestState, *models.GenesisDoc, error) {
	// Get genesis doc
	genDoc, err := loadGenesisDoc(stateDB)
	if err != nil {
		genDoc, err = genesisDocProvider()
		if err != nil {
			return models.LatestState{}, nil, err
		}
		err = genDoc.ValidateAndComplete()
		if err != nil {
			return models.LatestState{}, nil, fmt.Errorf("error in genesis doc: %w", err)
		}
		if err := saveGenesisDoc(stateDB, genDoc); err != nil {
			return models.LatestState{}, nil, err
		}
	}
	stateStore := statestore.NewDefaultStateStore(stateDB)
	state, err := stateStore.LoadFromDBOrGenesisDoc(genDoc)
	if err != nil {
		return models.LatestState{}, nil, err
	}
	return state, genDoc, nil
}

func saveGenesisDoc(db services2.DB, genDoc *models.GenesisDoc) error {
	b, err := json.Marshal(genDoc)
	if err != nil {
		return fmt.Errorf("failed to save genesis doc due to marshaling error: %w", err)
	}
	if err := db.SetSync(genesisDocKey, b); err != nil {
		return err
	}

	return nil
}

// panics if failed to unmarshal bytes
func loadGenesisDoc(db services2.DB) (*models.GenesisDoc, error) {
	b, err := db.Get(genesisDocKey)
	if err != nil {
		panic(err)
	}
	if len(b) == 0 {
		return nil, errors.New("genesis doc not found")
	}
	var genDoc *models.GenesisDoc
	err = json.Unmarshal(b, &genDoc)
	if err != nil {
		panic(fmt.Sprintf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, b))
	}
	return genDoc, nil
}

func (this *blockChainLogicServiceImpl) OnStart() error {
	this.consensusLogicService.Start(services3.ASYNC_START_WAIT_READY)

	return nil
}

func (b *blockChainLogicServiceImpl) GetServiceId() types.ServiceID {
	return LOGICSERVICE_BLOCKCHAIN
}

func (b *blockChainLogicServiceImpl) ChoseInterestProtocolsOrTopics(pProtocol *models2.P2PProtocol) error {
	return nil
}

func (b *blockChainLogicServiceImpl) ChoseInterestEvents(bus base.IP2PEventBusComponent) {
}
