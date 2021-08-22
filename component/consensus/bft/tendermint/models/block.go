/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 10:02 上午
# @File : block.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	protobuflibs "github.com/hyperledger/fabric-droplib/libs/protobuf"
	libutils "github.com/hyperledger/fabric-droplib/libs/utils"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"sync"
	"time"
)

type BlockIDFlag byte

func CommitToVoteSet(chainID string, commit *Commit, vals *ValidatorSet) *VoteSet {
	voteSet := NewVoteSet(chainID, commit.Height, commit.Round, types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT, vals)
	for idx, commitSig := range commit.Signatures {
		if commitSig.Absent() {
			continue // OK, some precommits can be missing.
		}
		added, err := voteSet.AddVote(commit.GetVote(int32(idx)))
		if !added || err != nil {
			panic(fmt.Sprintf("Failed to reconstruct LastCommit: %v", err))
		}
	}
	return voteSet
}
// BlockID
type BlockID struct {
	Hash          libs.HexBytes `json:"hash"`
	PartSetHeader PartSetHeader `json:"part_set_header"`
}
func (blockID *BlockID) ToProto() *types3.BlockIDProto {
	if blockID == nil {
		return &types3.BlockIDProto{}
	}

	return &types3.BlockIDProto{
		Hash:          blockID.Hash,
		PartSetHeader: blockID.PartSetHeader.ToProto(),
	}
}
// IsComplete returns true if this is a valid BlockID of a non-nil block.
func (blockID BlockID) IsComplete() bool {
	return len(blockID.Hash) == sha256.Size &&
		blockID.PartSetHeader.Total > 0 &&
		len(blockID.PartSetHeader.Hash) == sha256.Size
}

// Equals returns true if the BlockID matches the given BlockID
func (blockID BlockID) Equals(other BlockID) bool {
	return bytes.Equal(blockID.Hash, other.Hash) &&
		blockID.PartSetHeader.Equals(other.PartSetHeader)
}

// IsZero returns true if this is the BlockID of a nil block.
func (blockID BlockID) IsZero() bool {
	return len(blockID.Hash) == 0 &&
		blockID.PartSetHeader.IsZero()
}

func (blockID BlockID) ValidateBasic() error {
	// Hash can be empty in case of POLBlockID in Proposal.
	if err := ValidateHash(blockID.Hash); err != nil {
		return fmt.Errorf("wrong Hash")
	}
	if err := blockID.PartSetHeader.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong PartSetHeader: %v", err)
	}
	return nil
}

// Key returns a machine-readable string representation of the BlockID
func (blockID BlockID) Key() string {
	pbph := blockID.PartSetHeader.ToProto()
	bz, err := protobuflibs.Marshal(pbph)
	// bz, err := pbph.Marshal()
	if err != nil {
		panic(err)
	}

	return string(blockID.Hash) + string(bz)
}

