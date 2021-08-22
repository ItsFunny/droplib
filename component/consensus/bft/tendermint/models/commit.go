/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:12 上午 
# @File : commit.go
# @Description : 
# @Attention : 
*/
package models

import (
	"errors"
	"fmt"
	commonutils "github.com/hyperledger/fabric-droplib/common/utils"
	"github.com/hyperledger/fabric-droplib/libs"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	protobuflibs "github.com/hyperledger/fabric-droplib/libs/protobuf"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"strings"
	"time"
)

type Commit struct {
	// NOTE: The signatures are in order of address to preserve the bonded
	// ValidatorSet order.
	// Any peer with a block can gossip signatures by index with a peer without
	// recalculating the active ValidatorSet.
	Height     uint64       `json:"height"`
	Round      int32       `json:"round"`
	BlockID    BlockID     `json:"block_id"`
	Signatures []CommitSig `json:"signatures"`

	// Memoized in first call to corresponding method.
	// NOTE: can't memoize in constructor because constructor isn't used for
	// unmarshaling.
	hash     libs.HexBytes
	bitArray *libs.BitArray
}
func (commit *Commit) StringIndented(indent string) string {
	if commit == nil {
		return "nil-Commit"
	}
	commitSigStrings := make([]string, len(commit.Signatures))
	for i, commitSig := range commit.Signatures {
		commitSigStrings[i] = commitSig.String()
	}
	return fmt.Sprintf(`Commit{
%s  Height:     %d
%s  Round:      %d
%s  BlockID:    %v
%s  Signatures:
%s    %v
%s}#%v`,
		indent, commit.Height,
		indent, commit.Round,
		indent, commit.BlockID,
		indent,
		indent, strings.Join(commitSigStrings, "\n"+indent+"    "),
		indent, commit.hash)
}
func (commit *Commit) VoteSignBytes(chainID string, valIdx int32) []byte {
	v := commit.GetVote(valIdx).ToProto()
	return VoteSignBytes(chainID, v)
}
func (commit *Commit) GetVote(valIdx int32) *Vote {
	commitSig := commit.Signatures[valIdx]
	return &Vote{
		Type:             types3.SignedMsgType_SIGNED_MSG_TYPE_PRECOMMIT,
		Height:           commit.Height,
		Round:            commit.Round,
		BlockID:          commitSig.BlockID(commit.BlockID),
		Timestamp:        commitSig.Timestamp,
		ValidatorAddress: commitSig.ValidatorAddress,
		ValidatorIndex:   valIdx,
		Signature:        commitSig.Signature,
	}
}
func (this *Commit) ToProto() *types3.CommitProto {
	l := len(this.Signatures)
	sigs := make([]*types3.CommitSigProto, l)
	for i:=0;i<l;i++{
		sigs[i]=this.Signatures[i].ToProto()
	}
	r := &types3.CommitProto{
		Height:     int64(this.Height),
		Round:      this.Round,
		BlockId:    this.BlockID.ToProto(),
		Signatures: sigs,
	}
	return r
}
func (commit *Commit) GetHeight() uint64 {
	panic("implement me")
}

func (commit *Commit) GetRound() int32 {
	panic("implement me")
}

func (commit *Commit) Type() byte {
	panic("implement me")
}

func (commit *Commit) Size() int {
	panic("implement me")
}

func (commit *Commit) BitArray() *libs.BitArray {
	panic("implement me")
}

func (commit *Commit) GetByIndex(i int32) *Vote {
	panic("implement me")
}

func (commit *Commit) IsCommit() bool {
	panic("implement me")
}

