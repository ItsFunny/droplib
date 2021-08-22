/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:48 上午
# @File : from_model.go
# @Description :
# @Attention :
*/
package events

import "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"

func RoundStateEvent(rs models.RoundState) EventDataRoundState {
	return EventDataRoundState{
		Height: rs.Height,
		Round:  rs.Round,
		Step:   rs.Step.String(),
	}
}


func  NewRoundEvent(rs models.RoundState) EventDataNewRound {
	addr := rs.Validators.GetProposer().Address
	idx, _ := rs.Validators.GetByAddress(addr)

	return EventDataNewRound{
		Height: rs.Height,
		Round:  rs.Round,
		Step:   rs.Step.String(),
		Proposer: models.ValidatorInfo{
			Address: addr,
			Index:   idx,
		},
	}
}
func  CompleteProposalEvent(rs models.RoundState) EventDataCompleteProposal {
	// We must construct BlockID from ProposalBlock and ProposalBlockParts
	// cs.Proposal is not guaranteed to be set when this function is called
	blockID := models.BlockID{
		Hash:          rs.ProposalBlock.Hash(),
		PartSetHeader: rs.ProposalBlockParts.Header(),
	}

	return EventDataCompleteProposal{
		Height:  rs.Height,
		Round:   rs.Round,
		Step:    rs.Step.String(),
		BlockID: blockID,
	}
}