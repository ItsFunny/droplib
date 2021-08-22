/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/9 1:03 下午
# @File : verify.go
# @Description :
# @Attention :
*/
package light

import (
	"bytes"
	"errors"
	"fmt"
	commonerrors "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	libmath "github.com/hyperledger/fabric-droplib/libs/math"
	"time"
)

// VerifyAdjacent verifies directly adjacent untrustedHeader against
// trustedHeader. It ensures that:
//
//  a) trustedHeader can still be trusted (if not, ErrOldHeaderExpired is returned)
//  b) untrustedHeader is valid (if not, ErrInvalidHeader is returned)
//  c) untrustedHeader.ValidatorsHash equals trustedHeader.NextValidatorsHash
//  d) more than 2/3 of new validators (untrustedVals) have signed h2
//    (otherwise, ErrInvalidHeader is returned)
//  e) headers are adjacent.
//
// maxClockDrift defines how much untrustedHeader.Time can drift into the
// future.
// trustedHeader must have a ChainID, Height, Time and NextValidatorsHash
func VerifyAdjacent(
	trustedHeader *models.SignedHeader, // height=X
	untrustedHeader *models.SignedHeader, // height=X+1
	untrustedVals *models.ValidatorSet, // height=X+1
	trustingPeriod time.Duration,
	now time.Time,
	maxClockDrift time.Duration) error {

	checkRequiredHeaderFields(trustedHeader)

	if len(trustedHeader.NextValidatorsHash) == 0 {
		panic("next validators hash in trusted header is empty")
	}

	if untrustedHeader.Height != trustedHeader.Height+1 {
		return errors.New("headers must be adjacent in height")
	}

	if HeaderExpired(trustedHeader, trustingPeriod, now) {
		return ErrOldHeaderExpired{trustedHeader.Time.Add(trustingPeriod), now}
	}

	if err := verifyNewHeaderAndVals(
		untrustedHeader, untrustedVals,
		trustedHeader,
		now, maxClockDrift); err != nil {
		return ErrInvalidHeader{err}
	}

	// Check the validator hashes are the same
	if !bytes.Equal(untrustedHeader.ValidatorsHash, trustedHeader.NextValidatorsHash) {
		err := fmt.Errorf("expected old header's next validators (%X) to match those from new header (%X)",
			trustedHeader.NextValidatorsHash,
			untrustedHeader.ValidatorsHash,
		)
		return ErrInvalidHeader{err}
	}

	// Ensure that +2/3 of new validators signed correctly.
	if err := untrustedVals.VerifyCommitLight(trustedHeader.ChainID, untrustedHeader.Commit.BlockID,
		untrustedHeader.Height, untrustedHeader.Commit); err != nil {
		return ErrInvalidHeader{err}
	}

	return nil
}
func HeaderExpired(h *models.SignedHeader, trustingPeriod time.Duration, now time.Time) bool {
	expirationTime := h.Time.Add(trustingPeriod)
	return !expirationTime.After(now)
}
func checkRequiredHeaderFields(h *models.SignedHeader) {
	if h.Height == 0 {
		panic("height in trusted header must be set (non zero")
	}

	zeroTime := time.Time{}
	if h.Time == zeroTime {
		panic("time in trusted header must be set")
	}

	if h.ChainID == "" {
		panic("chain ID in trusted header must be set")
	}
}
func verifyNewHeaderAndVals(
	untrustedHeader *models.SignedHeader,
	untrustedVals *models.ValidatorSet,
	trustedHeader *models.SignedHeader,
	now time.Time,
	maxClockDrift time.Duration) error {

	if err := untrustedHeader.ValidateBasic(trustedHeader.ChainID); err != nil {
		return fmt.Errorf("untrustedHeader.ValidateBasic failed: %w", err)
	}

	if untrustedHeader.Height <= trustedHeader.Height {
		return fmt.Errorf("expected new header height %d to be greater than one of old header %d",
			untrustedHeader.Height,
			trustedHeader.Height)
	}

	if !untrustedHeader.Time.After(trustedHeader.Time) {
		return fmt.Errorf("expected new header time %v to be after old header time %v",
			untrustedHeader.Time,
			trustedHeader.Time)
	}

	if !untrustedHeader.Time.Before(now.Add(maxClockDrift)) {
		return fmt.Errorf("new header has a time from the future %v (now: %v; max clock drift: %v)",
			untrustedHeader.Time,
			now,
			maxClockDrift)
	}

	if !bytes.Equal(untrustedHeader.ValidatorsHash, untrustedVals.Hash()) {
		return fmt.Errorf("expected new header validators (%X) to match those that were supplied (%X) at height %d",
			untrustedHeader.ValidatorsHash,
			untrustedVals.Hash(),
			untrustedHeader.Height,
		)
	}

	return nil
}


