/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 6:06 上午
# @File : light.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/protos/protobufs/types"
)
type SignedHeader struct {
	*TendermintBlockHeader `json:"header"`

	Commit *Commit `json:"commit"`
}
func (sh SignedHeader) StringIndented(indent string) string {
	return fmt.Sprintf(`SignedHeader{
%s  %v
%s  %v
%s}`,
		indent, sh.TendermintBlockHeader.StringIndented(indent+"  "),
		indent, sh.Commit.StringIndented(indent+"  "),
		indent)
}
// ToProto converts SignedHeader to protobuf
func (sh *SignedHeader) ToProto() *types.SignedHeaderProto {
	if sh == nil {
		return nil
	}

	psh := new(types.SignedHeaderProto)
	if sh.TendermintBlockHeader != nil {
		psh.Header = sh.TendermintBlockHeader.ToProtoHeader()
	}
	if sh.Commit != nil {
		psh.Commit = sh.Commit.ToProto()
	}

	return psh
}
// LightBlock is a SignedHeader and a ValidatorSet.
// It is the basis of the light client
type LightBlock struct {
	*SignedHeader `json:"signed_header"`
	ValidatorSet  *ValidatorSet `json:"validator_set"`
}

func (lb *LightBlock) ToProto() (*types.LightBlockProto, error) {
	if lb == nil {
		return nil, nil
	}

	lbp := new(types.LightBlockProto)
	var err error
	if lb.SignedHeader != nil {
		lbp.SignedHeader = lb.SignedHeader.ToProto()
	}
	if lb.ValidatorSet != nil {
		lbp.ValidatorSet, err = lb.ValidatorSet.ToProto()
		if err != nil {
			return nil, err
		}
	}

	return lbp, nil
}

// ValidateBasic checks that the data is correct and consistent
//
// This does no verification of the signatures
func (lb LightBlock) ValidateBasic(chainID string) error {
	if lb.SignedHeader == nil {
		return errors.New("missing signed header")
	}
	if lb.ValidatorSet == nil {
		return errors.New("missing validator set")
	}

	if err := lb.SignedHeader.ValidateBasic(chainID); err != nil {
		return fmt.Errorf("invalid signed header: %w", err)
	}
	if err := lb.ValidatorSet.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid validator set: %w", err)
	}

	// make sure the validator set is consistent with the header
	if valSetHash := lb.ValidatorSet.Hash(); !bytes.Equal(lb.SignedHeader.ValidatorsHash, valSetHash) {
		return fmt.Errorf("expected validator hash of header to match validator set hash (%X != %X)",
			lb.SignedHeader.ValidatorsHash, valSetHash,
		)
	}

	return nil
}

// String returns a string representation of the LightBlock
func (lb LightBlock) String() string {
	return lb.StringIndented("")
}
func (lb LightBlock) StringIndented(indent string) string {
	return fmt.Sprintf(`LightBlock{
%s  %v
%s  %v
%s}`,
		indent, lb.SignedHeader.StringIndented(indent+"  "),
		indent, lb.ValidatorSet.StringIndented(indent+"  "),
		indent)
}