type TendermintBlockWrapper struct {
	mtx                     sync.Mutex
	FabricBlockMetadata     [][]byte
	TendermintBlockHeader   `json:"header"`
	TendermintBlockMetadata `json:"metadata"`
	TendermintBlockData     `json:"data"`
	Evidence                EvidenceData `json:"evidence"`
	LastCommit              *Commit      `json:"last_commit"`
	// FIXME ,应该需要添加一个字段,代表是否是配置块
}
func (b *TendermintBlockWrapper) MakePartSet(partSize uint32) *PartSet {
	if b == nil {
		return nil
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()

	pbb, err := b.ToProto()
	if nil != err {
		panic(err)
	}
	bz, err := proto.Marshal(pbb)
	if err != nil {
		panic(err)
	}
	return NewPartSetFromData(bz, partSize)
}
func (this *TendermintBlockWrapper) ToProto() (*types3.TendermintBlockProtoWrapper, error) {
	listProto, err := this.Evidence.ToEvidenceListProto()
	if err != nil {
		return nil, err
	}
	r := &types3.TendermintBlockProtoWrapper{
		Header: this.TendermintBlockHeader.ToProtoHeader(),
		Data:   this.TendermintBlockData.ToProto(),
		// BlockMetadata: &common.BlockMetadata{
		// 	Metadata: this.MetaData.Metadata,
		// },
		LastCommit: this.LastCommit.ToProto(),
		Evidence:   listProto,
	}
	return r, nil
}
func NewPartSetFromData(data []byte, partSize uint32) *PartSet {
	// divide data into 4kb parts.
	total := (uint32(len(data)) + partSize - 1) / partSize
	parts := make([]*Part, total)
	partsBytes := make([][]byte, total)
	partsBitArray := libs.NewBitArray(int(total))
	for i := uint32(0); i < total; i++ {
		part := &Part{
			Index: i,
			Bytes: data[i*partSize : libs.MinInt(len(data), int((i+1)*partSize))],
		}
		parts[i] = part
		partsBytes[i] = part.Bytes
		partsBitArray.SetIndex(int(i), true)
	}
	// Compute merkle proofs
	root, proofs := merkle.ProofsFromByteSlices(partsBytes)
	for i := uint32(0); i < total; i++ {
		parts[i].Proof = *proofs[i]
	}
	return &PartSet{
		total:         total,
		hash:          root,
		parts:         parts,
		partsBitArray: partsBitArray,
		count:         total,
		byteSize:      int64(len(data)),
	}
}
func (b *TendermintBlockWrapper) HashesTo(hash []byte) bool {
	if len(hash) == 0 {
		return false
	}
	if b == nil {
		return false
	}
	return bytes.Equal(b.Hash(), hash)
}
func (b *TendermintBlockWrapper) Size() int {
	return 0
}

// Hash computes and returns the block hash.
// If the block is incomplete, block hash is nil for safety.
func (b *TendermintBlockWrapper) Hash() libs.HexBytes {
	if b == nil {
		return nil
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if b.LastCommit == nil {
		return nil
	}
	b.fillHeader()
	return b.TendermintBlockHeader.Hash()
}
func (b *TendermintBlockWrapper) fillHeader() {
	if b.LastCommitHash == nil {
		b.LastCommitHash = b.LastCommit.Hash()
	}
	if b.DataHash == nil {
		b.DataHash = b.TendermintBlockData.Hash()
	}
	if b.EvidenceHash == nil {
		b.EvidenceHash = b.Evidence.Hash()
	}

	// FIXME
	// var metadataContents [][]byte
	// for i := 0; i < len(common.BlockMetadataIndex_name); i++ {
	// 	metadataContents = append(metadataContents, []byte{})
	// }
	// b.MetaData = &common.BlockMetadata{
	// 	Metadata: metadataContents,
}

type TendermintBlockMetadata struct {
	MetaData [][]byte `json:"metaData"`
}
type TendermintBlockHeader struct {
	ChainID string    `json:"chain_id"`
	Height  int64     `json:"height"`
	Time    time.Time `json:"time"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	LastCommitHash libs.HexBytes `json:"last_commit_hash"` // commit from validators from the last block
	DataHash       libs.HexBytes `json:"data_hash"`        // transactions

	// hashes from the app output from the prev block
	ValidatorsHash     libs.HexBytes `json:"validators_hash"`      // validators for the current block
	NextValidatorsHash libs.HexBytes `json:"next_validators_hash"` // validators for the next block
	ConsensusHash      libs.HexBytes `json:"consensus_hash"`       // consensus params for current block
	// root hash of all results from the txs from the previous block
	LastResultsHash libs.HexBytes `json:"last_results_hash"`

	// consensus info
	EvidenceHash    libs.HexBytes         `json:"evidence_hash"`    // evidence included in the block
	ProposerAddress cryptolibs.Address `json:"proposer_address"` // original proposer of the block
}
func (h *TendermintBlockHeader) Populate(
	chainID string,
	timestamp time.Time, lastBlockID BlockID,
	valHash, nextValHash []byte,
	consensusHash, appHash, lastResultsHash []byte,
	proposerAddress cryptolibs.Address,
) {
	// h.Version = version
	h.ChainID = chainID
	h.Time = timestamp
	h.LastBlockID = lastBlockID
	h.ValidatorsHash = valHash
	h.NextValidatorsHash = nextValHash
	h.ConsensusHash = consensusHash
	// h.AppHash = appHash
	h.LastResultsHash = lastResultsHash
	h.ProposerAddress = proposerAddress
}

func (h *TendermintBlockHeader) StringIndented(indent string) string {
	if h == nil {
		return "nil-Header"
	}
	return fmt.Sprintf(`Header{
%s  ChainID:        %v
%s  Height:         %v
%s  Time:           %v
%s  LastBlockID:    %v
%s  LastCommit:     %v
%s  Data:           %v
%s  Validators:     %v
%s  NextValidators: %v
%s  Consensus:      %v
%s  Results:        %v
%s  Evidence:       %v
%s  Proposer:       %v
%s}#%v`,
		indent, h.ChainID,
		indent, h.Height,
		indent, h.Time,
		indent, h.LastBlockID,
		indent, h.LastCommitHash,
		indent, h.DataHash,
		indent, h.ValidatorsHash,
		indent, h.NextValidatorsHash,
		indent, h.ConsensusHash,
		indent, h.LastResultsHash,
		indent, h.EvidenceHash,
		indent, h.ProposerAddress,
		indent, h.Hash())
}
func (h *TendermintBlockHeader) ToProtoHeader() *types3.TendermintBlockHeaderProto {
	if h == nil {
		return nil
	}

	return &types3.TendermintBlockHeaderProto{
		// Version:            h.Version.ToProto(),
		ChainId:            h.ChainID,
		Height:             h.Height,
		Time:               commonutils.ConvtTime2GoogleTime(h.Time),
		LastBlockId:        h.LastBlockID.ToProto(),
		ValidatorsHash:     h.ValidatorsHash,
		NextValidatorsHash: h.NextValidatorsHash,
		ConsensusHash:      h.ConsensusHash,
		// AppHash:            h.AppHash,
		DataHash:        h.DataHash,
		EvidenceHash:    h.EvidenceHash,
		LastResultsHash: h.LastResultsHash,
		LastCommitHash:  h.LastCommitHash,
		ProposerAddress: h.ProposerAddress,
	}
}
func (sh SignedHeader) ValidateBasic(chainID string) error {
	if sh.TendermintBlockHeader == nil {
		return errors.New("missing header")
	}
	if sh.Commit == nil {
		return errors.New("missing commit")
	}

	if err := sh.TendermintBlockHeader.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid header: %w", err)
	}
	if err := sh.Commit.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid commit: %w", err)
	}

	if sh.ChainID != chainID {
		return fmt.Errorf("header belongs to another chain %q, not %q", sh.ChainID, chainID)
	}

	// Make sure the header is consistent with the commit.
	if int64(sh.Commit.Height) != sh.Height {
		return fmt.Errorf("header and commit height mismatch: %d vs %d", sh.Height, sh.Commit.Height)
	}
	if hhash, chash := sh.Hash(), sh.Commit.BlockID.Hash; !bytes.Equal(hhash, chash) {
		return fmt.Errorf("commit signs block %X, header is block %X", chash, hhash)
	}

	return nil
}
func (h TendermintBlockHeader) ValidateBasic() error {
	// if h.Version.Block != version.BlockProtocol {
	// 	return fmt.Errorf("block protocol is incorrect: got: %d, want: %d ", h.Version.Block, version.BlockProtocol)
	// }
	if len(h.ChainID) > MaxChainIDLen {
		return fmt.Errorf("chainID is too long; got: %d, max: %d", len(h.ChainID), MaxChainIDLen)
	}

	if h.Height < 0 {
		return errors.New("negative Height")
	}
	// else if h.Height == 0 {
	// 	return errors.New("zero Height")
	// }

	if err := h.LastBlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong LastBlockID: %w", err)
	}

	if err := ValidateHash(h.LastCommitHash); err != nil {
		return fmt.Errorf("wrong LastCommitHash: %v", err)
	}

	if err := ValidateHash(h.DataHash); err != nil {
		return fmt.Errorf("wrong DataHash: %v", err)
	}

	if err := ValidateHash(h.EvidenceHash); err != nil {
		return fmt.Errorf("wrong EvidenceHash: %v", err)
	}

	if len(h.ProposerAddress) != cryptolibs.AddressSize {
		return fmt.Errorf(
			"invalid ProposerAddress length; got: %d, expected: %d",
			len(h.ProposerAddress), cryptolibs.AddressSize,
		)
	}

	// Basic validation of hashes related to application data.
	// Will validate fully against state in state#ValidateBlock.
	if err := ValidateHash(h.ValidatorsHash); err != nil {
		return fmt.Errorf("wrong ValidatorsHash: %v", err)
	}
	if err := ValidateHash(h.NextValidatorsHash); err != nil {
		return fmt.Errorf("wrong NextValidatorsHash: %v", err)
	}
	if err := ValidateHash(h.ConsensusHash); err != nil {
		return fmt.Errorf("wrong ConsensusHash: %v", err)
	}
	// NOTE: AppHash is arbitrary length
	if err := ValidateHash(h.LastResultsHash); err != nil {
		return fmt.Errorf("wrong LastResultsHash: %v", err)
	}

	return nil
}