func (commit *Commit) ValidateBasic() error {
	if commit.Height < 0 {
		return errors.New("negative Height")
	}
	if commit.Round < 0 {
		return errors.New("negative Round")
	}

	if commit.Height >= 1 {
		if commit.BlockID.IsZero() {
			return errors.New("commit cannot be for nil block")
		}

		if len(commit.Signatures) == 0 {
			return errors.New("no signatures in commit")
		}
		for i, commitSig := range commit.Signatures {
			if err := commitSig.ValidateBasic(); err != nil {
				return fmt.Errorf("wrong CommitSig #%d: %v", i, err)
			}
		}
	}
	return nil
}
func (commit *Commit) Hash() libs.HexBytes {
	if commit == nil {
		return nil
	}
	if commit.hash == nil {
		bs := make([][]byte, len(commit.Signatures))
		for i, commitSig := range commit.Signatures {
			pbcs := commitSig.ToProto()
			bz, err := protobuflibs.Marshal(pbcs)
			// bz, err := pbcs.Marshal()
			if err != nil {
				panic(err)
			}

			bs[i] = bz
		}
		// commit.hash = merkle.HashFromByteSlices(bs)
	}
	return commit.hash
}
type CommitSig struct {
	BlockIDFlag      BlockIDFlag   `json:"block_id_flag"`
	ValidatorAddress cryptolibs.Address `json:"validator_address"`
	Timestamp        time.Time     `json:"timestamp"`
	Signature        []byte        `json:"signature"`
}
func (cs CommitSig) String() string {
	return fmt.Sprintf("CommitSig{%X by %X on %v @ %s}",
		libs.Fingerprint(cs.Signature),
		libs.Fingerprint(cs.ValidatorAddress),
		cs.BlockIDFlag,
		CanonicalTime(cs.Timestamp))
}
const TimeFormat = time.RFC3339Nano
func CanonicalTime(t time.Time) string {
	// Note that sending time over amino resets it to
	// local time, we need to force UTC here, so the
	// signatures match
	return libs.Canonical(t).Format(TimeFormat)
}
func (cs CommitSig) BlockID(commitBlockID BlockID) BlockID {
	var blockID BlockID
	switch cs.BlockIDFlag {
	case BlockIDFlagAbsent:
		blockID = BlockID{}
	case BlockIDFlagCommit:
		blockID = commitBlockID
	case BlockIDFlagNil:
		blockID = BlockID{}
	default:
		panic(fmt.Sprintf("Unknown BlockIDFlag: %v", cs.BlockIDFlag))
	}
	return blockID
}
func (cs CommitSig) Absent() bool {
	return cs.BlockIDFlag == BlockIDFlagAbsent
}
func (cs CommitSig) ForBlock() bool {
	return cs.BlockIDFlag == BlockIDFlagCommit
}
func (cs *CommitSig) FromProto(csp *types3.CommitSigProto) error {

	cs.BlockIDFlag = BlockIDFlag(csp.BlockIdFlag)
	cs.ValidatorAddress =csp.ValidatorAddress
	cs.Timestamp = commonutils.ConvGoogleTime2Time(csp.Timestamp)
	cs.Signature = csp.Signature

	return cs.ValidateBasic()
}
const (
	// BlockIDFlagAbsent - no vote was received from a validator.
	BlockIDFlagAbsent BlockIDFlag = iota + 1
	// BlockIDFlagCommit - voted for the Commit.BlockID.
	BlockIDFlagCommit
	// BlockIDFlagNil - voted for nil.
	BlockIDFlagNil
)
func (cs CommitSig) ValidateBasic() error {
	switch cs.BlockIDFlag {
	case BlockIDFlagAbsent:
	case BlockIDFlagCommit:
	case BlockIDFlagNil:
	default:
		return fmt.Errorf("unknown BlockIDFlag: %v", cs.BlockIDFlag)
	}

	switch cs.BlockIDFlag {
	case BlockIDFlagAbsent:
		if len(cs.ValidatorAddress) != 0 {
			return errors.New("validator address is present")
		}
		if !cs.Timestamp.IsZero() {
			return errors.New("time is present")
		}
		if len(cs.Signature) != 0 {
			return errors.New("signature is present")
		}
	default:
		if len(cs.ValidatorAddress) != cryptolibs.AddressSize {
			return fmt.Errorf("expected ValidatorAddress size to be %d bytes, got %d bytes",
				cryptolibs.AddressSize,
				len(cs.ValidatorAddress),
			)
		}
		// NOTE: Timestamp validation is subtle and handled elsewhere.
		if len(cs.Signature) == 0 {
			return errors.New("signature is missing")
		}
		if len(cs.Signature) > MaxBlockSizeBytes {
			return fmt.Errorf("signature is too big (max: %d)", MaxBlockSizeBytes)
		}
	}

	return nil
}
func (cs *CommitSig) ToProto() *types3.CommitSigProto {
	if cs == nil {
		return nil
	}
	return &types3.CommitSigProto{
		BlockIdFlag:      types3.BlockIDFlagProto(cs.BlockIDFlag),
		ValidatorAddress: cs.ValidatorAddress,
		Timestamp:        commonutils.ConvtTime2GoogleTime(cs.Timestamp),
		Signature:        cs.Signature,
	}
}