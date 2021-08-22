/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/11 4:23 下午
# @File : consensus.go
# @Description :
# @Attention :
*/
package config

import (
	"path/filepath"
	"time"
)
var (
	defaultDataDir       = "data"
)
type ConsensusConfiguration struct {
	RootDir string `mapstructure:"home"`
	WalPath string `mapstructure:"wal-file"`
	walFile string // overrides WalPath if s
	PeerGossipSleepDuration     time.Duration `mapstructure:"peer-gossip-sleep-duration"`
	PeerQueryMaj23SleepDuration time.Duration `mapstructure:"peer-query-maj23-sleep-duration"`
	TimeoutPropose time.Duration `mapstructure:"timeout-propose"`
	TimeoutProposeDelta time.Duration `mapstructure:"timeout-propose-delta"`

	// Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)
	SkipTimeoutCommit bool `mapstructure:"skip-timeout-commit"`

	CreateEmptyBlocks         bool          `mapstructure:"create-empty-blocks"`
	CreateEmptyBlocksInterval time.Duration `mapstructure:"create-empty-blocks-interval"`

	TimeoutPrevote time.Duration `mapstructure:"timeout-prevote"`
	// How much the timeout-prevote increases with each round
	TimeoutPrevoteDelta time.Duration `mapstructure:"timeout-prevote-delta"`
	TimeoutCommit time.Duration `mapstructure:"timeout-commit"`

	TimeoutPrecommit time.Duration `mapstructure:"timeout-precommit"`
	// How much the timeout-precommit increases with each round
	TimeoutPrecommitDelta time.Duration `mapstructure:"timeout-precommit-delta"`
	DoubleSignCheckHeight int64 `mapstructure:"double-sign-check-height"`
}
func NewDefaultConsensusConfiguration()*ConsensusConfiguration{
	c:=&ConsensusConfiguration{
		WalPath:                     filepath.Join(defaultDataDir, "cs.wal", "wal"),
		TimeoutPropose:              3000 * time.Millisecond,
		TimeoutProposeDelta:         500 * time.Millisecond,
		TimeoutPrevote:              1000 * time.Millisecond,
		TimeoutPrevoteDelta:         500 * time.Millisecond,
		TimeoutPrecommit:            1000 * time.Millisecond,
		TimeoutPrecommitDelta:       500 * time.Millisecond,
		TimeoutCommit:               1000 * time.Millisecond,
		SkipTimeoutCommit:           false,
		CreateEmptyBlocks:           false,
		CreateEmptyBlocksInterval:   0 * time.Second,
		PeerGossipSleepDuration:     100 * time.Millisecond,
		PeerQueryMaj23SleepDuration: 2000 * time.Millisecond,
		// DoubleSignCheckHeight:       int64(0),
	}
	return c
}
// Commit returns the amount of time to wait for straggler votes after receiving +2/3 precommits
// for a single block (ie. a commit).
func (cfg *ConsensusConfiguration) Commit(t time.Time) time.Time {
	return t.Add(cfg.TimeoutCommit)
}
// Precommit returns the amount of time to wait for straggler votes after receiving any +2/3 precommits
func (cfg *ConsensusConfiguration) Precommit(round int32) time.Duration {
	return time.Duration(
		cfg.TimeoutPrecommit.Nanoseconds()+cfg.TimeoutPrecommitDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}
// Prevote returns the amount of time to wait for straggler votes after receiving any +2/3 prevotes
func (cfg *ConsensusConfiguration) Prevote(round int32) time.Duration {
	return time.Duration(
		cfg.TimeoutPrevote.Nanoseconds()+cfg.TimeoutPrevoteDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// WaitForTxs returns true if the consensus should wait for transactions before entering the propose step
func (cfg *ConsensusConfiguration) WaitForTxs() bool {
	return !cfg.CreateEmptyBlocks || cfg.CreateEmptyBlocksInterval > 0
}
// Propose returns the amount of time to wait for a proposal
func (cfg *ConsensusConfiguration) Propose(round int32) time.Duration {
	return time.Duration(
		cfg.TimeoutPropose.Nanoseconds()+cfg.TimeoutProposeDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

func(this ConsensusConfiguration)WalFile()string{
	if this.walFile != "" {
		return this.walFile
	}
	return rootify(this.WalPath, this.RootDir)
}
