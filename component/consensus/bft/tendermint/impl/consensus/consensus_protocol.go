/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/14 9:06 上午
# @File : consensus_protocol.go
# @Description :
# @Attention :
*/
package consensus

import (
	"github.com/hyperledger/fabric-droplib/component/p2p/models"
	"github.com/hyperledger/fabric-droplib/protos"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/helper"
)



func GetConsensusProtocol() models.ProtocolInfo {
	info := models.ProtocolInfo{
		ProtocolId:  CONSENSUS_PROTOCOL_ID,
		MessageType: &protos.StreamCommonProtocolWrapper{},
		Channels:    nil,
	}
	consensusChannels := make([]models.ChannelInfo, 0)

	// state
	stateChannelInfo := models.ChannelInfo{
		ChannelId:   StateChannel,
		ChannelSize: 20,
		MessageType: &helper.ConsensusStateChannelMessageWrapper{},
	}

	// data
	dataChannelInfo := models.ChannelInfo{
		ChannelId:   DataChannel,
		ChannelSize: 20,
		MessageType: &helper.ConsensusDataChannelMessageWrapper{},
	}

	// vote
	voteChannelInfo := models.ChannelInfo{
		ChannelId:   VoteChannel,
		ChannelSize: 20,
		MessageType: &helper.ConsensusVoteChannelMessageWrapper{},
	}

	// voteSetBits
	voteSetBitsChannelInfo := models.ChannelInfo{
		ChannelId:   VoteSetBitsChannel,
		ChannelSize: 20,
		MessageType: &helper.ConsensusVoteSetBitsChannelMessageWrapper{},
	}

	consensusChannels = append(consensusChannels, stateChannelInfo, dataChannelInfo, voteChannelInfo, voteSetBitsChannelInfo)
	info.Channels = consensusChannels

	return info
}
