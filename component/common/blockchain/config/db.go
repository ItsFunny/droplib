/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/12 8:27 上午
# @File : db.go
# @Description :
# @Attention :
*/
package config

import (
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
)

const (
	// 只支持这2个
	GoLevelDBBackend BackendType = "goleveldb"
	MemDBBackend     BackendType = "memdb"
)

type BackendType string
type DbCreator func(name string, dir string) (services.DB, error)

type DBConfiguration struct {
	// Database backend: goleveldb | cleveldb | boltdb | rocksdb
	// * goleveldb (github.com/syndtr/goleveldb - most popular implementation)
	//   - pure go
	//   - stable
	// * cleveldb (uses levigo wrapper)
	//   - fast
	//   - requires gcc
	//   - use cleveldb build tag (go build -tags cleveldb)
	// * boltdb (uses etcd's fork of bolt - github.com/etcd-io/bbolt)
	//   - EXPERIMENTAL
	//   - may be faster is some use-cases (random reads - indexer)
	//   - use boltdb build tag (go build -tags boltdb)
	// * rocksdb (uses github.com/tecbot/gorocksdb)
	//   - EXPERIMENTAL
	//   - requires gcc
	//   - use rocksdb build tag (go build -tags rocksdb)
	// * badgerdb (uses github.com/dgraph-io/badger)
	//   - EXPERIMENTAL
	//   - use badgerdb build tag (go build -tags badgerdb)
	// 默认写死为level-db
	DBBackend string `mapstructure:"db-backend"`

	// Database directory
	DBPath string `mapstructure:"db-dir"`
}

func NewDefaultDBConfiguration() *DBConfiguration {
	c := &DBConfiguration{
		DBBackend: string(GoLevelDBBackend),
		DBPath:    ".",
	}
	return c
}
