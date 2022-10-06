package gmerge

import (
	"fmt"
	"io"
	"testing"

	"golang.org/x/exp/slices"
)

func TestMerger(t *testing.T) {
	streams := [][]counter{
		{counter{2, 3}, counter{3, 1}, counter{4, 4}},
		{counter{1, 2}, counter{3, 5}},
		{counter{3, 2}, counter{4, 1}, counter{6, 10}},
	}
	m := newIterMerger()
	for _, s := range streams {
		iter := &sliceIter{s, 0}
		if err := m.Add(iter); err != nil {
			t.Fatalf("Add(%v) failed: %v", iter, err)
		}
	}
	want := []counter{
		{1, 2}, {2, 3}, {3, 8}, {4, 5}, {6, 10},
	}
	var got []counter
	for {
		c, err := m.Next()
		if err != nil {
			if err != io.EOF {
				t.Fatalf("Next() failed: %v", err)
			}
			break
		}
		got = append(got, c)
	}
	if !slices.EqualFunc(got, want, func(a, b counter) bool {
		return a.c == b.c && a.t == b.t
	}) {
		t.Fatalf("Next()=%v, want %v", got, want)
	}
}

func TestMerger_empty(t *testing.T) {
	m := newIterMerger()
	iter := &sliceIter{nil, 0}
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", iter, err)
	}
	c, err := m.Next()
	if err != io.EOF {
		t.Fatalf("Next()=%v, %v, want EOF", c, err)
	}
}

func TestMerger_badOrder(t *testing.T) {
	m := newIterMerger()
	iter := &sliceIter{[]counter{
		{2, 3}, {1, 2},
	}, 0}
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", iter, err)
	}
	c, err := m.Next()
	if err == nil || err == io.EOF {
		t.Fatalf("Next()=%v, %v, want fail", c, err)
	}
}

func TestMerger_addError(t *testing.T) {
	m := newIterMerger()
	iter := &sliceIter{[]counter{
		{2, 123}, {3, 2},
	}, 0}
	if err := m.Add(iter); err == nil {
		t.Fatalf("Add(%v) succeeded, want error", iter)
	}
}

func TestMerger_nextError(t *testing.T) {
	m := newIterMerger()
	iter := &sliceIter{[]counter{
		{2, 2}, {3, 123},
	}, 0}
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", iter, err)
	}
	c, err := m.Next()
	if err == nil || err == io.EOF {
		t.Fatalf("Next()=%v, %v, want fail", c, err)
	}
}

func newIterMerger() *Merger[*sliceIter, counter] {
	return NewMerger(
		func(si *sliceIter) (counter, error) {
			return si.next()
		}, func(c1, c2 counter) counter {
			if c1.t != c2.t {
				panic(fmt.Sprintf("mismatching keys: %d, %d", c1.t, c2.t))
			}
			return counter{c1.t, c1.c + c2.c}
		}, func(c1, c2 counter) int {
			return c1.t - c2.t
		},
	)
}

type counter struct {
	t int
	c int
}

type sliceIter struct {
	s []counter
	i int
}

func (s *sliceIter) next() (counter, error) {
	if s.i >= len(s.s) {
		return counter{}, io.EOF
	}
	if s.s[s.i].c == 123 {
		return counter{}, fmt.Errorf("test error")
	}
	s.i++
	return s.s[s.i-1], nil
}
