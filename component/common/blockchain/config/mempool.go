/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/10 4:56 下午
# @File : mempool.go
# @Description :
# @Attention :
*/
package config

type MemPoolConfiguration struct {
	RootDir string `mapstructure:"home"`

	Size        int   `mapstructure:"size"`
	MaxTxsBytes int64 `mapstructure:"max-txs-bytes"`
	Recheck     bool  `mapstructure:"recheck"`
	CacheSize   int   `mapstructure:"cache-size"`

	WalPath string `mapstructure:"wal-dir"`
}

func NewDefaultMemPoolConfiguration() *MemPoolConfiguration {
	return &MemPoolConfiguration{
		RootDir:     "/Users/joker/Desktop/bft/",
		Size:        int(DEFAULT_CONSTANT),
		MaxTxsBytes: DEFAULT_CONSTANT,
		Recheck:     true,
		CacheSize:   int(DEFAULT_CONSTANT),
		WalPath:     "tx",
	}
}

// WalDir returns the full path to the mempool's write-ahead log
func (cfg *MemPoolConfiguration) WalDir() string {
	return rootify(cfg.WalPath, cfg.RootDir)
}
