/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 7:38 下午
# @File : latest_state.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"fmt"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/constants"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/libs/json"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/states"
	"io/ioutil"
	"time"
)

type LatestState struct {
	// FIXME 也是没用的东西
	// Version Version

	// immutable
	ChainID       string
	InitialHeight uint64 // should be 1, not 0, when starting from height 1

	// LastBlockHeight=0 at genesis (ie. block(H=0) does not exist)
	LastBlockHeight int64
	LastBlockID     BlockID
	LastBlockTime   time.Time

	// LastValidators is used to validate block.LastCommit.
	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1 + 1
	// Extra +1 due to nextValSet delay.
	NextValidators              *ValidatorSet
	Validators                  *ValidatorSet
	LastValidators              *ValidatorSet
	LastHeightValidatorsChanged int64

	// Consensus parameters used for validating blocks.
	// Changes returned by EndBlock and updated after Commit.
	ConsensusParams                  ConsensusParams
	LastHeightConsensusParamsChanged int64

	// Merkle root of the results from executing prev block
	LastResultsHash []byte

	// the latest AppHash we've received from calling abci.Commit()
	// FIXME 没用的东西
	AppHash []byte
}

func (state LatestState) Copy() LatestState {

	return LatestState{
		// Version:       state.Version,
		ChainID:       state.ChainID,
		InitialHeight: state.InitialHeight,

		LastBlockHeight: state.LastBlockHeight,
		LastBlockID:     state.LastBlockID,
		LastBlockTime:   state.LastBlockTime,

		NextValidators:              state.NextValidators.Copy(),
		Validators:                  state.Validators.Copy(),
		LastValidators:              state.LastValidators.Copy(),
		LastHeightValidatorsChanged: state.LastHeightValidatorsChanged,

		ConsensusParams:                  state.ConsensusParams,
		LastHeightConsensusParamsChanged: state.LastHeightConsensusParamsChanged,

		AppHash: state.AppHash,

		LastResultsHash: state.LastResultsHash,
	}
}
func (state LatestState) IsEmpty() bool {
	return state.Validators == nil // XXX can't compare to Empty
}

func (state LatestState) MakeBlock(
	height int64,
	txs []types.Tx,
	commit *Commit,
	evidence []transfer.Evidence,
	proposerAddress []byte,
) (*TendermintBlockWrapper, *PartSet) {

	// Build base block with block data.
	block := MakeBlock(height, txs, commit, evidence)

	// Set time.
	var timestamp time.Time
	if uint64(height) == state.InitialHeight {
		timestamp = state.LastBlockTime // genesis time
	} else {
		timestamp = MedianTime(commit, state.LastValidators)
	}

	// Fill rest of header with state data.
	block.TendermintBlockHeader.Populate(state.ChainID,
		timestamp, state.LastBlockID,
		state.Validators.Hash(), state.NextValidators.Hash(),
		state.ConsensusParams.HashConsensusParams(), state.AppHash, state.LastResultsHash,
		proposerAddress,
	)

	return block, block.MakePartSet(constants.BlockPartSizeBytes)
}
func MedianTime(commit *Commit, validators *ValidatorSet) time.Time {
	weightedTimes := make([]*libs.WeightedTime, len(commit.Signatures))
	totalVotingPower := int64(0)

	for i, commitSig := range commit.Signatures {
		if commitSig.Absent() {
			continue
		}
		_, validator := validators.GetByAddress(commitSig.ValidatorAddress)
		// If there's no condition, TestValidateBlockCommit panics; not needed normally.
		if validator != nil {
			totalVotingPower += validator.VotingPower
			weightedTimes[i] = libs.NewWeightedTime(commitSig.Timestamp, validator.VotingPower)
		}
	}

	return libs.WeightedMedian(weightedTimes, totalVotingPower)
}

