// Package gmerge provides merging of generic sorted streams.
package gmerge

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/heaps"
)

type item[S any, T any] struct {
	s S
	t T
}

// Merger merges sorted streams (S) that yield T's.
type Merger[S any, T any] struct {
	h       *heaps.Heap[item[S, T]]
	next    func(S) (T, error)
	merge   func(T, T) T
	compare func(T, T) int
}

// NewMerger returns a merger of instances of type T and streams of type S.
// Next should return the next element from S, EOF if no more elements are
// available, or any other non-nil error which will be propagated back.
// Merge should take two T's and return a merged T, possibly consuming the
// inputs.
// Compare should return 0 if t1 and t2 are key-equal and should be merged,
// <0 if t1 is key-lesser than t2, or >0 if t1 is key-greater than t2.
func NewMerger[S any, T any](
	next func(S) (T, error),
	merge func(T, T) T,
	compare func(T, T) int) *Merger[S, T] {
	return &Merger[S, T]{
		heaps.New(func(a, b item[S, T]) bool {
			return compare(a.t, b.t) < 0
		}), next, merge, compare,
	}
}

// Add adds s to the merger. Calls next once and returns the error.
func (m *Merger[S, T]) Add(s S) error {
	t, err := m.next(s)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	m.h.Push(item[S, T]{s, t})
	return nil
}

// Next returns a merged T from the next group of key-equal elements.
func (m *Merger[S, T]) Next() (T, error) {
	if m.h.Len() == 0 {
		var t T
		return t, io.EOF
	}
	cur := m.h.Head().t
	for {
		if err := m.advance(); err != nil {
			var t T
			return t, err
		}
		if m.h.Len() == 0 {
			break
		}
		t := m.h.Head().t
		cmp := m.compare(cur, t)
		if cmp == 0 {
			cur = m.merge(cur, t)
		} else if cmp < 0 {
			break
		} else { // cmp > 0  ==>  t < cur
			var tt T
			return tt, fmt.Errorf("next value %v is less than current %v",
				t, cur)
		}
	}
	return cur, nil
}

// Advances the head iterator in the heap, possibly discarding it if out of
// elements.
func (m *Merger[S, T]) advance() error {
	min := m.h.Pop()
	t, err := m.next(min.s)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	min.t = t
	m.h.Push(min)
	return nil
}
