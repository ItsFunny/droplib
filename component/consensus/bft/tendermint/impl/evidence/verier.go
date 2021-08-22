/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 1:18 下午
# @File : verier.go
# @Description :
# @Attention :
*/
package evidence

import (
	"bytes"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
)

func isInvalidHeader(trusted, conflicting *models.TendermintBlockHeader) bool {
	return !bytes.Equal(trusted.ValidatorsHash, conflicting.ValidatorsHash) ||
		!bytes.Equal(trusted.NextValidatorsHash, conflicting.NextValidatorsHash) ||
		!bytes.Equal(trusted.ConsensusHash, conflicting.ConsensusHash) ||
		// !bytes.Equal(trusted.AppHash, conflicting.AppHash) ||
		!bytes.Equal(trusted.LastResultsHash, conflicting.LastResultsHash)
}
