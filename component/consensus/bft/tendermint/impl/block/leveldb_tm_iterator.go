/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/2/26 10:18 上午
# @File : leveldb_tm_go
# @Description :
# @Attention :
*/
package block

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type goLevelDBIterator struct {
	source    iterator.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}


func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

func newGoLevelDBIterator(source iterator.Iterator, start, end []byte, isReverse bool) *goLevelDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		} else {
			valid := source.Seek(end)
			if valid {
				eoakey := source.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.Last()
			}
		}
	} else {
		if start == nil {
			source.First()
		} else {
			source.Seek(start)
		}
	}
	return &goLevelDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Domain implements 
func (itr *goLevelDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Valid implements 
func (itr *goLevelDBIterator) Valid() bool {

	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// If source errors, invalid.
	if err := itr.Error(); err != nil {
		itr.isInvalid = true
		return false
	}

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	var start = itr.start
	var end = itr.end
	var key = itr.source.Key()

	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

// Key implements 
func (itr *goLevelDBIterator) Key() []byte {
	// Key returns a copy of the current key.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Key())
}

// Value implements 
func (itr *goLevelDBIterator) Value() []byte {
	// Value returns a copy of the current value.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Value())
}

// Next implements 
func (itr *goLevelDBIterator) Next() {
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Error implements 
func (itr *goLevelDBIterator) Error() error {
	return itr.source.Error()
}

// Close implements 
func (itr *goLevelDBIterator) Close() error {
	itr.source.Release()
	return nil
}

func (itr goLevelDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