func MakeBlock(height int64, txs []types.Tx, lastCommit *Commit, evidence []transfer.Evidence) *TendermintBlockWrapper {
	block := &TendermintBlockWrapper{
		TendermintBlockHeader: TendermintBlockHeader{
			// Version: version.Consensus{Block: version.BlockProtocol, App: 0},
			Height: height,
		},
		TendermintBlockData: TendermintBlockData{
			Txs: txs,
		},
		Evidence:   EvidenceData{Evidence: evidence},
		LastCommit: lastCommit,
	}
	block.fillHeader()
	return block
}
func MakeGenesisStateFromFile(genDocFile string) (LatestState, error) {
	genDoc, err := MakeGenesisDocFromFile(genDocFile)
	if err != nil {
		return LatestState{}, err
	}
	return MakeGenesisState(genDoc),nil
}
func MakeGenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	genDocJSON, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc file: %v", err)
	}
	genDoc, err := GenesisDocFromJSON(genDocJSON)
	if err != nil {
		return nil, fmt.Errorf("error reading GenesisDoc: %v", err)
	}
	return genDoc, nil
}

func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	genDoc := GenesisDoc{}
	err := json.Unmarshal(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return &genDoc, err
}
func MakeGenesisState(genDoc *GenesisDoc) LatestState {
	var validatorSet, nextValidatorSet *ValidatorSet
	if genDoc.Validators == nil {
		validatorSet = NewValidatorSet(nil)
		nextValidatorSet = NewValidatorSet(nil)
	} else {
		validators := make([]*Validator, len(genDoc.Validators))
		for i, val := range genDoc.Validators {
			validators[i] = NewValidator(val.PubKey.ToPublicKey(), val.Power)
		}
		validatorSet = NewValidatorSet(validators)
		nextValidatorSet = NewValidatorSet(validators).CopyIncrementProposerPriority(1)
	}
	return LatestState{
		ChainID:       genDoc.ChainID,
		InitialHeight: uint64(genDoc.InitialHeight),

		LastBlockHeight: -1,
		LastBlockID:     BlockID{},
		LastBlockTime:   genDoc.GenesisTime,

		NextValidators:              nextValidatorSet,
		Validators:                  validatorSet,
		LastValidators:              NewValidatorSet(nil),
		LastHeightValidatorsChanged: genDoc.InitialHeight,

		ConsensusParams:                  *genDoc.ConsensusParams,
		LastHeightConsensusParamsChanged: genDoc.InitialHeight,

		AppHash: genDoc.AppHash,
	}
}

func (state LatestState) Bytes() []byte {
	sm, err := state.ToProto()
	if err != nil {
		panic(err)
	}
	bz, err := commonutils.Marshal(sm)
	if err != nil {
		panic(err)
	}
	return bz
}

func (state *LatestState) ToProto() (*states.LatestStateProto, error) {
	if state == nil {
		return nil, errors.New("state is nil")
	}

	sm := new(states.LatestStateProto)

	// sm.Version = state.Version.ToProto()
	sm.ChainId = state.ChainID
	sm.InitialHeight = int64(state.InitialHeight)
	sm.LastBlockHeight = int64(state.LastBlockHeight)

	sm.LastBlockId = state.LastBlockID.ToProto()
	sm.LastBlockTime = commonutils.ConvtTime2GoogleTime(state.LastBlockTime)
	vals, err := state.Validators.ToProto()
	if err != nil {
		return nil, err
	}
	sm.Validators = vals

	nVals, err := state.NextValidators.ToProto()
	if err != nil {
		return nil, err
	}
	sm.NextValidators = nVals

	if state.LastBlockHeight >= 1 { // At Block 1 LastValidators is nil
		lVals, err := state.LastValidators.ToProto()
		if err != nil {
			return nil, err
		}
		sm.LastValidators = lVals
	}

	sm.LastHeightValidatorsChanged = state.LastHeightValidatorsChanged
	sm.ConsensusParams = state.ConsensusParams.ToProto()
	sm.LastHeightConsensusParamsChanged = state.LastHeightConsensusParamsChanged
	sm.LastResultsHash = state.LastResultsHash
	sm.AppHash = state.AppHash

	return sm, nil
}
