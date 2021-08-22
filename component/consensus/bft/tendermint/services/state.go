/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:33 上午
# @File : state.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
)

type IStateComponent interface {
	services.IBaseService
	StateNotify() <-chan models.MsgInfo

	PeerMsgNotify() chan models.MsgInfo

	LoadCommit(height uint64) *models.Commit

	GetInitialHeight(lock bool) uint64

	GetCurrentHeightWithVote(lock bool) (uint64, *models.HeightVoteSet)
	GetSizeWithHeightWithLastCommitSize(lock bool) (uint64, int, int)

	GetRoundState() *models.RoundState

	GetBlockStore()IBlockStore
}


type IStateStore interface {
	// LoadFromDBOrGenesisFile loads the most recent state.
	// If the chain is new it will use the genesis file from the provided genesis file path as the current state.
	LoadFromDBOrGenesisFile(string) (models.LatestState, error)
	// LoadFromDBOrGenesisDoc loads the most recent state.
	// If the chain is new it will use the genesis doc as the current state.
	LoadFromDBOrGenesisDoc(*models.GenesisDoc) (models.LatestState, error)
	// Load loads the current state of the blockchain
	Load() (models.LatestState, error)
	// LoadValidators loads the validator set at a given height
	LoadValidators(int64) (*models.ValidatorSet, error)
	// LoadConsensusParams loads the consensus params for a given height
	LoadConsensusParams(int64) (models.ConsensusParams, error)
	// Save overwrites the previous state with the updated one
	Save(state models.LatestState) error
	// Bootstrap is used for bootstrapping state when not starting from a initial height.
	Bootstrap(latestState models.LatestState) error
	// PruneStates takes the height from which to prune up to (exclusive)
	PruneStates(uint642 uint64) error
}