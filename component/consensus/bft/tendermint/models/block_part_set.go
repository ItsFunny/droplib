/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 9:23 上午
# @File : block_part_set.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	"github.com/hyperledger/fabric-droplib/libs"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)

type PartSetHeader struct {
	Total uint32        `json:"total"`
	Hash  libs.HexBytes `json:"hash"`
}
func NewPartSetFromHeader(header PartSetHeader) *PartSet {
	return &PartSet{
		total:         header.Total,
		hash:          header.Hash,
		parts:         make([]*Part, header.Total),
		partsBitArray: libs.NewBitArray(int(header.Total)),
		count:         0,
		byteSize:      0,
	}
}

func (psh PartSetHeader) IsZero() bool {
	return psh.Total == 0 && len(psh.Hash) == 0
}

func (psh PartSetHeader) Equals(other PartSetHeader) bool {
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash)
}

// ToProto converts PartSetHeader to protobuf
func (psh *PartSetHeader) ToProto() *types3.PartSetHeaderProto {
	if psh == nil {
		return &types3.PartSetHeaderProto{}
	}

	return &types3.PartSetHeaderProto{
		Total: psh.Total,
		Hash:  psh.Hash,
	}
}

// ValidateBasic performs basic validation.
func (psh PartSetHeader) ValidateBasic() error {
	// Hash can be empty in case of POLBlockID.PartSetHeader in Proposal.
	if err := ValidateHash(psh.Hash); err != nil {
		return fmt.Errorf("wrong Hash: %w", err)
	}
	return nil
}

// ValidateHash returns an error if the hash is not empty, but its
// size != tmhash.Size.
func ValidateHash(h []byte) error {
	if len(h) > 0 && len(h) != sha256.Size {
		return fmt.Errorf("expected size to be %d bytes, got %d bytes",
			sha256.Size,
			len(h),
		)
	}
	return nil
}


type Part struct {
	Index uint32        `json:"index"`
	Bytes libs.HexBytes `json:"bytes"`
	Proof merkle.Proof  `json:"proof"`
}
func(part *Part)ToProtoPart()*types3.PartProto{
	return &types3.PartProto{
		Index:                part.Index,
		Bytes:                part.Bytes,
		Proof:                part.Proof.ToProto(),
	}
}
func (part *Part) ValidateBasic() error {
	if len(part.Bytes) > int(BlockPartSizeBytes) {
		return fmt.Errorf("too big: %d bytes, max: %d", len(part.Bytes), BlockPartSizeBytes)
	}
	if err := part.Proof.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong Proof: %w", err)
	}
	return nil
}