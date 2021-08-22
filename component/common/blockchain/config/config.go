/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 3:34 下午
# @File : config.go
# @Description :
# @Attention :
*/
package config

type BaseConfiguration struct {
	ChainId string
	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`

	// Path to the JSON file containing the initial validator set and other meta data
	Genesis string `mapstructure:"genesis-file"`

	NodeUintID uint16 `mapstructure:"node-id"`
}

func (cfg BaseConfiguration) ChainID() string {
	return cfg.ChainId
}

type GlobalConfiguration struct {
	BaseConfiguration
	MemPoolConfiguration   *MemPoolConfiguration
	BlockConfiguration     *BlockConfiguration
	EvidenceConfiguration  *EvidenceConfiguration
	ValidatorConfiguration *ValidatorConfiguration
	ConsensusConfiguration *ConsensusConfiguration
	DBConfiguration        *DBConfiguration
	TicketConfiguration    *TicketConfiguration
	P2PConfiguration       *P2PConfiguration
}

func NewDefaultGlobalConfiguration() *GlobalConfiguration {
	r := GlobalConfiguration{
		BaseConfiguration:      *DefaultBaseConfiguration(),
		MemPoolConfiguration:   NewDefaultMemPoolConfiguration(),
		BlockConfiguration:     NewDefaultBlockConfiguration(),
		EvidenceConfiguration:  NewDefaultEvidenceConfiguration(),
		ValidatorConfiguration: NewDefaultValidatorConfiguration(),
		ConsensusConfiguration: NewDefaultConsensusConfiguration(),
		DBConfiguration:        NewDefaultDBConfiguration(),
		TicketConfiguration:    NewDefaultTicketConfiguration(),
		P2PConfiguration:       NewDefaultP2PConfiguration(),
	}
	return &r
}
func DefaultBaseConfiguration() *BaseConfiguration {
	r := &BaseConfiguration{
		ChainId: "test_chain",
		RootDir: "./testdata",
		Genesis: "./a.json",
	}

	return r
}

// DBDir returns the full path to the database directory
func (cfg GlobalConfiguration) DBDir() string {
	return rootify(cfg.DBConfiguration.DBPath, cfg.RootDir)
}

// GenesisFile returns the full path to the genesis.json file
func (cfg GlobalConfiguration) GenesisFile() string {
	return rootify(cfg.Genesis, cfg.RootDir)
}
