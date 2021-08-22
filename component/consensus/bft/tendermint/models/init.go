/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/20 10:24 下午
# @File : init.go
# @Description :
# @Attention :
*/
package models

import tmjson "github.com/hyperledger/fabric-droplib/libs/json"

func init() {
	tmjson.RegisterType(&NewRoundStepMessage{}, "tendermint/NewRoundStepMessage")
	tmjson.RegisterType(&NewValidBlockMessage{}, "tendermint/NewValidBlockMessage")
	tmjson.RegisterType(&ProposalMessage{}, "tendermint/Proposal")
	tmjson.RegisterType(&ProposalPOLMessage{}, "tendermint/ProposalPOL")
	tmjson.RegisterType(&BlockPartMessage{}, "tendermint/BlockPart")
	tmjson.RegisterType(&VoteMessage{}, "tendermint/Vote")
	tmjson.RegisterType(&HasVoteMessage{}, "tendermint/HasVote")
	tmjson.RegisterType(&VoteSetMaj23Message{}, "tendermint/VoteSetMaj23")
	tmjson.RegisterType(&VoteSetBitsMessage{}, "tendermint/VoteSetBits")
}
