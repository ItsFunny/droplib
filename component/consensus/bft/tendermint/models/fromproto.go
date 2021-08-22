/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/7 6:17 上午
# @File : fromproto.go
# @Description :
# @Attention :
*/
package models

import (
	"errors"
	"fmt"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	types2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	"github.com/hyperledger/fabric-droplib/protos/libs/bits"
	consensus2 "github.com/hyperledger/fabric-droplib/protos/protobufs/consensus"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/states"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

// FromProto converts a proto generetad type to a handwritten type
// return type, nil if everything converts safely, otherwise nil, error
func VoteFromProto(pv *types.VoteProto) (*Vote, error) {
	if pv == nil {
		return nil, errors.New("nil vote")
	}

	blockID, err := BlockIDFromProto(pv.BlockId)
	if err != nil {
		return nil, err
	}

	vote := new(Vote)
	vote.Type = pv.Type
	vote.Height = uint64(pv.Height)
	vote.Round = pv.Round
	vote.BlockID = *blockID
	vote.Timestamp = commonutils.ConvGoogleTime2Time(pv.Timestamp)
	vote.ValidatorAddress = pv.ValidatorAddress
	vote.ValidatorIndex = pv.ValidatorIndex
	vote.Signature = pv.Signature

	return vote, vote.ValidateBasic()
}



// FromProto converts a proto generetad type to a handwritten type
// return type, nil if everything converts safely, otherwise nil, error

func NewRoundStepFromProto(msg *consensus2.NewRoundStepProtoWrapper) (*NewRoundStepMessage, error) {
	v := msg.NewRoundStep
	r := &NewRoundStepMessage{
		Height:                uint64(v.Height),
		Round:                 v.Round,
		Step:                  types2.RoundStepType(v.Step),
		SecondsSinceStartTime: v.SecondsSinceStartTime,
		LastCommitRound:       v.LastCommitRound,
	}
	return r, nil
}

func NewValidBlockFromProto(wrapper *consensus2.NewValidBlockProtoWrapper) (*NewValidBlockMessage, error) {
	r := &NewValidBlockMessage{
		Height:             uint64(wrapper.Height),
		Round:              wrapper.Round,
		BlockPartSetHeader: BlockPartSetHeaderFromProto(wrapper.BlockPartSetHeader),
		BlockParts:         BlockPartsFromProto(wrapper.BlockParts),
		IsCommit:           wrapper.IsCommit,
	}
	return r, nil
}
func BitArrayFromProto(p *bits.BitArrayProto) *libs.BitArray {
	r := libs.BitArray{
		Bits:  int(p.Bits),
		Elems: p.Elems,
	}
	return &r
}
func BlockPartsFromProto(p *bits.BitArrayProto) *libs.BitArray {
	r := libs.BitArray{
		Bits:  int(p.Bits),
		Elems: p.Elems,
	}
	return &r
}
func BlockPartSetHeaderFromProto(proto *types.PartSetHeaderProto) PartSetHeader {
	r := PartSetHeader{
		Total: proto.Total,
		Hash:  proto.Hash,
	}
	return r
}
func HasVoteFromProto(wrapper *consensus2.HasVoteProtoWrapper) (*HasVoteMessage, error) {
	r := &HasVoteMessage{
		Height: uint64(wrapper.Height),
		Round:  wrapper.Round,
		Type:   wrapper.Type,
		Index:  wrapper.Index,
	}
	return r, nil
}

func VoteSetMAJ23FromProto(wrapper *consensus2.VoteSetMaj23ProtoWrapper) (*VoteSetMaj23Message, error) {
	msg := wrapper.VoteSetMaj23
	p, err := BlockIDFromProto(msg.BlockId)
	if nil != err {
		return nil, err
	}
	r := &VoteSetMaj23Message{
		Height:  msg.Height,
		Round:   msg.Round,
		Type:    msg.Type,
		BlockID: *p,
	}
	return r, nil
}

func ProposalFromProto(pp *consensus2.ProposalProtoWrapper) (*ProposalMessage, error) {
	if pp == nil {
		return nil, errors.New("nil proposal")
	}

	p := new(ProposalMessage)
	pro := new(Proposal)

	blockID, err := BlockIDFromProto(pp.Proposal.BlockId)
	if err != nil {
		return nil, err
	}
	pro.BlockID = *blockID
	pro.Type = pp.Proposal.Type
	pro.Height = uint64(pp.Proposal.Height)
	pro.Round = pp.Proposal.Round
	pro.POLRound = pp.Proposal.PolRound
	pro.Timestamp = commonutils.ConvGoogleTime2Time(commonutils.ConvtPBTime2GoogleTime(pp.Proposal.Timestamp))
	pro.Signature = pp.Proposal.Signature

	p.Proposal = pro
	p.ExtraData = pp.ExtraData

	return p, p.ValidateBasic()
}

func ProposalPolFromProto(pp *consensus2.ProposalPOLProtoWrapper) (*ProposalPOLMessage, error) {
	if nil == pp {
		return nil, errors.New("nil")
	}

	r := new(ProposalPOLMessage)

	r.Height = uint64(pp.Height)
	r.ProposalPOLRound = pp.ProposalPolRound
	r.ProposalPOL = BitArrayFromProto(pp.ProposalPol)

	return r, r.ValidateBasic()
}

func BlockPartFromProto(pp *consensus2.BlockPartProtoWrapper) (*BlockPartMessageWrapper, error) {
	r := new(BlockPartMessageWrapper)
	r.Height = pp.Height
	r.Round = pp.Round
	if part, err := PartFromProto(pp.Part); nil != err {
		return nil, err
	} else {
		r.Part = part
	}

	return r, r.ValidateBasic()
}

func VoteSetBitFromProto(pp *consensus2.VoteSetBitsProtoWrapper) (*VoteSetBitsMessage, error) {
	r := new(VoteSetBitsMessage)
	r.Round = pp.Round
	r.Height = uint64(pp.Height)
	r.Type = pp.Type
	if b, err := BlockIDFromProto(pp.BlockId); nil != err {
		return nil, err
	} else {
		r.BlockID = *b
	}
	r.Votes = BitArrayFromProto(pp.Votes)

	return r, r.ValidateBasic()
}


func PartFromProto(pb *types.PartProto) (*Part, error) {
	if pb == nil {
		return nil, errors.New("nil part")
	}

	part := new(Part)
	proof, err := merkle.ProofFromProto(pb.Proof)
	if err != nil {
		return nil, err
	}
	part.Index = pb.Index
	part.Bytes = pb.Bytes
	part.Proof = *proof

	return part, part.ValidateBasic()
}

func SignedHeaderFromProto(shp *types.SignedHeaderProto) (*SignedHeader, error) {
	if shp == nil {
		return nil, errors.New("nil SignedHeader")
	}

	sh := new(SignedHeader)

	if shp.Header != nil {
		h, err := HeaderFromProto(shp.Header)
		if err != nil {
			return nil, err
		}
		sh.TendermintBlockHeader = &h
	}

	if shp.Commit != nil {
		c, err := CommitFromProto(shp.Commit)
		if err != nil {
			return nil, err
		}
		sh.Commit = c
	}

	return sh, nil
}
func LightBlockFromProto(pb *types.LightBlockProto) (*LightBlock, error) {
	if pb == nil {
		return nil, errors.New("nil light block")
	}

	lb := new(LightBlock)

	if pb.SignedHeader != nil {
		sh, err := SignedHeaderFromProto(pb.SignedHeader)
		if err != nil {
			return nil, err
		}
		lb.SignedHeader = sh
	}

	if pb.ValidatorSet != nil {
		vals, err := ValidatorSetFromProto(pb.ValidatorSet)
		if err != nil {
			return nil, err
		}
		lb.ValidatorSet = vals
	}

	return lb, nil
}
func ValidatorSetFromProto(vp *types.ValidatorSetProto) (*ValidatorSet, error) {
	if vp == nil {
		return nil, errors.New("nil validator set") // validator set should never be nil, bigger issues are at play if empty
	}
	vals := new(ValidatorSet)

	valsProto := make([]*Validator, len(vp.Validators))
	for i := 0; i < len(vp.Validators); i++ {
		v, err := ValidatorFromProto(vp.Validators[i])
		if err != nil {
			return nil, err
		}
		valsProto[i] = v
	}
	vals.Validators = valsProto

	p, err := ValidatorFromProto(vp.GetProposer())
	if err != nil {
		return nil, fmt.Errorf("fromProto: validatorSet proposer error: %w", err)
	}

	vals.Proposer = p

	vals.totalVotingPower = vp.GetTotalVotingPower()

	return vals, vals.ValidateBasic()
}

func ValidatorFromProto(vp *types.ValidatorProto) (*Validator, error) {
	if vp == nil {
		return nil, errors.New("nil validator")
	}

	pk, err := PubKeyFromProto(vp.PubKey)
	if err != nil {
		return nil, err
	}
	v := new(Validator)
	v.Address = vp.GetAddress()
	v.PubKey = pk
	v.VotingPower = vp.GetVotingPower()
	v.ProposerPriority = vp.GetProposerPriority()

	return v, nil
}

func PubKeyFromProto(key []byte) (cryptolibs.IPublicKey,error) {
	return nil,nil
}
func LightClientAttackEvidenceFromProto(lpb *types.LightClientAttackEvidenceProto) (*LightClientAttackEvidence, error) {

	if lpb == nil {
		return nil, errors.New("empty light client attack evidence")
	}

	conflictingBlock, err := LightBlockFromProto(lpb.ConflictingBlock)
	if err != nil {
		return nil, err
	}

	byzVals := make([]*Validator, len(lpb.ByzantineValidators))
	for idx, valpb := range lpb.ByzantineValidators {
		val, err := ValidatorFromProto(valpb)
		if err != nil {
			return nil, err
		}
		byzVals[idx] = val
	}

	l := &LightClientAttackEvidence{
		ConflictingBlock:    conflictingBlock,
		CommonHeight:        lpb.CommonHeight,
		ByzantineValidators: byzVals,
		TotalVotingPower:    lpb.TotalVotingPower,
		Timestamp:           commonutils.ConvGoogleTime2Time(commonutils.ConvtPBTime2GoogleTime(lpb.Timestamp)),
	}

	return l, l.ValidateBasic()
}

func LatestStateFromProto(pb *states.LatestStateProto) (*LatestState, error) { //nolint:golint
	if pb == nil {
		return nil, errors.New("nil State")
	}

	state := new(LatestState)

	// state.Version = VersionFromProto(pb.Version)
	state.ChainID = pb.ChainId
	state.InitialHeight = uint64(pb.InitialHeight)

	bi, err := BlockIDFromProto(pb.LastBlockId)
	if err != nil {
		return nil, err
	}
	state.LastBlockID = *bi
	state.LastBlockHeight = pb.LastBlockHeight
	state.LastBlockTime = commonutils.ConvGoogleTime2Time(commonutils.ConvtPBTime2GoogleTime(pb.LastBlockTime))

	vals, err := ValidatorSetFromProto(pb.Validators)
	if err != nil {
		return nil, err
	}
	state.Validators = vals

	nVals, err := ValidatorSetFromProto(pb.NextValidators)
	if err != nil {
		return nil, err
	}
	state.NextValidators = nVals

	if state.LastBlockHeight >= 1 { // At Block 1 LastValidators is nil
		lVals, err := ValidatorSetFromProto(pb.LastValidators)
		if err != nil {
			return nil, err
		}
		state.LastValidators = lVals
	} else {
		state.LastValidators = NewValidatorSet(nil)
	}

	state.LastHeightValidatorsChanged = pb.LastHeightValidatorsChanged
	state.ConsensusParams = ConsensusParamsFromProto(pb.ConsensusParams)
	state.LastHeightConsensusParamsChanged = pb.LastHeightConsensusParamsChanged
	state.LastResultsHash = pb.LastResultsHash
	state.AppHash = pb.AppHash

	return state, nil
}