/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 12:56 下午
# @File : evidence.go
# @Description :
# @Attention :
*/
package commonerrors

import (
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/transfer"
)

//-------------------------------------------- ERRORS --------------------------------------

// ErrInvalidEvidence wraps a piece of evidence and the error denoting how or why it is invalid.
type ErrInvalidEvidence struct {
	Evidence transfer.Evidence
	Reason   error
}

// NewErrInvalidEvidence returns a new EvidenceInvalid with the given err.
func NewErrInvalidEvidence(ev transfer.Evidence, err error) *ErrInvalidEvidence {
	return &ErrInvalidEvidence{ev, err}
}

// Error returns a string representation of the error.
func (err *ErrInvalidEvidence) Error() string {
	return fmt.Sprintf("Invalid evidence: %v. Evidence: %v", err.Reason, err.Evidence)
}

