/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/13 1:59 下午
# @File : genesis.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	"io/ioutil"
	"time"
)

const (
	// MaxChainIDLen is a maximum length of the chain ID.
	MaxChainIDLen = 50
)

type Base64PubKeyString string

func (this Base64PubKeyString) ToPublicKey() cryptolibs.IPublicKey {
	return cryptolibs.NewMockPublicKey(string(this))
}

type GenesisValidator struct {
	Address cryptolibs.Address `json:"address"`
	PubKey  Base64PubKeyString `json:"pub_key"`
	Power   int64              `json:"power"`
	Name    string             `json:"name"`
}

type GenesisValidatorNode struct {
	Address cryptolibs.Address `json:"address"`
	// base64 编码
	PubKey string `json:"pub_key"`
	Power  int64  `json:"power"`
	Name   string `json:"name"`
}

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time          `json:"genesis_time"`
	ChainID         string             `json:"chain_id"`
	InitialHeight   int64              `json:"initial_height"`
	ConsensusParams *ConsensusParams   `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator `json:"validators,omitempty"`
	AppHash         libs.HexBytes      `json:"app_hash"`
	AppState        json.RawMessage    `json:"app_state,omitempty"`
}

func initDemoGensisDoc() {

}

func NewDefaultGensisiDoc() *GenesisDoc {
	b64Pubk := ""
	r := &GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         "test_chain",
		InitialHeight:   0,
		ConsensusParams: DefaultConsensusParams(),
		Validators:      nil,
		AppHash:         nil,
		AppState:        nil,
	}

	return r
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc file: %w", err)
	}
	genDoc, err := GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading GenesisDoc at %s: %w", genDocFile, err)
	}
	return genDoc, nil
}

func (genDoc *GenesisDoc) ValidateAndComplete() error {
	if genDoc.ChainID == "" {
		return errors.New("genesis doc must include non-empty chain_id")
	}
	if len(genDoc.ChainID) > MaxChainIDLen {
		return fmt.Errorf("chain_id in genesis doc is too long (max: %d)", MaxChainIDLen)
	}
	if genDoc.InitialHeight < 0 {
		return fmt.Errorf("initial_height cannot be negative (got %v)", genDoc.InitialHeight)
	}
	if genDoc.InitialHeight == 0 {
		genDoc.InitialHeight = 0
	}

	if genDoc.ConsensusParams == nil {
		genDoc.ConsensusParams = DefaultConsensusParams()
	} else if err := genDoc.ConsensusParams.ValidateConsensusParams(); err != nil {
		return err
	}

	for i, v := range genDoc.Validators {
		if v.Power == 0 {
			return fmt.Errorf("the genesis file cannot contain validators with no voting power: %v", v)
		}
		if len(v.Address) > 0 && !bytes.Equal(v.PubKey.ToPublicKey().Address(), v.Address) {
			return fmt.Errorf("incorrect address for validator %v in the genesis file, should be %v", v, v.PubKey.ToPublicKey().Address())
		}
		if len(v.Address) == 0 {
			genDoc.Validators[i].Address = v.PubKey.ToPublicKey().Address()
		}
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = libs.Now()
	}

	return nil
}
