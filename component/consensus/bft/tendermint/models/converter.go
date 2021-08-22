/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 4:44 下午
# @File : converter.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

// TendermintBlockProtoWrapper
func BlockFromProto(bp *types3.TendermintBlockProtoWrapper) (*TendermintBlockWrapper, error) {
	if bp == nil {
		return nil, errors.New("nil block")
	}

	b := new(TendermintBlockWrapper)
	h, err := HeaderFromProto(bp.Header)
	if err != nil {
		return nil, err
	}
	b.TendermintBlockHeader = h
	data, err := DataFromProto(bp.Data)
	if err != nil {
		return nil, err
	}
	b.TendermintBlockData = data
	if err := b.Evidence.FromProto(bp.Evidence); err != nil {
		return nil, err
	}

	if bp.LastCommit != nil {
		lc, err := CommitFromProto(bp.LastCommit)
		if err != nil {
			return nil, err
		}
		b.LastCommit = lc
	}

	return b, b.ValidateBasic()
}

func HeaderFromProto(ph *types3.TendermintBlockHeaderProto) (TendermintBlockHeader, error) {
	if ph == nil {
		return TendermintBlockHeader{}, errors.New("nil Header")
	}

	h := new(TendermintBlockHeader)

	bi, err := BlockIDFromProto(ph.LastBlockId)
	if err != nil {
		return TendermintBlockHeader{}, err
	}

	// h.Version = version.Consensus{Block: ph.Version.Block, App: ph.Version.App}
	h.ChainID = ph.ChainId
	h.Height = ph.Height
	h.Time = commonutils.ConvGoogleTime2Time(ph.Time)
	h.Height = ph.Height
	h.LastBlockID = *bi
	h.ValidatorsHash = ph.ValidatorsHash
	h.NextValidatorsHash = ph.NextValidatorsHash
	h.ConsensusHash = ph.ConsensusHash
	// h.AppHash = ph.AppHash
	h.DataHash = ph.DataHash
	h.EvidenceHash = ph.EvidenceHash
	h.LastResultsHash = ph.LastResultsHash
	h.LastCommitHash = ph.LastCommitHash
	h.ProposerAddress = ph.ProposerAddress

	return *h, h.ValidateBasic()
}

func DataFromProto(dp *types3.TendermintBlockDataProto) (TendermintBlockData, error) {
	if dp == nil {
		return TendermintBlockData{}, errors.New("nil data")
	}
	data := new(TendermintBlockData)

	if len(dp.Txs) > 0 {
		txBzs := make(types.Txs, len(dp.Txs))
		for i := range dp.Txs {
			txBzs[i] = types.Tx(dp.Txs[i])
		}
		data.Txs = txBzs
	} else {
		data.Txs = types.Txs{}
	}

	return *data, nil
}
func EvidenceFromProto(evidence *types3.EvidenceProto) (transfer.Evidence, error) {
	if evidence == nil {
		return nil, errors.New("nil evidence")
	}

	// switch evi := evidence.Sum.(type) {
	// case *types.DuplicateVoteEvidenceProto:
	// 	return DuplicateVoteEvidenceFromProto(evi.DuplicateVoteEvidence)
	// case *types.LightClientAttackEvidenceProto:
	// 	return LightClientAttackEvidenceFromProto(evi.LightClientAttackEvidence)
	// default:
	// 	return nil, errors.New("evidence is not recognized")
	// }
	return nil,errors.New("fix me ")
}

func CommitFromProto(cp *types3.CommitProto) (*Commit, error) {
	if cp == nil {
		return nil, errors.New("nil Commit")
	}

	var (
		commit = new(Commit)
	)

	bi, err := BlockIDFromProto(cp.BlockId)
	if err != nil {
		return nil, err
	}

	sigs := make([]CommitSig, len(cp.Signatures))
	for i := range cp.Signatures {
		if err := sigs[i].FromProto(cp.Signatures[i]); err != nil {
			return nil, err
		}
	}
	commit.Signatures = sigs

	commit.Height = uint64(cp.Height)
	commit.Round = cp.Round
	commit.BlockID = *bi

	return commit, commit.ValidateBasic()
}