/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 6:42 下午
# @File : block.go
# @Description :
# @Attention :
*/
package services

import "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"

type IBlock interface {
	// ToPart()*models.Part
	// ToCommit()*models.Commit
}


type IBlockSupport interface {
	Block(number uint64) IBlock
	Height() uint64
}

type IBlockExecutor interface {
	ValidateBlock(state models.LatestState, block *models.TendermintBlockWrapper) error
	ApplyBlock(state models.LatestState, blockID models.BlockID, block *models.TendermintBlockWrapper) (models.LatestState, uint64, error)
	CreateProposalBlock(height int64, state models.LatestState, commit *models.Commit, proposerAddr []byte) (*models.TendermintBlockWrapper, *models.PartSet)
	// SetEventBus(bus types.IEventBus)
	GetStateStore() IStateStore
}