/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/3 5:26 下午
# @File : queue.go
# @Description :
# @Attention :
*/
package concurrentpriority

import (
	"fmt"
	"reflect"
)

type PriorityQueue struct {
	list *pqList
}

func NewPriorityQueue(cmp CmpFunc) *PriorityQueue {
	return &PriorityQueue{&pqList{cmp: cmp}}
}

type CmpFunc func(a, b interface{}) int

func (cmp CmpFunc) Merge(a, b []interface{}) []interface{} {
	na, nb := len(a), len(b)
	res := make([]interface{}, na+nb)
	for k, l, m := 0, 0, 0; l < na || m < nb; k++ {
		if m >= nb || l < na && cmp(a[l], b[m]) <= 0 {
			res[k] = a[l]
			l++
		} else {
			res[k] = b[m]
			m++
		}
	}

	return res
}

type Slice []interface{}

func (s *Slice) Add(e ...interface{}) *Slice {
	*s = append(*s, e...)
	return s
}
func (s *Slice) Insert(index int, e ...interface{}) {
	if cap(*s) >= len(*s)+len(e) {
		*s = (*s)[:len(*s)+len(e)]
	} else {
		*s = append(*s, e...)
	} // else
	copy((*s)[index+len(e):], (*s)[index:])
	copy((*s)[index:], e)
}

func (s *Slice) InsertSlice(index int, src interface{}) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("%v is not a slice!", src))
	}

	n := v.Len()
	if cap(*s) >= len(*s)+n {
		*s = (*s)[:len(*s)+n]
		copy((*s)[index+n:], (*s)[index:])
	} else {
		ss := make([]interface{}, len(*s)+n)
		copy(ss[:index], *s)
		copy(ss[index+n:], (*s)[index:])
		*s = ss
	}

	for i := 0; i < n; i++ {
		(*s)[i+index] = v.Index(i).Interface()
	}
}

func (s Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *Slice) Remove(index int) interface{} {
	e := (*s)[index]
	copy((*s)[index:], (*s)[index+1:])
	(*s)[len(*s)-1] = nil
	*s = (*s)[:len(*s)-1]
	return e
}

func (s *Slice) RemoveRange(from, to int) {
	if to <= from {
		return
	}

	copy((*s)[from:], (*s)[to:])
	n := len(*s)
	l := n - to + from
	for i := l; i < n; i++ {
		(*s)[i] = nil
	}
	*s = (*s)[:l]
}

func (s *Slice) Pop() interface{} {
	return s.Remove(len(*s) - 1)
}

func (s Slice) Fill(from, to int, vl interface{}) {
	for i := from; i < to; i++ {
		s[i] = vl
	}
}

func (s *Slice) Clear() {
	s.Fill(0, len(*s), nil)
	*s = (*s)[:0]
}

func (s Slice) Equals(t []interface{}) bool {
	if len(s) != len(t) {
		return false
	} // if

	for i := range s {
		if s[i] != t[i] {
			return false
		} // if
	} // for i

	return true
}

type pqList struct {
	Slice
	cmp CmpFunc
}

// The Push method in Interface.
func (l *pqList) Push(e interface{}) {
	l.Add(e)
}

// The Pop method in Interface.
func (l *pqList) Pop() interface{} {
	return l.Remove(len(l.Slice) - 1)
}

// The Len method in sort.Interface.
func (l *pqList) Len() int {
	return len(l.Slice)
}

// The Less method in sort.Interface
func (l *pqList) Less(i, j int) bool {
	return l.cmp(l.Slice[i], l.Slice[j]) <= 0
}

// NewPriorityQueue creates a PriorityQueue instance with a specified compare function and a capacity
func NewPriorityQueueCap(cmp CmpFunc, cap int) *PriorityQueue {
	return &PriorityQueue{&pqList{Slice: make(Slice, 0, cap), cmp: cmp}}
}

// Push inserts the specified element into this priority queue.
func (pq *PriorityQueue) Push(x interface{}) {
	Push(pq.list, x)
}

// Pop retrieves and removes the head of this queue, or returns nil if this queue is empty.
func (pq *PriorityQueue) Pop() interface{} {
	return Pop(pq.list)
}

// Peek retrieves the head of this queue, or returns nil if this queue is empty.
func (pq *PriorityQueue) Peek() interface{} {
	if pq.list.Len() > 0 {
		return pq.list.Slice[0]
	} // if

	return nil
}

// Remove removes the element at index i from the priority queue.
func (pq *PriorityQueue) Remove(i int) interface{} {
	return Remove(pq.list, i)
}

// Len returns the number of elements in this queue.
func (pq *PriorityQueue) Len() int {
	return pq.list.Len()
}

// String returns a string with value of "PriorityQueue(Len())"
func (pq *PriorityQueue) String() string {
	return fmt.Sprintf("PriorityQueue(%d)", pq.list.Len())
}
