/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:02 上午
# @File : validator_set.go
# @Description :
# @Attention :
*/
package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/batch"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/crypto/merkle"
	commonerrors2 "github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/error"
	cryptolibs "github.com/hyperledger/fabric-droplib/libs/crypto"
	libmath "github.com/hyperledger/fabric-droplib/libs/math"
	types3 "github.com/hyperledger/fabric-droplib/protos/protobufs/types"
	"math"
	"math/big"
	"sort"
	"strings"
)

const (
	// MaxTotalVotingPower - the maximum allowed total voting power.
	// It needs to be sufficiently small to, in all cases:
	// 1. prevent clipping in incrementProposerPriority()
	// 2. let (diff+diffMax-1) not overflow in IncrementProposerPriority()
	// (Proof of 1 is tricky, left to the reader).
	// It could be higher, but this is sufficiently large for our purposes,
	// and leaves room for defensive purposes.
	MaxTotalVotingPower = int64(math.MaxInt64) / 8

	// PriorityWindowSizeFactor - is a constant that when multiplied with the
	// total voting power gives the maximum allowed distance between validator
	// priorities.
	PriorityWindowSizeFactor = 2
)
var ErrTotalVotingPowerOverflow = fmt.Errorf("total voting power of resulting valset exceeds max %d",
	MaxTotalVotingPower)

