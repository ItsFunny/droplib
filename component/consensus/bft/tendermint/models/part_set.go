/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:09 上午 
# @File : part_set.go
# @Description : 
# @Attention : 
*/
package models

import (
	"bytes"
	"errors"
	"github.com/hyperledger/fabric-droplib/libs"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"io"
	"sync"
)

var (
	ErrPartSetUnexpectedIndex = errors.New("error part set unexpected index")
	ErrPartSetInvalidProof    = errors.New("error part set invalid proof")
)

const (
	// MaxBlockSizeBytes is the maximum permitted size of the blocks.
	// MaxBlockSizeBytes = 104857600 // 100MB
	MaxBlockSizeBytes = 104857600 // 100MB

	// BlockPartSizeBytes is the size of one block part.
	BlockPartSizeBytes uint32 = 65536 // 64kB

	// MaxBlockPartsCount is the maximum number of block parts.
	MaxBlockPartsCount = (MaxBlockSizeBytes / BlockPartSizeBytes) + 1

)
type PartSet struct {
	total uint32
	hash  []byte

	mtx           sync.Mutex
	parts         []*Part
	partsBitArray *libs.BitArray
	count         uint32
	// a count of the total size (in bytes). Used to ensure that the
	// part set doesn't exceed the maximum block bytes
	byteSize int64
}

func (ps *PartSet) IsComplete() bool {
	return ps.count == ps.total
}

func (ps *PartSet) Total() uint32 {
	if ps == nil {
		return 0
	}
	return ps.total
}
func (ps *PartSet) GetReader() io.Reader {
	if !ps.IsComplete() {
		panic("Cannot GetReader() on incomplete PartSet")
	}
	return NewPartSetReader(ps.parts)
}
type PartSetReader struct {
	i      int
	parts  []*Part
	reader *bytes.Reader
}
func NewPartSetReader(parts []*Part) *PartSetReader {
	return &PartSetReader{
		i:      0,
		parts:  parts,
		reader: bytes.NewReader(parts[0].Bytes),
	}
}

func (psr *PartSetReader) Read(p []byte) (n int, err error) {
	readerLen := psr.reader.Len()
	if readerLen >= len(p) {
		return psr.reader.Read(p)
	} else if readerLen > 0 {
		n1, err := psr.Read(p[:readerLen])
		if err != nil {
			return n1, err
		}
		n2, err := psr.Read(p[readerLen:])
		return n1 + n2, err
	}

	psr.i++
	if psr.i >= len(psr.parts) {
		return 0, io.EOF
	}
	psr.reader = bytes.NewReader(psr.parts[psr.i].Bytes)
	return psr.Read(p)
}
func (ps *PartSet) ByteSize() int64 {
	if ps == nil {
		return 0
	}
	return ps.byteSize
}
func (ps *PartSet) AddPart(part *Part) (bool, error) {
	if ps == nil {
		return false, nil
	}
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// Invalid part index
	if part.Index >= ps.total {
		return false, ErrPartSetUnexpectedIndex
	}

	// If part already exists, return false.
	if ps.parts[part.Index] != nil {
		return false, nil
	}

	// Check hash proof
	// if part.Proof.Verify(ps.Hash(), part.Bytes) != nil {
	// 	return false, ErrPartSetInvalidProof
	// }

	// Add part
	ps.parts[part.Index] = part
	ps.partsBitArray.SetIndex(int(part.Index), true)
	ps.count++
	ps.byteSize += int64(len(part.Bytes))
	return true, nil
}
func (part *Part) ToProto() (*types.PartProto, error) {
	if part == nil {
		return nil, errors.New("nil part")
	}
	pb := new(types.PartProto)
	proof := part.Proof.ToProto()

	pb.Index = part.Index
	pb.Bytes = part.Bytes
	pb.Proof = proof

	return pb, nil
}

func (ps *PartSet) GetPart(index int) *Part {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.parts[index]
}
func (ps *PartSet) HasHeader(header PartSetHeader) bool {
	if ps == nil {
		return false
	}
	return ps.Header().Equals(header)
}
func (ps *PartSet) BitArray() *libs.BitArray {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.partsBitArray.Copy()
}
func (ps *PartSet) Header() PartSetHeader {
	if ps == nil {
		return PartSetHeader{}
	}
	return PartSetHeader{
		Total: ps.total,
		Hash:  ps.hash,
	}
}