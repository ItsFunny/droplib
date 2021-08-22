/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 11:47 上午
# @File : evidence_pool_model.go
# @Description :
# @Attention :
*/
package evidence

import "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"

type duplicateVoteSet struct {
	VoteA *models.Vote
	VoteB *models.Vote
}
