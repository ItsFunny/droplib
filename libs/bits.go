/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/2 11:44 上午
# @File : bits.go
# @Description :
# @Attention :
*/
package libs

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

// BitArray is a thread-safe implementation of a bit array.
type BitArray struct {
	mtx   sync.Mutex
	Bits  int      `json:"bits"`  // NOTE: persisted via reflect, must be exported
	Elems []uint64 `json:"elems"` // NOTE: persisted via reflect, must be exported
}

// Size returns the number of bits in the bitarray
func (bA *BitArray) Size() int {
	if bA == nil {
		return 0
	}
	return bA.Bits
}


// Copy returns a copy of the provided bit array.
func (bA *BitArray) Copy() *BitArray {
	if bA == nil {
		return nil
	}
	bA.mtx.Lock()
	defer bA.mtx.Unlock()
	return bA.copy()
}

func (bA *BitArray) copy() *BitArray {
	c := make([]uint64, len(bA.Elems))
	copy(c, bA.Elems)
	return &BitArray{
		Bits:  bA.Bits,
		Elems: c,
	}
}




// Not returns a bit array resulting from a bitwise Not of the provided bit array.
func (bA *BitArray) Not() *BitArray {
	if bA == nil {
		return nil // Degenerate
	}
	bA.mtx.Lock()
	defer bA.mtx.Unlock()
	return bA.not()
}
func (bA *BitArray) not() *BitArray {
	c := bA.copy()
	for i := 0; i < len(c.Elems); i++ {
		c.Elems[i] = ^c.Elems[i]
	}
	return c
}
// NewBitArray returns a new bit array.
// It returns nil if the number of bits is zero.
func NewBitArray(bits int) *BitArray {
	if bits <= 0 {
		return nil
	}
	return &BitArray{
		Bits:  bits,
		Elems: make([]uint64, numElems(bits)),
	}
}

func numElems(bits int) int {
	return (bits + 63) / 64
}

// SetIndex sets the bit at index i within the bit array.
// The behavior is undefined if i >= bA.Bits
func (bA *BitArray) SetIndex(i int, v bool) bool {
	if bA == nil {
		return false
	}
	bA.mtx.Lock()
	defer bA.mtx.Unlock()
	return bA.setIndex(i, v)
}

func (bA *BitArray) setIndex(i int, v bool) bool {
	if i >= bA.Bits {
		return false
	}
	if v {
		bA.Elems[i/64] |= (uint64(1) << uint(i%64))
	} else {
		bA.Elems[i/64] &= ^(uint64(1) << uint(i%64))
	}
	return true
}

// Sub subtracts the two bit-arrays bitwise, without carrying the bits.
// Note that carryless subtraction of a - b is (a and not b).
// The output is the same as bA, regardless of o's size.
// If bA is longer than o, o is right padded with zeroes
func (bA *BitArray) Sub(o *BitArray) *BitArray {
	if bA == nil || o == nil {
		// TODO: Decide if we should do 1's complement here?
		return nil
	}
	bA.mtx.Lock()
	o.mtx.Lock()
	// output is the same size as bA
	c := bA.copyBits(bA.Bits)
	// Only iterate to the minimum size between the two.
	// If o is longer, those bits are ignored.
	// If bA is longer, then skipping those iterations is equivalent
	// to right padding with 0's
	smaller := MinInt(len(bA.Elems), len(o.Elems))
	for i := 0; i < smaller; i++ {
		// &^ is and not in golang
		c.Elems[i] &^= o.Elems[i]
	}
	bA.mtx.Unlock()
	o.mtx.Unlock()
	return c
}


func (bA *BitArray) copyBits(bits int) *BitArray {
	c := make([]uint64, numElems(bits))
	copy(c, bA.Elems)
	return &BitArray{
		Bits:  bits,
		Elems: c,
	}
}


