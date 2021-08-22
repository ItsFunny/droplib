/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:55 上午
# @File : block_store_test.go.go
# @Description :
# @Attention :
*/
package block

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"testing"
	"time"
)

var (
	db services.DB
)

func prepare() {
	d, err := NewGoLevelDB("test_block_store", ".")
	if nil != err {
		panic(err)
	}
	db = d
}

func Test_SaveBlock(t *testing.T) {
	storeImpl := NewBlockStoreImpl(db)
	block := makeMockBlock()
	set := block.MakePartSet(12333)
	storeImpl.SaveBlock(block, set, nil)
}

func makeMockBlock() *models.TendermintBlockWrapper {
	r := models.TendermintBlockWrapper{
		FabricBlockMetadata:     [][]byte{[]byte{123}},
		TendermintBlockHeader:   models.TendermintBlockHeader{},
		TendermintBlockMetadata: models.TendermintBlockMetadata{},
		TendermintBlockData:     models.TendermintBlockData{},
		Evidence:                models.EvidenceData{},
		LastCommit:              nil,
	}

	return &r
}

func makeMockBlockHeader() models.TendermintBlockHeader {
	r := models.TendermintBlockHeader{
		ChainID:            "test",
		Height:             0,
		Time:               time.Now(),
		LastBlockID:        models.BlockID{},
		LastCommitHash:     nil,
		DataHash:           nil,
		ValidatorsHash:     nil,
		NextValidatorsHash: nil,
		ConsensusHash:      nil,
		LastResultsHash:    nil,
		EvidenceHash:       nil,
		ProposerAddress:    nil,
	}
	return r
}
