/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 9:11 下午
# @File : validation.go
# @Description :
# @Attention :
*/
package block

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/libs"
	"time"
)

func validateBlock(state models.LatestState, block *models.TendermintBlockWrapper) error {
	// Validate internal consistency.
	if err := block.ValidateBasic(); err != nil {
		return err
	}

	// Validate basic info.
	// if block.Version.App != state.Version.Consensus.App ||
	// 	block.Version.Block != state.Version.Consensus.Block {
	// 	return fmt.Errorf("wrong Block.Header.Version. Expected %v, got %v",
	// 		state.Version.Consensus,
	// 		block.Version,
	// 	)
	// }
	if block.ChainID != state.ChainID {
		return fmt.Errorf("wrong Block.Header.ChainID. Expected %v, got %v",
			state.ChainID,
			block.ChainID,
		)
	}
	if state.LastBlockHeight == 0 && uint64(block.Height) != state.InitialHeight {
		return fmt.Errorf("wrong Block.Header.Height. Expected %v for initial block, got %v",
			block.Height, state.InitialHeight)
	}
	if state.LastBlockHeight > 0 && block.Height != state.LastBlockHeight+1 {
		return fmt.Errorf("wrong Block.Header.Height. Expected %v, got %v",
			state.LastBlockHeight+1,
			block.Height,
		)
	}
	// Validate prev block info.
	if !block.LastBlockID.Equals(state.LastBlockID) {
		return fmt.Errorf("wrong Block.Header.LastBlockID.  Expected %v, got %v",
			state.LastBlockID,
			block.LastBlockID,
		)
	}

	// Validate app info
	// if !bytes.Equal(block.AppHash, state.AppHash) {
	// 	return fmt.Errorf("wrong Block.Header.AppHash.  Expected %X, got %v",
	// 		state.AppHash,
	// 		block.AppHash,
	// 	)
	// }

	hashCP := state.ConsensusParams.HashConsensusParams()
	if !bytes.Equal(block.ConsensusHash, hashCP) {
		return fmt.Errorf("wrong Block.Header.ConsensusHash.  Expected %X, got %v",
			hashCP,
			block.ConsensusHash,
		)
	}
	if !bytes.Equal(block.LastResultsHash, state.LastResultsHash) {
		return fmt.Errorf("wrong Block.Header.LastResultsHash.  Expected %X, got %v",
			state.LastResultsHash,
			block.LastResultsHash,
		)
	}
	if !bytes.Equal(block.ValidatorsHash, state.Validators.Hash()) {
		return fmt.Errorf("wrong Block.Header.ValidatorsHash.  Expected %X, got %v",
			state.Validators.Hash(),
			block.ValidatorsHash,
		)
	}
	if !bytes.Equal(block.NextValidatorsHash, state.NextValidators.Hash()) {
		return fmt.Errorf("wrong Block.Header.NextValidatorsHash.  Expected %X, got %v",
			state.NextValidators.Hash(),
			block.NextValidatorsHash,
		)
	}

	// Validate block LastCommit.
	if uint64(block.Height) == state.InitialHeight {
		if len(block.LastCommit.Signatures) != 0 {
			return errors.New("initial block can't have LastCommit signatures")
		}
	} else {
		// LastCommit.Signatures length is checked in VerifyCommit.
		if err := state.LastValidators.VerifyCommit(
			state.ChainID, state.LastBlockID, uint64(block.Height-1), block.LastCommit); err != nil {
			return err
		}
	}

	// NOTE: We can't actually verify it's the right proposer because we don't
	// know what round the block was first proposed. So just check that it's
	// a legit address and a known validator.
	// The length is checked in ValidateBasic above.
	if !state.Validators.HasAddress(block.ProposerAddress) {
		return fmt.Errorf("block.Header.ProposerAddress %X is not a validator",
			block.ProposerAddress,
		)
	}

	// Validate block Time
	switch {
	case uint64(block.Height) > state.InitialHeight:
		if !block.Time.After(state.LastBlockTime) {
			return fmt.Errorf("block time %v not greater than last block time %v",
				block.Time,
				state.LastBlockTime,
			)
		}
		medianTime := MedianTime(block.LastCommit, state.LastValidators)
		if !block.Time.Equal(medianTime) {
			return fmt.Errorf("invalid block time. Expected %v, got %v",
				medianTime,
				block.Time,
			)
		}

	case uint64(block.Height) == state.InitialHeight:
		genesisTime := state.LastBlockTime
		if !block.Time.Equal(genesisTime) {
			return fmt.Errorf("block time %v is not equal to genesis time %v",
				block.Time,
				genesisTime,
			)
		}

	default:
		return fmt.Errorf("block height %v lower than initial height %v",
			block.Height, state.InitialHeight)
	}
	// Check evidence doesn't exceed the limit amount of bytes.
	if max, got := state.ConsensusParams.Evidence.MaxBytes, block.Evidence.ByteSize(); got > max {
		return models.NewErrEvidenceOverflow(max, got)
	}

	return nil
}
func MedianTime(commit *models.Commit, validators *models.ValidatorSet) time.Time {
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
