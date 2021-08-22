/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:47 上午
# @File : params.go
# @Description :
# @Attention :
*/
package config

import "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"

func ConsensusParamsFromConfig(this *GlobalConfiguration) *models.ConsensusParams {
	ret := models.ConsensusParams{
		Block: models.BlockParams{
			MaxBytes: this.BlockConfiguration.MaxBytes,
			MaxGas:   this.BlockConfiguration.MaxGas,
		},
		Evidence: models.EvidenceParams{
			MaxAgeNumBlocks: this.EvidenceConfiguration.MaxAgeNumBlocks,
			MaxAgeDuration:  this.EvidenceConfiguration.MaxAgeDuration,
			MaxBytes:        this.EvidenceConfiguration.MaxBytes,
		},
		Validator: models.ValidatorParams{
			PubKeyTypes: this.ValidatorConfiguration.PubKeyTypes,
		},
	}
	return &ret
}
