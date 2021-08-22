/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 7:39 下午
# @File : empty.go
# @Description :
# @Attention :
*/
package evidence

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
)

// EmptyEvidencePool is an empty implementation of EvidencePool, useful for testing. It also complies
// to the consensus evidence pool interface
type EmptyEvidencePool struct{}

func NewEmptyEvidencePool() *EmptyEvidencePool {
	return &EmptyEvidencePool{}
}
func (EmptyEvidencePool) PendingEvidence(maxBytes int64) (ev []transfer.Evidence, size int64) {
	return nil, 0
}
func (EmptyEvidencePool) AddEvidence(transfer.Evidence) error              { return nil }
func (EmptyEvidencePool) Update(models.LatestState, models.EvidenceList)   {}
func (EmptyEvidencePool) CheckEvidence(evList models.EvidenceList) error   { return nil }
func (EmptyEvidencePool) ReportConflictingVotes(voteA, voteB *models.Vote) {}