// PickRandom returns a random index for a set bit in the bit array.
// If there is no such value, it returns 0, false.
// It uses the global randomness in `random.go` to get this index.
func (bA *BitArray) PickRandom() (int, bool) {
	if bA == nil {
		return 0, false
	}

	bA.mtx.Lock()
	trueIndices := bA.getTrueIndices()
	bA.mtx.Unlock()

	if len(trueIndices) == 0 { // no bits set to true
		return 0, false
	}

	return trueIndices[rand.Intn(len(trueIndices))], true
}


func (bA *BitArray) getTrueIndices() []int {
	trueIndices := make([]int, 0, bA.Bits)
	curBit := 0
	numElems := len(bA.Elems)
	// set all true indices
	for i := 0; i < numElems-1; i++ {
		elem := bA.Elems[i]
		if elem == 0 {
			curBit += 64
			continue
		}
		for j := 0; j < 64; j++ {
			if (elem & (uint64(1) << uint64(j))) > 0 {
				trueIndices = append(trueIndices, curBit)
			}
			curBit++
		}
	}
	// handle last element
	lastElem := bA.Elems[numElems-1]
	numFinalBits := bA.Bits - curBit
	for i := 0; i < numFinalBits; i++ {
		if (lastElem & (uint64(1) << uint64(i))) > 0 {
			trueIndices = append(trueIndices, curBit)
		}
		curBit++
	}
	return trueIndices
}


// String returns a string representation of BitArray: BA{<bit-string>},
// where <bit-string> is a sequence of 'x' (1) and '_' (0).
// The <bit-string> includes spaces and newlines to help people.
// For a simple sequence of 'x' and '_' characters with no spaces or newlines,
// see the MarshalJSON() method.
// Example: "BA{_x_}" or "nil-BitArray" for nil.
func (bA *BitArray) String() string {
	return bA.StringIndented("")
}

// StringIndented returns the same thing as String(), but applies the indent
// at every 10th bit, and twice at every 50th bit.
func (bA *BitArray) StringIndented(indent string) string {
	if bA == nil {
		return "nil-BitArray"
	}
	bA.mtx.Lock()
	defer bA.mtx.Unlock()
	return bA.stringIndented(indent)
}

func (bA *BitArray) stringIndented(indent string) string {
	lines := []string{}
	bits := ""
	for i := 0; i < bA.Bits; i++ {
		if bA.getIndex(i) {
			bits += "x"
		} else {
			bits += "_"
		}
		if i%100 == 99 {
			lines = append(lines, bits)
			bits = ""
		}
		if i%10 == 9 {
			bits += indent
		}
		if i%50 == 49 {
			bits += indent
		}
	}
	if len(bits) > 0 {
		lines = append(lines, bits)
	}
	return fmt.Sprintf("BA{%v:%v}", bA.Bits, strings.Join(lines, indent))
}


func (bA *BitArray) getIndex(i int) bool {
	if i >= bA.Bits {
		return false
	}
	return bA.Elems[i/64]&(uint64(1)<<uint(i%64)) > 0
}
// Update sets the bA's bits to be that of the other bit array.
// The copying begins from the begin of both bit arrays.
func (bA *BitArray) Update(o *BitArray) {
	if bA == nil || o == nil {
		return
	}

	bA.mtx.Lock()
	o.mtx.Lock()
	copy(bA.Elems, o.Elems)
	o.mtx.Unlock()
	bA.mtx.Unlock()
}

// Or returns a bit array resulting from a bitwise OR of the two bit arrays.
// If the two bit-arrys have different lengths, Or right-pads the smaller of the two bit-arrays with zeroes.
// Thus the size of the return value is the maximum of the two provided bit arrays.
func (bA *BitArray) Or(o *BitArray) *BitArray {
	if bA == nil && o == nil {
		return nil
	}
	if bA == nil && o != nil {
		return o.Copy()
	}
	if o == nil {
		return bA.Copy()
	}
	bA.mtx.Lock()
	o.mtx.Lock()
	c := bA.copyBits(MaxInt(bA.Bits, o.Bits))
	smaller := MinInt(len(bA.Elems), len(o.Elems))
	for i := 0; i < smaller; i++ {
		c.Elems[i] |= o.Elems[i]
	}
	bA.mtx.Unlock()
	o.mtx.Unlock()
	return c
}