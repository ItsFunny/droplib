/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 2:02 下午
# @File : block.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
)

type IBlockStore interface {
	services.IBaseService
	Base() uint64
	LoadBlock(height int64) IBlock

	// 获取block的meta数据
	LoadBlockMeta(height uint64) *models.BlockMeta
	Height() uint64
	LoadBlockPart(height uint64, index int) *models.Part
	LoadBlockCommit(height uint64) *models.Commit
	LoadSeenCommit(height uint64) *models.Commit
	SaveBlock(block *models.TendermintBlockWrapper, blockParts *models.PartSet, seenCommit *models.Commit)
	PruneBlocks(height uint64) (uint64, error)
}