func (h *TendermintBlockHeader) Hash() libs.HexBytes {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	// hpb := h.Version.ToProto()
	// hbz, err := hpb.Marshal()
	// if err != nil {
	// 	return nil
	// }

	pbt, err := gogotypes.StdTimeMarshal(h.Time)
	if err != nil {
		return nil
	}

	pbbi := h.LastBlockID.ToProto()
	bzbi, err := proto.Marshal(pbbi)
	if err != nil {
		return nil
	}
	return merkle.HashFromByteSlices([][]byte{
		libutils.CdcEncode(h.ChainID),
		libutils.CdcEncode(h.Height),
		pbt,
		bzbi,
		libutils.CdcEncode(h.LastCommitHash),
		libutils.CdcEncode(h.DataHash),
		libutils.CdcEncode(h.ValidatorsHash),
		libutils.CdcEncode(h.NextValidatorsHash),
		libutils.CdcEncode(h.ConsensusHash),
		// cdcEncode(h.AppHash),
		libutils.CdcEncode(h.LastResultsHash),
		libutils.CdcEncode(h.EvidenceHash),
		libutils.CdcEncode(h.ProposerAddress),
	})
}

// Data contains the set of transactions included in the block
type TendermintBlockData struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs types.Txs `json:"txs"`

	// Volatile
	hash libs.HexBytes
}
func (this *TendermintBlockData) ToProto() *types3.TendermintBlockDataProto {
	l := len(this.Txs)
	datas := make([][]byte, l)
	for i := 0; i < l; i++ {
		datas[i] = this.Txs[i]
	}
	r := types3.TendermintBlockDataProto{
		Txs: datas,
	}
	return &r
}
func (data *TendermintBlockData) Hash() libs.HexBytes {
	if data == nil {
		return (types.Txs{}).Hash()
	}
	if data.hash == nil {
		data.hash = data.Txs.Hash() // NOTE: leaves of merkle tree are TxIDs
	}
	return data.hash
}

