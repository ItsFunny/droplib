/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 4:22 下午
# @File : state_test.go
# @Description :
# @Attention :
*/
package state

import (
	"bytes"
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/impl/mock"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	types2 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"os"
	"testing"
)

func startTestRound(cs *State, height uint64, round int32) {
	cs.enterNewRound(uint64(height), round)
	cs.startRoutines(0)
}

func TestStateProposerSelection0(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(globalCfg.RootDir)
	})

	notifyC := make(chan int, 1)
	cs1, vss, mp, bus := RandFabricState(t, 4)
	StartTx(t, mp, notifyC)

	newRoundCh := Subscribe(bus, types.EventQueryNewRound)
	proposalCh := Subscribe(bus, types.EventQueryCompleteProposal)
	height, round := cs1.Height, cs1.Round

	fmt.Println(vss)

	startTestRound(cs1, height, round)

	EnsureNewRound(newRoundCh, height, round)

	prop := cs1.GetRoundState().Validators.GetProposer()
	pv := cs1.signer.GetPublicKey()
	address := pv.Address()
	if !bytes.Equal(prop.Address, address) {
		t.Fatalf("expected proposer to be validator %d. Got %X", 0, prop.Address)
	}

	// 创建tx ,完成一次proposal
	notifyC <- 20
	EnsureNewProposal(proposalCh, int64(height), round)

	rs := cs1.GetRoundState()
	signAddVotes(cs1, types2.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, rs.ProposalBlock.Hash(), rs.ProposalBlockParts.Header(), vss[1:]...)

	// 开启新的一轮,代表收到了2/3
	EnsureNewRound(newRoundCh, height+1, 0)

	prop = cs1.GetRoundState().Validators.GetProposer()
	// 因为proposer 的选举是 round base,所以预期会是 下一个
	pv1 := vss[1].GetPublicKey()
	addr := pv1.Address()
	if !bytes.Equal(prop.Address, addr) {
		panic(fmt.Sprintf("expected proposer to be validator %d. Got %X", 1, prop.Address))
	}
}

func TestStateProposerSelection2(t *testing.T) {
	notifyC := make(chan int, 1)
	cs1, vss, mp, bus := RandFabricState(t, 4)
	StartTx(t, mp, notifyC)

	height := cs1.Height

	newRoundCh := Subscribe(bus, types.EventQueryNewRound)
	// this time we jump in at round 2
	incrementRound(vss[1:]...)
	incrementRound(vss[1:]...)

	var round int32 = 2
	startTestRound(cs1, height, round)

	EnsureNewRound(newRoundCh, height, round) // wait for the new round

	// everyone just votes nil. we get a new proposer each round
	for i := int32(0); int(i) < len(vss); i++ {
		prop := cs1.GetRoundState().Validators.GetProposer()
		pvk := vss[int(i+round)%len(vss)].GetPublicKey()
		addr := pvk.Address()
		correctProposer := addr
		if !bytes.Equal(prop.Address, correctProposer) {
			panic(fmt.Sprintf(
				"expected RoundState.Validators.GetProposer() to be validator %d. Got %X",
				int(i+2)%len(vss),
				prop.Address))
		}

		rs := cs1.GetRoundState()
		signAddVotes(cs1, types2.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, nil, rs.ProposalBlockParts.Header(), vss[1:]...)
		EnsureNewRound(newRoundCh, height, i+round+1) // wait for the new round event each round
		incrementRound(vss[1:]...)
	}

}

func TestStateEnterProposeNoPrivValidator(t *testing.T) {
	t.Cleanup(
		func() {
			os.RemoveAll(globalCfg.RootDir)
		},
	)
	notifyC := make(chan int, 1)
	cs, _, mem, bus := RandFabricState(t, 4)
	StartTx(t, mem, notifyC)

	cs.SetPrivValidator(nil)
	height, round := cs.Height, cs.Round

	// Listen for propose timeout event
	timeoutCh := Subscribe(bus, types.EventQueryTimeoutPropose)

	startTestRound(cs, height, round)
	// if we're not a validator, EnterPropose should timeout
	ensureNewTimeout(timeoutCh, height, round, cs.config.TimeoutPropose.Nanoseconds())

	if cs.GetRoundState().Proposal != nil {
		t.Error("Expected to make no proposal, since no privValidator")
	}
}

func incrementRound(vss ...*mock.MockValidator) {
	for _, vs := range vss {
		vs.Round++
	}
}
