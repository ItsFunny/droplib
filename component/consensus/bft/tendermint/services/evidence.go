/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:16 上午
# @File : evidence.go
# @Description :
# @Attention :
*/
package services

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
)

// EvidencePool defines the EvidencePool interface used by State.
type IEvidencePool interface {
	PendingEvidence(maxBytes int64) (ev []transfer.Evidence, size int64)
	AddEvidence(transfer.Evidence) error
	Update(models.LatestState, models.EvidenceList)
	CheckEvidence(models.EvidenceList) error
}