type ValidatorSet struct {
	// NOTE: persisted via reflect, must be exported.
	Validators []*Validator `json:"validators"`
	Proposer   *Validator   `json:"proposer"`

	// cached (unexported)
	totalVotingPower int64
}
func NewValidatorSet(valz []*Validator) *ValidatorSet {
	vals := &ValidatorSet{}
	err := vals.updateWithChangeSet(valz, false)
	if err != nil {
		panic(fmt.Sprintf("Cannot create validator set: %v", err))
	}
	if len(valz) > 0 {
		vals.IncrementProposerPriority(1)
	}
	return vals
}
func (vals *ValidatorSet) VerifyCommitLightTrusting(chainID string, commit *Commit, trustLevel libmath.Fraction) error {
	// sanity checks
	if trustLevel.Denominator == 0 {
		return errors.New("trustLevel has zero Denominator")
	}
	if commit == nil {
		return errors.New("nil commit")
	}

	var (
		talliedVotingPower int64
		seenVals           = make(map[int32]int, len(commit.Signatures)) // validator index -> commit index
		err                error
		cacheSignBytes     = make(map[string][]byte, len(commit.Signatures))
	)

	// Safely calculate voting power needed.
	totalVotingPowerMulByNumerator, overflow := safeMul(vals.TotalVotingPower(), int64(trustLevel.Numerator))
	if overflow {
		return errors.New("int64 overflow while calculating voting power needed. please provide smaller trustLevel numerator")
	}
	votingPowerNeeded := totalVotingPowerMulByNumerator / int64(trustLevel.Denominator)

	bv, ok := batch.CreateBatchVerifier(vals.GetProposer().PubKey)
	if ok && len(commit.Signatures) > 1 {
		for idx, commitSig := range commit.Signatures {
			// No need to verify absent or nil votes.
			if !commitSig.ForBlock() {
				continue
			}

			// We don't know the validators that committed this block, so we have to
			// check for each vote if its validator is already known.
			valIdx, val := vals.GetByAddress(commitSig.ValidatorAddress)

			if val != nil {
				// check for double vote of validator on the same commit
				if firstIndex, ok := seenVals[valIdx]; ok {
					secondIndex := idx
					return fmt.Errorf("double vote from %v (%d and %d)", val, firstIndex, secondIndex)
				}
				seenVals[valIdx] = idx

				// Validate signature.
				voteSignBytes := commit.VoteSignBytes(chainID, int32(idx))
				// cache the signed bytes in case we fail verification
				cacheSignBytes[string(val.PubKey.ToBytes())] = voteSignBytes
				// if batch verification is supported add the key, sig and message to the verifier
				if err := bv.Add(val.PubKey, voteSignBytes, commitSig.Signature); err != nil {
					return err
				}

				talliedVotingPower += val.VotingPower

				if talliedVotingPower > votingPowerNeeded {
					return nil
				}
			}
		}
		if !bv.Verify() {
			talliedVotingPower, err = verifyCommitLightTrustingSingle(
				chainID, vals, commit, votingPowerNeeded, cacheSignBytes)
			if err != nil {
				return err
			} else if talliedVotingPower > votingPowerNeeded {
				return nil
			}
		}
	} else {
		talliedVotingPower, err = verifyCommitLightTrustingSingle(
			chainID, vals, commit, votingPowerNeeded, cacheSignBytes)
		if err != nil {
			return err
		} else if talliedVotingPower > votingPowerNeeded {
			return nil
		}
	}

	return commonerrors2.ErrNotEnoughVotingPowerSigned{Got: talliedVotingPower, Needed: votingPowerNeeded}
}
func verifyCommitLightTrustingSingle(
	chainID string, vals *ValidatorSet, commit *Commit, votingPowerNeeded int64,
	cachedVals map[string][]byte) (int64, error) {
	var (
		seenVals                 = make(map[int32]int, len(commit.Signatures))
		talliedVotingPower int64 = 0
	)
	for idx, commitSig := range commit.Signatures {
		// No need to verify absent or nil votes.
		if !commitSig.ForBlock() {
			continue
		}

		var voteSignBytes []byte

		// We don't know the validators that committed this block, so we have to
		// check for each vote if its validator is already known.
		valIdx, val := vals.GetByAddress(commitSig.ValidatorAddress)

		if val != nil {
			// check for double vote of validator on the same commit
			if firstIndex, ok := seenVals[valIdx]; ok {
				secondIndex := idx
				return 0, fmt.Errorf("double vote from %v (%d and %d)", val, firstIndex, secondIndex)
			}
			seenVals[valIdx] = idx

			// Validate signature.
			// voteSignBytes := commit.VoteSignBytes(chainID, int32(idx))
			if val, ok := cachedVals[string(val.PubKey.ToBytes())]; !ok {
				voteSignBytes = commit.VoteSignBytes(chainID, int32(idx))
			} else {
				voteSignBytes = val
			}
			if !val.PubKey.VerifySignature(voteSignBytes, commitSig.Signature) {
				return 0, fmt.Errorf("wrong signature (#%d): %X", idx, commitSig.Signature)
			}

			talliedVotingPower += val.VotingPower

			if talliedVotingPower > votingPowerNeeded {
				return talliedVotingPower, nil
			}
		}
	}
	return talliedVotingPower, nil
}
func safeMul(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, false
	}

	absOfB := b
	if b < 0 {
		absOfB = -b
	}

	absOfA := a
	if a < 0 {
		absOfA = -a
	}

	if absOfA > math.MaxInt64/absOfB {
		return 0, true
	}

	return a * b, false
}
func (vals *ValidatorSet) VerifyCommitLight(chainID string, blockID BlockID,
	height int64, commit *Commit) error {
	if commit == nil {
		return errors.New("nil commit")
	}


	if vals.Size() != len(commit.Signatures) {
		return commonerrors2.NewErrInvalidCommitSignatures(vals.Size(), len(commit.Signatures))
	}

	// Validate Height and BlockID.
	if height != int64(commit.Height) {
		return commonerrors2.NewErrInvalidCommitHeight(uint64(height), commit.Height)
	}
	if !blockID.Equals(commit.BlockID) {
		return fmt.Errorf("invalid commit -- wrong block ID: want %v, got %v",
			blockID, commit.BlockID)
	}

	talliedVotingPower := int64(0)
	votingPowerNeeded := vals.TotalVotingPower() * 2 / 3
	cacheSignBytes := make(map[string][]byte, len(commit.Signatures))
	var err error

	// need to check if batch verification is supported
	// if batch is supported and the there are more than x key(s) run batch, otherwise run single.
	// if batch verification fails reset tally votes to 0 and single verify until we have 2/3+
	// check if the key supports batch verification
	bv, ok := batch.CreateBatchVerifier(vals.GetProposer().PubKey)
	if ok && len(commit.Signatures) > 1 {
		for idx, commitSig := range commit.Signatures {
			// No need to verify absent or nil votes.
			if !commitSig.ForBlock() {
				continue
			}

			// The vals and commit have a 1-to-1 correspondance.
			// This means we don't need the validator address or to do any lookup.
			val := vals.Validators[idx]
			voteSignBytes := commit.VoteSignBytes(chainID, int32(idx))
			cacheSignBytes[string(val.PubKey.ToBytes())] = voteSignBytes
			// add the key, sig and message to the verifier
			if err := bv.Add(val.PubKey, voteSignBytes, commitSig.Signature); err != nil {
				return err
			}

			talliedVotingPower += val.VotingPower

			// return as soon as +2/3 of the signatures are verified
			if talliedVotingPower > votingPowerNeeded {
				return nil
			}
		}

		if !bv.Verify() {
			// reset talliedVotingPower to verify enough signatures to meet the 2/3+ threshold
			talliedVotingPower, err = verifyCommitLightSingle(
				chainID, vals, commit, votingPowerNeeded, cacheSignBytes)
			if err != nil {
				return err
			} else if talliedVotingPower > votingPowerNeeded {
				return nil
			}
		}
	} else {
		talliedVotingPower, err = verifyCommitLightSingle(
			chainID, vals, commit, votingPowerNeeded, cacheSignBytes)
		if err != nil {
			return err
		} else if talliedVotingPower > votingPowerNeeded {
			return nil
		}
	}
	return commonerrors2.ErrNotEnoughVotingPowerSigned{Got: talliedVotingPower, Needed: votingPowerNeeded}
}
func (vals *ValidatorSet) UpdateWithChangeSet(changes []*Validator) error {
	return vals.updateWithChangeSet(changes, true)
}

