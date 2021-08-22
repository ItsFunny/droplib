/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 6:30 下午
# @File : block_store.go
# @Description :
# @Attention :
*/
package block

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

var (
	_ services.IBlockStore = (*defaultFabricBlockStoreImpl)(nil)
)

// FIXME block 存储
type defaultFabricBlockStoreImpl struct {
	*impl.BaseServiceImpl
	db services.DB
}

func NewBlockStoreImpl(db services.DB) *defaultFabricBlockStoreImpl {
	d := &defaultFabricBlockStoreImpl{
		db: db,
	}
	d.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_BLOCK_STORE, d)
	return d
}

func (bs *defaultFabricBlockStoreImpl) LoadBlock(height int64) services.IBlock {
	var blockMeta = bs.LoadBlockMeta(uint64(height))
	if blockMeta == nil {
		return nil
	}

	pbb := new(types.TendermintBlockProtoWrapper)
	buf := []byte{}
	for i := 0; i < int(blockMeta.BlockID.PartSetHeader.Total); i++ {
		part := bs.LoadBlockPart(uint64(height), i)
		// If the part is missing (e.g. since it has been deleted after we
		// loaded the block meta) we consider the whole block to be missing.
		if part == nil {
			return nil
		}
		buf = append(buf, part.Bytes...)
	}
	err := proto.Unmarshal(buf, pbb)
	if err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		panic(fmt.Sprintf("Error reading block: %v", err))
	}

	block, err := models.BlockFromProto(pbb)
	if err != nil {
		panic(fmt.Errorf("error from proto block: %w", err))
	}

	return block
}

func (d *defaultFabricBlockStoreImpl) Base() uint64 {
	iter, err := d.db.Iterator(
		blockMetaKey(1),
		blockMetaKey(1<<63-1),
	)
	if err != nil {
		panic(err)
	}
	defer iter.Close()

	if iter.Valid() {
		height, err := decodeBlockMetaKey(iter.Key())
		if err == nil {
			return uint64(height)
		}
	}
	if err := iter.Error(); err != nil {
		panic(err)
	}

	return 0
}

func (bs *defaultFabricBlockStoreImpl) LoadBlockMeta(height uint64) *models.BlockMeta {
	var pbbm = new(types.BlockMetaProto)
	bz, err := bs.db.Get(blockMetaKey(height))

	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil
	}

	err = proto.Unmarshal(bz, pbbm)
	if err != nil {
		panic(fmt.Errorf("unmarshal to tmproto.BlockMeta: %w", err))
	}

	blockMeta, err := models.BlockMetaFromProto(pbbm)
	if err != nil {
		panic(fmt.Errorf("error from proto blockMeta: %w", err))
	}

	return blockMeta
}

func (bs *defaultFabricBlockStoreImpl) Height() uint64 {
	iter, err := bs.db.ReverseIterator(
		blockMetaKey(1),
		blockMetaKey(1<<63-1),
	)
	if err != nil {
		panic(err)
	}
	defer iter.Close()

	if iter.Valid() {
		height, err := decodeBlockMetaKey(iter.Key())
		if err == nil {
			return uint64(height)
		}
	}
	if err := iter.Error(); err != nil {
		panic(err)
	}

	return 0
}

func (bs *defaultFabricBlockStoreImpl) LoadBlockPart(height uint64, index int) *models.Part {
	// return d.LoadBlock(int64(height)).ToPart()
	var pbpart = new(types.PartProto)

	bz, err := bs.db.Get(blockPartKey(height, index))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	err = proto.Unmarshal(bz, pbpart)
	if err != nil {
		panic(fmt.Errorf("unmarshal to tmproto.Part failed: %w", err))
	}

	part, err := models.PartFromProto(pbpart)
	if err != nil {
		panic(fmt.Sprintf("Error reading block part: %v", err))
	}

	return part
}

