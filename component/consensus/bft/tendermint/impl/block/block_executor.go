/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 7:38 下午
# @File : block_executor.go
# @Description :
# @Attention :
*/
package block

import (
	"fmt"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/events"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/libs"
	"math"
	"strconv"
)

var (
	_ services.IBlockExecutor = (*DefaultBlockExecutor)(nil)
)

type DefaultBlockExecutor struct {
	logger v2.Logger
	// 需要持有memPool的引用,用于过渡tx
	memPool services.IMemPool
	evPool  services.IEvidencePool

	stateStore services.IStateStore

	eventBus services.IConsensusEventBusComponent
}

func NewDefaultBlockExecutor(memPool services.IMemPool, evPool services.IEvidencePool, ss services.IStateStore, bus services.IConsensusEventBusComponent) *DefaultBlockExecutor {
	r := &DefaultBlockExecutor{
		memPool:    memPool,
		evPool:     evPool,
		stateStore: ss,
		eventBus:   bus,
	}
	r.logger = logrusplugin.NewLogrusLogger(modules.MODULE_BLOCK_EXECUTOR)
	return r
}

func (d *DefaultBlockExecutor) ValidateBlock(state models.LatestState, block *models.TendermintBlockWrapper) error {
	err := validateBlock(state, block)
	if err != nil {
		return err
	}
	return d.evPool.CheckEvidence(block.Evidence.Evidence)
}

func (d *DefaultBlockExecutor) ApplyBlock(state models.LatestState, blockID models.BlockID, block *models.TendermintBlockWrapper) (models.LatestState, uint64, error) {
	// 校验区块是否合法
	if err := validateBlock(state, block); err != nil {
		return state, 0, commonerrors.ErrInvalidBlock(err)
	}

	// FIXME 这里暂时写死为nil,表明不需要更新
	st, err := updateState(state, blockID, &block.TendermintBlockHeader, nil)
	if err != nil {
		return state, 0, fmt.Errorf("commit failed for application: %v", err)
	}
	state = st

	if err := d.Commit(state, block); nil != err {
		return state, 0, err
	}

	// Update evpool with the latest state.
	d.evPool.Update(state, block.Evidence.Evidence)
	libs.Fail()

	if err := d.stateStore.Save(state); nil != err {
		return state, 0, err
	}

	// Events are fired after everything else.
	// NOTE: if we crash between Commit and Save, events wont be fired during replay
	fireEvents(d.logger, d.eventBus, block, nil)

	return state, 0, nil
}
func (blockExec *DefaultBlockExecutor) Commit(
	state models.LatestState,
	block *models.TendermintBlockWrapper,
) error {
	blockExec.memPool.Lock()
	defer blockExec.memPool.Unlock()

	// while mempool is Locked, flush to ensure all async requests have completed
	// in the ABCI app before Commit.
	// err := blockExec.memPool.FlushAppConn()
	// if err != nil {
	// 	blockExec.logger.Error("Client error during mempool.FlushAppConn", "err", err)
	// 	return nil, 0, err
	// }

	// Commit block

	blockExec.logger.Info(
		"Committed state",
		"height", block.Height,
		"txs", len(block.Txs),
	)

	// Update mempool.
	err := blockExec.memPool.Update(
		block.Height,
		block.Txs,
		nil, nil,
		// TxPreCheck(state),
		// TxPostCheck(state),
	)

	return err
}

func (d *DefaultBlockExecutor) CreateProposalBlock(height int64, state models.LatestState, commit *models.Commit, proposerAddr []byte) (*models.TendermintBlockWrapper, *models.PartSet) {
	d.logger.Info("create ProposalBlock ,height:", height)

	// maxBytes := state.ConsensusParams.Block.MaxBytes
	// maxGas := state.ConsensusParams.Block.MaxGas
	evidence, _ := d.evPool.PendingEvidence(math.MaxInt64)
	// 这里直接获取全部,而不是通过bytes 控制,因此此处暂时先写死maxInt64
	// 或者是获取指定条数的交易数量
	txs := d.memPool.AcquireOnePatch()
	if len(txs) == 0 {
		// 当到了这一步,必然是收到了通知才会到这一步
		// panic("此时对于无交易的事件,通通都是panic,后续注释即可")
		// d.logger.Info("交易为空,退出roundbase")
		// return nil,nil
	}

	d.logger.Info("获取到" + strconv.Itoa(len(txs)) + ",条交易")
	// 获取evidence
	return state.MakeBlock(height, txs, commit, evidence, proposerAddr)
}

func (d *DefaultBlockExecutor) GetStateStore() services.IStateStore {
	return d.stateStore
}