func verifyUpdates(
	updates []*Validator,
	vals *ValidatorSet,
	removedPower int64,
) (tvpAfterUpdatesBeforeRemovals int64, err error) {

	delta := func(update *Validator, vals *ValidatorSet) int64 {
		_, val := vals.GetByAddress(update.Address)
		if val != nil {
			return update.VotingPower - val.VotingPower
		}
		return update.VotingPower
	}

	updatesCopy := validatorListCopy(updates)
	sort.Slice(updatesCopy, func(i, j int) bool {
		return delta(updatesCopy[i], vals) < delta(updatesCopy[j], vals)
	})

	tvpAfterRemovals := vals.TotalVotingPower() - removedPower
	for _, upd := range updatesCopy {
		tvpAfterRemovals += delta(upd, vals)
		if tvpAfterRemovals > MaxTotalVotingPower {
			return 0, ErrTotalVotingPowerOverflow
		}
	}
	return tvpAfterRemovals + removedPower, nil
}

func (vals *ValidatorSet) updateWithChangeSet(changes []*Validator, allowDeletes bool) error {
	if len(changes) == 0 {
		return nil
	}

	// Check for duplicates within changes, split in 'updates' and 'deletes' lists (sorted).
	updates, deletes, err := processChanges(changes)
	if err != nil {
		return err
	}

	if !allowDeletes && len(deletes) != 0 {
		return fmt.Errorf("cannot process validators with voting power 0: %v", deletes)
	}

	// Check that the resulting set will not be empty.
	if numNewValidators(updates, vals) == 0 && len(vals.Validators) == len(deletes) {
		return errors.New("applying the validator changes would result in empty set")
	}

	// Verify that applying the 'deletes' against 'vals' will not result in error.
	// Get the voting power that is going to be removed.
	removedVotingPower, err := verifyRemovals(deletes, vals)
	if err != nil {
		return err
	}

	// Verify that applying the 'updates' against 'vals' will not result in error.
	// Get the updated total voting power before removal. Note that this is < 2 * MaxTotalVotingPower
	tvpAfterUpdatesBeforeRemovals, err := verifyUpdates(updates, vals, removedVotingPower)
	if err != nil {
		return err
	}

	// Compute the priorities for updates.
	computeNewPriorities(updates, vals, tvpAfterUpdatesBeforeRemovals)

	// Apply updates and removals.
	vals.applyUpdates(updates)
	vals.applyRemovals(deletes)

	vals.updateTotalVotingPower() // will panic if total voting power > MaxTotalVotingPower

	// Scale and center.
	vals.RescalePriorities(PriorityWindowSizeFactor * vals.TotalVotingPower())
	vals.shiftByAvgProposerPriority()

	sort.Sort(ValidatorsByVotingPower(vals.Validators))

	return nil
}
func (vals *ValidatorSet) applyRemovals(deletes []*Validator) {
	existing := vals.Validators

	merged := make([]*Validator, len(existing)-len(deletes))
	i := 0

	// Loop over deletes until we removed all of them.
	for len(deletes) > 0 {
		if bytes.Equal(existing[0].Address, deletes[0].Address) {
			deletes = deletes[1:]
		} else { // Leave it in the resulting slice.
			merged[i] = existing[0]
			i++
		}
		existing = existing[1:]
	}

	// Add the elements which are left.
	for j := 0; j < len(existing); j++ {
		merged[i] = existing[j]
		i++
	}

	vals.Validators = merged[:i]
}
func (vals *ValidatorSet) applyUpdates(updates []*Validator) {
	existing := vals.Validators
	sort.Sort(ValidatorsByAddress(existing))

	merged := make([]*Validator, len(existing)+len(updates))
	i := 0

	for len(existing) > 0 && len(updates) > 0 {
		if bytes.Compare(existing[0].Address, updates[0].Address) < 0 { // unchanged validator
			merged[i] = existing[0]
			existing = existing[1:]
		} else {
			// Apply add or update.
			merged[i] = updates[0]
			if bytes.Equal(existing[0].Address, updates[0].Address) {
				// Validator is present in both, advance existing.
				existing = existing[1:]
			}
			updates = updates[1:]
		}
		i++
	}

	// Add the elements which are left.
	for j := 0; j < len(existing); j++ {
		merged[i] = existing[j]
		i++
	}
	// OR add updates which are left.
	for j := 0; j < len(updates); j++ {
		merged[i] = updates[j]
		i++
	}

	vals.Validators = merged[:i]
}
func computeNewPriorities(updates []*Validator, vals *ValidatorSet, updatedTotalVotingPower int64) {
	for _, valUpdate := range updates {
		address := valUpdate.Address
		_, val := vals.GetByAddress(address)
		if val == nil {
			// add val
			// Set ProposerPriority to -C*totalVotingPower (with C ~= 1.125) to make sure validators can't
			// un-bond and then re-bond to reset their (potentially previously negative) ProposerPriority to zero.
			//
			// Contract: updatedVotingPower < 2 * MaxTotalVotingPower to ensure ProposerPriority does
			// not exceed the bounds of int64.
			//
			// Compute ProposerPriority = -1.125*totalVotingPower == -(updatedVotingPower + (updatedVotingPower >> 3)).
			valUpdate.ProposerPriority = -(updatedTotalVotingPower + (updatedTotalVotingPower >> 3))
		} else {
			valUpdate.ProposerPriority = val.ProposerPriority
		}
	}

}
func verifyRemovals(deletes []*Validator, vals *ValidatorSet) (votingPower int64, err error) {
	removedVotingPower := int64(0)
	for _, valUpdate := range deletes {
		address := valUpdate.Address
		_, val := vals.GetByAddress(address)
		if val == nil {
			return removedVotingPower, fmt.Errorf("failed to find validator %X to remove", address)
		}
		removedVotingPower += val.VotingPower
	}
	if len(deletes) > len(vals.Validators) {
		panic("more deletes than validators")
	}
	return removedVotingPower, nil
}
func numNewValidators(updates []*Validator, vals *ValidatorSet) int {
	numNewValidators := 0
	for _, valUpdate := range updates {
		if !vals.HasAddress(valUpdate.Address) {
			numNewValidators++
		}
	}
	return numNewValidators
}
type ValidatorsByAddress []*Validator
func (valz ValidatorsByAddress) Len() int { return len(valz) }