func (d *defaultFabricBlockStoreImpl) LoadBlockCommit(height uint64) *models.Commit {
	var pbc = new(types.CommitProto)
	bz, err := d.db.Get(blockCommitKey(height))
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil
	}
	err = proto.Unmarshal(bz, pbc)
	if err != nil {
		panic(fmt.Errorf("error reading block commit: %w", err))
	}
	commit, err := models.CommitFromProto(pbc)
	if err != nil {
		panic(fmt.Sprintf("Error reading block commit: %v", err))
	}
	return commit
}

func (d *defaultFabricBlockStoreImpl) LoadSeenCommit(height uint64) *models.Commit {
	var pbc = new(types.CommitProto)
	bz, err := d.db.Get(seenCommitKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = proto.Unmarshal(bz, pbc)
	if err != nil {
		panic(fmt.Sprintf("error reading block seen commit: %v", err))
	}

	commit, err := models.CommitFromProto(pbc)
	if err != nil {
		panic(fmt.Errorf("error from proto commit: %w", err))
	}
	return commit
}

func (bs *defaultFabricBlockStoreImpl) SaveBlock(block *models.TendermintBlockWrapper, blockParts *models.PartSet, seenCommit *models.Commit) {
	if block == nil {
		panic("BlockStore can only save a non-nil block")
	}

	batch := bs.db.NewBatch()

	height := block.Height
	hash := block.Hash()

	if g, w := uint64(height), bs.Height()+1; bs.Base() > 0 && g != w {
		panic(fmt.Sprintf("BlockStore can only save contiguous blocks. Wanted %v, got %v", w, g))
	}
	if !blockParts.IsComplete() {
		panic("BlockStore can only save complete block part sets")
	}

	// Save block parts. This must be done before the block meta, since callers
	// typically load the block meta first as an indication that the block exists
	// and then go on to load block parts - we must make sure the block is
	// complete as soon as the block meta is written.
	for i := 0; i < int(blockParts.Total()); i++ {
		part := blockParts.GetPart(i)
		bs.saveBlockPart(uint64(height), i, part, batch)
	}
	blockMeta := models.NewBlockMeta(block, blockParts)
	pbm := blockMeta.ToProto()
	if pbm == nil {
		panic("nil blockmeta")
	}

	metaBytes := mustEncode(pbm)
	if err := batch.Set(blockMetaKey(uint64(height)), metaBytes); err != nil {
		panic(err)
	}

	if err := batch.Set(blockHashKey(hash), []byte(fmt.Sprintf("%d", height))); err != nil {
		panic(err)
	}

	pbc := block.LastCommit.ToProto()
	blockCommitBytes := mustEncode(pbc)
	if err := batch.Set(blockCommitKey(uint64(height-1)), blockCommitBytes); err != nil {
		panic(err)
	}

	// Save seen commit (seen +2/3 precommits for block)
	pbsc := seenCommit.ToProto()
	seenCommitBytes := mustEncode(pbsc)
	if err := batch.Set(seenCommitKey(uint64(height)), seenCommitBytes); err != nil {
		panic(err)
	}

	// remove the previous seen commit that we have just replaced with the
	// canonical commit
	if err := batch.Delete(seenCommitKey(uint64(height - 1))); err != nil {
		panic(err)
	}

	if err := batch.WriteSync(); err != nil {
		panic(err)
	}

	if err := batch.Close(); err != nil {
		panic(err)
	}
}

func (bs *defaultFabricBlockStoreImpl) saveBlockPart(height uint64, index int, part *models.Part, batch services.Batch) {
	pbp := part.ToProtoPart()
	// if err != nil {
	// 	panic(fmt.Errorf("unable to make part into proto: %w", err))
	// }
	partBytes := mustEncode(pbp)
	if err := batch.Set(blockPartKey(height, index), partBytes); err != nil {
		panic(err)
	}
}

func (bs *defaultFabricBlockStoreImpl) PruneBlocks(height uint64) (uint64, error) {
	if height <= 0 {
		return 0, fmt.Errorf("height must be greater than 0")
	}

	if height > bs.Height() {
		return 0, fmt.Errorf("height must be equal to or less than the latest height %d", bs.Height())
	}

	// when removing the block meta, use the hash to remove the hash key at the same time
	removeBlockHash := func(key, value []byte, batch services.Batch) error {
		// unmarshal block meta
		var pbbm = new(types.BlockMetaProto)
		err := proto.Unmarshal(value, pbbm)
		if err != nil {
			return fmt.Errorf("unmarshal to tmproto.BlockMeta: %w", err)
		}

		blockMeta, err := models.BlockMetaFromProto(pbbm)
		if err != nil {
			return fmt.Errorf("error from proto blockMeta: %w", err)
		}

		// delete the hash key corresponding to the block meta's hash
		if err := batch.Delete(blockHashKey(blockMeta.BlockID.Hash)); err != nil {
			return fmt.Errorf("failed to delete hash key: %X: %w", blockHashKey(blockMeta.BlockID.Hash), err)
		}

		return nil
	}

	// remove block meta first as this is used to indicate whether the block exists.
	// For this reason, we also use ony block meta as a measure of the amount of blocks pruned
	pruned, err := bs.pruneRange(blockMetaKey(0), blockMetaKey(height), removeBlockHash)
	if err != nil {
		return pruned, err
	}

	if _, err := bs.pruneRange(blockPartKey(0, 0), blockPartKey(height, 0), nil); err != nil {
		return pruned, err
	}

	if _, err := bs.pruneRange(blockCommitKey(0), blockCommitKey(height), nil); err != nil {
		return pruned, err
	}

	if _, err := bs.pruneRange(seenCommitKey(0), seenCommitKey(height), nil); err != nil {
		return pruned, err
	}

	return pruned, nil
}
func (bs *defaultFabricBlockStoreImpl) pruneRange(start []byte, end []byte, preDeletionHook func(key, value []byte, batch services.Batch) error, ) (uint64, error) {
	var (
		err         error
		pruned      uint64
		totalPruned uint64 = 0
	)

	batch := bs.db.NewBatch()
	defer batch.Close()

	pruned, start, err = bs.batchDelete(batch, start, end, preDeletionHook)
	if err != nil {
		return totalPruned, err
	}

	// loop until we have finished iterating over all the keys by writing, opening a new batch
	// and incrementing through the next range of keys.
	for !bytes.Equal(start, end) {
		if err := batch.Write(); err != nil {
			return totalPruned, err
		}

		totalPruned += pruned

		if err := batch.Close(); err != nil {
			return totalPruned, err
		}

		batch = bs.db.NewBatch()

		pruned, start, err = bs.batchDelete(batch, start, end, preDeletionHook)
		if err != nil {
			return totalPruned, err
		}
	}

	// once we looped over all keys we do a final flush to disk
	if err := batch.WriteSync(); err != nil {
		return totalPruned, err
	}
	totalPruned += pruned
	return totalPruned, nil
}
func (bs *defaultFabricBlockStoreImpl) batchDelete(batch services.Batch, start, end []byte, preDeletionHook func(key, value []byte, batch services.Batch) error, ) (uint64, []byte, error) {
	var pruned uint64 = 0
	iter, err := bs.db.Iterator(start, end)
	if err != nil {
		return pruned, start, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if preDeletionHook != nil {
			if err := preDeletionHook(key, iter.Value(), batch); err != nil {
				return 0, start, fmt.Errorf("pruning error at key %X: %w", iter.Key(), err)
			}
		}

		if err := batch.Delete(key); err != nil {
			return 0, start, fmt.Errorf("pruning error at key %X: %w", iter.Key(), err)
		}

		pruned++
		if pruned == 1000 {
			return pruned, iter.Key(), iter.Error()
		}
	}

	return pruned, end, iter.Error()
}