// BlockMeta contains meta information.
type BlockMeta struct {
	BlockID   BlockID               `json:"block_id"`
	BlockSize int                   `json:"block_size"`
	Header    TendermintBlockHeader `json:"header"`
	NumTxs    int                   `json:"num_txs"`
}
func (bm *BlockMeta) ValidateBasic() error {
	if err := bm.BlockID.ValidateBasic(); err != nil {
		return err
	}
	if !bytes.Equal(bm.BlockID.Hash, bm.Header.Hash()) {
		return fmt.Errorf("expected BlockID#Hash and Header#Hash to be the same, got %X != %X",
			bm.BlockID.Hash, bm.Header.Hash())
	}
	return nil
}
func (bm *BlockMeta) ToProto() *types3.BlockMetaProto {
	if bm == nil {
		return nil
	}

	pb := &types3.BlockMetaProto{
		BlockId:   bm.BlockID.ToProto(),
		BlockSize: int64(bm.BlockSize),
		Header:    bm.Header.ToProtoHeader(),
		NumTxs:    int64(bm.NumTxs),
	}
	return pb
}

// NewBlockMeta returns a new BlockMeta.
func NewBlockMeta(block *TendermintBlockWrapper, blockParts *PartSet) *BlockMeta {
	return &BlockMeta{
		BlockID:   BlockID{block.Hash(), blockParts.Header()},
		BlockSize: block.Size(),
		Header:    block.TendermintBlockHeader,
		NumTxs:    len(block.TendermintBlockData.Txs),
	}
}