func (valz ValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(valz[i].Address, valz[j].Address) == -1
}

func (valz ValidatorsByAddress) Swap(i, j int) {
	valz[i], valz[j] = valz[j], valz[i]
}

func processChanges(origChanges []*Validator) (updates, removals []*Validator, err error) {
	// Make a deep copy of the changes and sort by address.
	changes := validatorListCopy(origChanges)
	sort.Sort(ValidatorsByAddress(changes))

	removals = make([]*Validator, 0, len(changes))
	updates = make([]*Validator, 0, len(changes))
	var prevAddr cryptolibs.Address

	// Scan changes by address and append valid validators to updates or removals lists.
	for _, valUpdate := range changes {
		if bytes.Equal(valUpdate.Address, prevAddr) {
			err = fmt.Errorf("duplicate entry %v in %v", valUpdate, changes)
			return nil, nil, err
		}

		switch {
		case valUpdate.VotingPower < 0:
			err = fmt.Errorf("voting power can't be negative: %d", valUpdate.VotingPower)
			return nil, nil, err
		case valUpdate.VotingPower > MaxTotalVotingPower:
			err = fmt.Errorf("to prevent clipping/overflow, voting power can't be higher than %d, got %d",
				MaxTotalVotingPower, valUpdate.VotingPower)
			return nil, nil, err
		case valUpdate.VotingPower == 0:
			removals = append(removals, valUpdate)
		default:
			updates = append(updates, valUpdate)
		}

		prevAddr = valUpdate.Address
	}

	return updates, removals, err
}
func (vals *ValidatorSet) StringIndented(indent string) string {
	if vals == nil {
		return "nil-ValidatorSet"
	}
	var valStrings []string
	vals.Iterate(func(index int, val *Validator) bool {
		valStrings = append(valStrings, val.String())
		return false
	})
	return fmt.Sprintf(`ValidatorSet{
%s  Proposer: %v
%s  Validators:
%s    %v
%s}`,
		indent, vals.GetProposer().String(),
		indent,
		indent, strings.Join(valStrings, "\n"+indent+"    "),
		indent)

}
func (vals *ValidatorSet) Iterate(fn func(index int, val *Validator) bool) {
	for i, val := range vals.Validators {
		stop := fn(i, val.Copy())
		if stop {
			break
		}
	}
}
func (vals *ValidatorSet) ValidateBasic() error {
	if vals.IsNilOrEmpty() {
		return errors.New("validator set is nil or empty")
	}

	for idx, val := range vals.Validators {
		if err := val.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid validator #%d: %w", idx, err)
		}
	}

	if err := vals.Proposer.ValidateBasic(); err != nil {
		return fmt.Errorf("proposer failed validate basic, error: %w", err)
	}

	return nil
}
func (vals *ValidatorSet) ToProto() (*types3.ValidatorSetProto, error) {
	if vals.IsNilOrEmpty() {
		return &types3.ValidatorSetProto{}, nil // validator set should never be nil
	}

	vp := new(types3.ValidatorSetProto)
	valsProto := make([]*types3.ValidatorProto, len(vals.Validators))
	for i := 0; i < len(vals.Validators); i++ {
		valp, err := vals.Validators[i].ToProto()
		if err != nil {
			return nil, err
		}
		valsProto[i] = valp
	}
	vp.Validators = valsProto

	valProposer, err := vals.Proposer.ToProto()
	if err != nil {
		return nil, fmt.Errorf("toProto: validatorSet proposer error: %w", err)
	}
	vp.Proposer = valProposer

	vp.TotalVotingPower = vals.totalVotingPower

	return vp, nil
}
func (vals *ValidatorSet) VerifyCommit(chainID string, blockID BlockID,
	height uint64, commit *Commit) error {
	if commit == nil {
		return errors.New("nil commit")
	}
	if vals.Size() != len(commit.Signatures) {
		return commonerrors2.NewErrInvalidCommitSignatures(vals.Size(), len(commit.Signatures))
	}

	// Validate Height and BlockID.
	if height != commit.Height {
		return commonerrors2.NewErrInvalidCommitHeight(height, commit.Height)
	}
	if !blockID.Equals(commit.BlockID) {
		return fmt.Errorf("invalid commit -- wrong block ID: want %v, got %v",
			blockID, commit.BlockID)
	}

	talliedVotingPower := int64(0)
	votingPowerNeeded := vals.TotalVotingPower() * 2 / 3
	for idx, commitSig := range commit.Signatures {
		if commitSig.Absent() {
			continue // OK, some signatures can be absent.
		}

		// The vals and commit have a 1-to-1 correspondance.
		// This means we don't need the validator address or to do any lookup.
		val := vals.Validators[idx]

		// Validate signature.
		voteSignBytes := commit.VoteSignBytes(chainID, int32(idx))
		if !val.PubKey.VerifySignature(voteSignBytes, commitSig.Signature) {
			return fmt.Errorf("wrong signature (#%d): %X", idx, commitSig.Signature)
		}
		// Good!
		if commitSig.ForBlock() {
			talliedVotingPower += val.VotingPower
		}
		// else {
		// It's OK. We include stray signatures (~votes for nil) to measure
		// validator availability.
		// }
	}

	if got, needed := talliedVotingPower, votingPowerNeeded; got <= needed {
		return commonerrors2.ErrNotEnoughVotingPowerSigned{Got: got, Needed: needed}
	}

	return nil
}
func (vals *ValidatorSet) Hash() []byte {
	bzs := make([][]byte, len(vals.Validators))
	for i, val := range vals.Validators {
		bzs[i] = val.Bytes()
	}
	return merkle.HashFromByteSlices(bzs)
}
func (vals *ValidatorSet) GetProposer() (proposer *Validator) {
	if len(vals.Validators) == 0 {
		return nil
	}
	if vals.Proposer == nil {
		vals.Proposer = vals.findProposer()
	}
	return vals.Proposer.Copy()
}
func (vals *ValidatorSet) findProposer() *Validator {
	var proposer *Validator
	for _, val := range vals.Validators {
		if proposer == nil || !bytes.Equal(val.Address, proposer.Address) {
			proposer = proposer.CompareProposerPriority(val)
		}
	}
	return proposer
}
func (vals *ValidatorSet) GetByAddress(address []byte) (index int32, val *Validator) {
	for idx, val := range vals.Validators {
		if bytes.Equal(val.Address, address) {
			return int32(idx), val.Copy()
		}
	}
	return -1, nil
}
func (vals *ValidatorSet) HasAddress(address []byte) bool {
	for _, val := range vals.Validators {
		if bytes.Equal(val.Address, address) {
			return true
		}
	}
	return false
}
func (vals *ValidatorSet) IncrementProposerPriority(times int32) {
	if vals.IsNilOrEmpty() {
		panic("empty validator set")
	}
	if times <= 0 {
		panic("Cannot call IncrementProposerPriority with non-positive times")
	}

	// Cap the difference between priorities to be proportional to 2*totalPower by
	// re-normalizing priorities, i.e., rescale all priorities by multiplying with:
	//  2*totalVotingPower/(maxPriority - minPriority)
	diffMax := PriorityWindowSizeFactor * vals.TotalVotingPower()
	vals.RescalePriorities(diffMax)
	vals.shiftByAvgProposerPriority()

	var proposer *Validator
	// Call IncrementProposerPriority(1) times times.
	for i := int32(0); i < times; i++ {
		proposer = vals.incrementProposerPriority()
	}

	vals.Proposer = proposer
}
func (vals *ValidatorSet) incrementProposerPriority() *Validator {
	for _, val := range vals.Validators {
		// Check for overflow for sum.
		newPrio := safeAddClip(val.ProposerPriority, val.VotingPower)
		val.ProposerPriority = newPrio
	}
	// Decrement the validator with most ProposerPriority.
	mostest := vals.getValWithMostPriority()
	// Mind the underflow.
	mostest.ProposerPriority = safeSubClip(mostest.ProposerPriority, vals.TotalVotingPower())

	return mostest
}
func (vals *ValidatorSet) getValWithMostPriority() *Validator {
	var res *Validator
	for _, val := range vals.Validators {
		res = res.CompareProposerPriority(val)
	}
	return res
}
func (v *Validator) CompareProposerPriority(other *Validator) *Validator {
	if v == nil {
		return other
	}
	switch {
	case v.ProposerPriority > other.ProposerPriority:
		return v
	case v.ProposerPriority < other.ProposerPriority:
		return other
	default:
		result := bytes.Compare(v.Address, other.Address)
		switch {
		case result < 0:
			return v
		case result > 0:
			return other
		default:
			panic("Cannot compare identical validators")
		}
	}
}
func (vals *ValidatorSet) computeAvgProposerPriority() int64 {
	n := int64(len(vals.Validators))
	sum := big.NewInt(0)
	for _, val := range vals.Validators {
		sum.Add(sum, big.NewInt(val.ProposerPriority))
	}
	avg := sum.Div(sum, big.NewInt(n))
	if avg.IsInt64() {
		return avg.Int64()
	}

	// This should never happen: each val.ProposerPriority is in bounds of int64.
	panic(fmt.Sprintf("Cannot represent avg ProposerPriority as an int64 %v", avg))
}
func (vals *ValidatorSet) shiftByAvgProposerPriority() {
	if vals.IsNilOrEmpty() {
		panic("empty validator set")
	}
	avgProposerPriority := vals.computeAvgProposerPriority()
	for _, val := range vals.Validators {
		val.ProposerPriority = safeSubClip(val.ProposerPriority, avgProposerPriority)
	}
}
func safeSubClip(a, b int64) int64 {
	c, overflow := safeSub(a, b)
	if overflow {
		if b > 0 {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	return c
}
func (vals *ValidatorSet) RescalePriorities(diffMax int64) {
	if vals.IsNilOrEmpty() {
		panic("empty validator set")
	}
	// NOTE: This check is merely a sanity check which could be
	// removed if all tests would init. voting power appropriately;
	// i.e. diffMax should always be > 0
	if diffMax <= 0 {
		return
	}

	// Calculating ceil(diff/diffMax):
	// Re-normalization is performed by dividing by an integer for simplicity.
	// NOTE: This may make debugging priority issues easier as well.
	diff := computeMaxMinPriorityDiff(vals)
	ratio := (diff + diffMax - 1) / diffMax
	if diff > diffMax {
		for _, val := range vals.Validators {
			val.ProposerPriority /= ratio
		}
	}
}
func computeMaxMinPriorityDiff(vals *ValidatorSet) int64 {
	if vals.IsNilOrEmpty() {
		panic("empty validator set")
	}
	max := int64(math.MinInt64)
	min := int64(math.MaxInt64)
	for _, v := range vals.Validators {
		if v.ProposerPriority < min {
			min = v.ProposerPriority
		}
		if v.ProposerPriority > max {
			max = v.ProposerPriority
		}
	}
	diff := max - min
	if diff < 0 {
		return -1 * diff
	}
	return diff
}
func (vals *ValidatorSet) IsNilOrEmpty() bool {
	return vals == nil || len(vals.Validators) == 0
}
func (vals *ValidatorSet) Copy() *ValidatorSet {
	return &ValidatorSet{
		Validators:       validatorListCopy(vals.Validators),
		Proposer:         vals.Proposer,
		totalVotingPower: vals.totalVotingPower,
	}
}
func validatorListCopy(valsList []*Validator) []*Validator {
	if valsList == nil {
		return nil
	}
	valsCopy := make([]*Validator, len(valsList))
	for i, val := range valsList {
		valsCopy[i] = val.Copy()
	}
	return valsCopy
}
func (vals *ValidatorSet) TotalVotingPower() int64 {
	if vals.totalVotingPower == 0 {
		vals.updateTotalVotingPower()
	}
	return vals.totalVotingPower
}
func (vals *ValidatorSet) updateTotalVotingPower() {
	sum := int64(0)
	for _, val := range vals.Validators {
		// mind overflow
		sum = safeAddClip(sum, val.VotingPower)
		if sum > MaxTotalVotingPower {
			panic(fmt.Sprintf(
				"Total voting power should be guarded to not exceed %v; got: %v",
				MaxTotalVotingPower,
				sum))
		}
	}

	vals.totalVotingPower = sum
}

func safeAddClip(a, b int64) int64 {
	c, overflow := safeAdd(a, b)
	if overflow {
		if b < 0 {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	return c
}

func safeAdd(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b {
		return -1, true
	} else if b < 0 && a < math.MinInt64-b {
		return -1, true
	}
	return a + b, false
}

func safeSub(a, b int64) (int64, bool) {
	if b > 0 && a < math.MinInt64+b {
		return -1, true
	} else if b < 0 && a > math.MaxInt64+b {
		return -1, true
	}
	return a - b, false
}

func (vals *ValidatorSet) Size() int {
	return len(vals.Validators)
}

func (vals *ValidatorSet) GetByIndex(index int32) (address []byte, val *Validator) {
	if index < 0 || int(index) >= len(vals.Validators) {
		return nil, nil
	}
	val = vals.Validators[index]
	return val.Address, val.Copy()
}

type ValidatorsByVotingPower []*Validator

func (valz ValidatorsByVotingPower) Len() int { return len(valz) }

func (valz ValidatorsByVotingPower) Less(i, j int) bool {
	if valz[i].VotingPower == valz[j].VotingPower {
		return bytes.Compare(valz[i].Address, valz[j].Address) == -1
	}
	return valz[i].VotingPower > valz[j].VotingPower
}

func (valz ValidatorsByVotingPower) Swap(i, j int) {
	valz[i], valz[j] = valz[j], valz[i]
}
func verifyCommitLightSingle(
	chainID string, vals *ValidatorSet, commit *Commit, votingPowerNeeded int64,
	cachedVals map[string][]byte) (int64, error) {
	var talliedVotingPower int64 = 0
	for idx, commitSig := range commit.Signatures {
		// No need to verify absent or nil votes.
		if !commitSig.ForBlock() {
			continue
		}

		// The vals and commit have a 1-to-1 correspondance.
		// This means we don't need the validator address or to do any lookup.
		var voteSignBytes []byte
		val := vals.Validators[idx]

		// Check if we have the validator in the cache
		if val, ok := cachedVals[string(val.PubKey.ToBytes())]; !ok {
			voteSignBytes = commit.VoteSignBytes(chainID, int32(idx))
		} else {
			voteSignBytes = val
		}
		// Validate signature.
		if !val.PubKey.VerifySignature(voteSignBytes, commitSig.Signature) {
			return 0, fmt.Errorf("wrong signature (#%d): %X", idx, commitSig.Signature)
		}

		talliedVotingPower += val.VotingPower

		// return as soon as +2/3 of the signatures are verified
		if talliedVotingPower > votingPowerNeeded {
			return talliedVotingPower, nil
		}
	}
	return talliedVotingPower, nil
}