func fireEvents(
	logger v2.Logger,
	eventBus services.IConsensusEventBusComponent,
	block *models.TendermintBlockWrapper,
	validatorUpdates []*models.Validator,
) {
	if err := eventBus.PublishEventNewBlock(events.EventDataNewBlock{
		Block: block,
		// ResultBeginBlock: *abciResponses.BeginBlock,
		// ResultEndBlock:   *abciResponses.EndBlock,
	}); err != nil {
		logger.Error("Error publishing new block", "err", err)
	}
	if err := eventBus.PublishEventNewBlockHeader(events.EventDataNewBlockHeader{
		Header: block.TendermintBlockHeader,
		NumTxs: int64(len(block.Txs)),
		// ResultBeginBlock: *abciResponses.BeginBlock,
		// ResultEndBlock:   *abciResponses.EndBlock,
	}); err != nil {
		logger.Error("Error publishing new block header", "err", err)
	}

	if len(block.Evidence.Evidence) != 0 {
		for _, ev := range block.Evidence.Evidence {
			if err := eventBus.PublishEventNewEvidence(events.EventDataNewEvidence{
				Evidence: ev,
				Height:   block.Height,
			}); err != nil {
				logger.Error("Error publishing new evidence", "err", err)
			}
		}
	}
	// FIXME 有待商榷,是否需要发布tx信息
	// for i, tx := range block.TendermintBlockData.Txs {
	// 	if err := eventBus.PublishEventTx(types.EventDataTx{TxResult: abci.TxResult{
	// 		Height: block.Height,
	// 		Index:  uint32(i),
	// 		Tx:     tx,
	// 		Result: *(abciResponses.DeliverTxs[i]),
	// 	}}); err != nil {
	// 		logger.Error("Error publishing event TX", "err", err)
	// 	}
	// }

	if len(validatorUpdates) > 0 {
		if err := eventBus.PublishEventValidatorSetUpdates(
			events.EventDataValidatorSetUpdates{ValidatorUpdates: validatorUpdates}); err != nil {
			logger.Error("Error publishing event", "err", err)
		}
	}
}

func updateState(
	st models.LatestState,
	blockID models.BlockID,
	header *models.TendermintBlockHeader,
	validatorUpdates []*models.Validator,
) (models.LatestState, error) {

	// Copy the valset so we can apply changes from EndBlock
	// and update s.LastValidators and s.Validators.
	nValSet := st.NextValidators.Copy()

	// Update the validator set with the latest abciResponses.
	lastHeightValsChanged := st.LastHeightValidatorsChanged
	if len(validatorUpdates) > 0 {
		err := nValSet.UpdateWithChangeSet(validatorUpdates)
		if err != nil {
			return st, fmt.Errorf("error changing validator set: %v", err)
		}
		// Change results from this height but only applies to the next next height.
		lastHeightValsChanged = header.Height + 1 + 1
	}

	// Update validator proposer priority and set state variables.
	nValSet.IncrementProposerPriority(1)

	// Update the params with the latest abciResponses.
	nextParams := st.ConsensusParams
	lastHeightParamsChanged := st.LastHeightConsensusParamsChanged
	// if abciResponses.EndBlock.ConsensusParamUpdates != nil {
	// 	// NOTE: must not mutate s.ConsensusParams
	// 	nextParams = state.ConsensusParams.UpdateConsensusParams(abciResponses.EndBlock.ConsensusParamUpdates)
	// 	err := nextParams.ValidateConsensusParams()
	// 	if err != nil {
	// 		return state, fmt.Errorf("error updating consensus params: %v", err)
	// 	}
	//
	// 	state.Version.Consensus.App = nextParams.Version.AppVersion
	//
	// 	// Change results from this height but only applies to the next height.
	// 	lastHeightParamsChanged = header.Height + 1
	// }

	// nextVersion := state.Version

	// NOTE: the AppHash has not been populated.
	// It will be filled on state.Save.

	return models.LatestState{
		// Version:                          nextVersion,
		ChainID:                          st.ChainID,
		InitialHeight:                    st.InitialHeight,
		LastBlockHeight:                  header.Height,
		LastBlockID:                      blockID,
		LastBlockTime:                    header.Time,
		NextValidators:                   nValSet,
		Validators:                       st.NextValidators.Copy(),
		LastValidators:                   st.Validators.Copy(),
		LastHeightValidatorsChanged:      lastHeightValsChanged,
		ConsensusParams:                  nextParams,
		LastHeightConsensusParamsChanged: lastHeightParamsChanged,
		// LastResultsHash:                  ABCIResponsesResultsHash(abciResponses),
		AppHash: nil,
	}, nil
}
