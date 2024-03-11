// Package gmerge provides merging of generic sorted streams.
package gmerge

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/heaps"
)

type stream[T any] struct {
	head T
	next func() (T, error)
}

// Merger merges sorted streams that yield T's.
type Merger[T any] struct {
	h       *heaps.Heap[stream[T]]
	merge   func(T, T) T
	compare func(T, T) int
}

// NewMerger returns a merger of instances of type T.
// Merge should take two T's and return a merged T, possibly consuming the
// inputs.
// Compare should return 0 if t1 and t2 are key-equal and should be merged,
// <0 if t1 is key-lesser than t2, or >0 if t1 is key-greater than t2.
func NewMerger[T any](
	compare func(T, T) int,
	merge func(T, T) T) *Merger[T] {
	return &Merger[T]{
		heaps.New(func(a, b stream[T]) bool {
			return compare(a.head, b.head) < 0
		}), merge, compare,
	}
}

// Add adds next to the merger. Calls next once and returns the error.
// Next should return the next element, EOF if no more elements are
// available, or any other non-nil error which will be propagated back.
func (m *Merger[T]) Add(next func() (T, error)) error {
	t, err := next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	m.h.Push(stream[T]{t, next})
	return nil
}

func (m *Merger[T]) AddSlice(s []T) {
	i := 0
	m.Add(func() (T, error) {
		if i >= len(s) {
			var t T
			return t, io.EOF
		}
		i++
		return s[i-1], nil
	})
}

// Next returns a merged T from the next group of key-equal elements.
func (m *Merger[T]) Next() (T, error) {
	if m.h.Len() == 0 {
		var t T
		return t, io.EOF
	}
	cur := m.h.Head().head
	for {
		if err := m.advance(); err != nil {
			var t T
			return t, err
		}
		if m.h.Len() == 0 {
			break
		}
		t := m.h.Head().head
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
func (m *Merger[T]) advance() error {
	min := m.h.Pop()
	t, err := min.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	min.head = t
	m.h.Push(min)
	return nil
}