func Verify(
	trustedHeader *models.SignedHeader, // height=X
	trustedVals *models.ValidatorSet, // height=X or height=X+1
	untrustedHeader *models.SignedHeader, // height=Y
	untrustedVals *models.ValidatorSet, // height=Y
	trustingPeriod time.Duration,
	now time.Time,
	maxClockDrift time.Duration,
	trustLevel libmath.Fraction) error {

	if untrustedHeader.Height != trustedHeader.Height+1 {
		return VerifyNonAdjacent(trustedHeader, trustedVals, untrustedHeader, untrustedVals,
			trustingPeriod, now, maxClockDrift, trustLevel)
	}

	return VerifyAdjacent(trustedHeader, untrustedHeader, untrustedVals, trustingPeriod, now, maxClockDrift)
}

func VerifyNonAdjacent(
	trustedHeader *models.SignedHeader, // height=X
	trustedVals *models.ValidatorSet, // height=X or height=X+1
	untrustedHeader *models.SignedHeader, // height=Y
	untrustedVals *models.ValidatorSet, // height=Y
	trustingPeriod time.Duration,
	now time.Time,
	maxClockDrift time.Duration,
	trustLevel libmath.Fraction) error {

	checkRequiredHeaderFields(trustedHeader)

	if untrustedHeader.Height == trustedHeader.Height+1 {
		return errors.New("headers must be non adjacent in height")
	}

	if err := ValidateTrustLevel(trustLevel); err != nil {
		return err
	}

	if HeaderExpired(trustedHeader, trustingPeriod, now) {
		return ErrOldHeaderExpired{trustedHeader.Time.Add(trustingPeriod), now}
	}

	if err := verifyNewHeaderAndVals(
		untrustedHeader, untrustedVals,
		trustedHeader,
		now, maxClockDrift); err != nil {
		return ErrInvalidHeader{err}
	}

	// Ensure that +`trustLevel` (default 1/3) or more in voting power of the last trusted validator
	// set signed correctly.
	err := trustedVals.VerifyCommitLightTrusting(trustedHeader.ChainID, untrustedHeader.Commit, trustLevel)
	if err != nil {
		switch e := err.(type) {
		case commonerrors.ErrNotEnoughVotingPowerSigned:
			return ErrNewValSetCantBeTrusted{e}
		default:
			return ErrInvalidHeader{e}
		}
	}

	// Ensure that +2/3 of new validators signed correctly.
	//
	// NOTE: this should always be the last check because untrustedVals can be
	// intentionally made very large to DOS the light client. not the case for
	// VerifyAdjacent, where validator set is known in advance.
	if err := untrustedVals.VerifyCommitLight(trustedHeader.ChainID, untrustedHeader.Commit.BlockID,
		untrustedHeader.Height, untrustedHeader.Commit); err != nil {
		return ErrInvalidHeader{err}
	}

	return nil
}

func ValidateTrustLevel(lvl libmath.Fraction) error {
	if lvl.Numerator*3 < lvl.Denominator || // < 1/3
		lvl.Numerator > lvl.Denominator || // > 1
		lvl.Denominator == 0 {
		return fmt.Errorf("trustLevel must be within [1/3, 1], given %v", lvl)
	}
	return nil
}