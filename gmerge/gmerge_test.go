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
	m := newCounterMerger()
	for _, s := range streams {
		iter := counterIter(s)
		if err := m.Add(iter); err != nil {
			t.Fatalf("Add(%v) failed: %v", s, err)
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
		return a.cnt == b.cnt && a.val == b.val
	}) {
		t.Fatalf("Next()=%v, want %v", got, want)
	}
}

func TestMerger_empty(t *testing.T) {
	m := newCounterMerger()
	iter := counterIter(nil)
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", nil, err)
	}
	c, err := m.Next()
	if err != io.EOF {
		t.Fatalf("Next()=%v, %v, want EOF", c, err)
	}
}

func TestMerger_badOrder(t *testing.T) {
	m := newCounterMerger()
	input := []counter{{2, 3}, {1, 2}}
	iter := counterIter(input)
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", input, err)
	}
	c, err := m.Next()
	if err == nil || err == io.EOF {
		t.Fatalf("Next()=%v, %v, want fail", c, err)
	}
}

func TestMerger_addError(t *testing.T) {
	m := newCounterMerger()
	input := []counter{{2, 123}, {3, 2}}
	iter := counterIter(input)
	if err := m.Add(iter); err == nil {
		t.Fatalf("Add(%v) succeeded, want error", input)
	}
}

func TestMerger_nextError(t *testing.T) {
	m := newCounterMerger()
	input := []counter{{2, 2}, {3, 123}}
	iter := counterIter(input)
	if err := m.Add(iter); err != nil {
		t.Fatalf("Add(%v) failed: %v", input, err)
	}
	c, err := m.Next()
	if err == nil || err == io.EOF {
		t.Fatalf("Next()=%v, %v, want fail", c, err)
	}
}

func TestMerger_mergeMergers(t *testing.T) {
	inputs := [][]counter{
		{counter{1, 2}, counter{3, 4}},
		{counter{2, 3}, counter{3, 10}, counter{5, 1}},
		{counter{2, 4}, counter{3, 7}, counter{5, 4}},
		{counter{1, 4}, counter{3, 2}, counter{5, 2}},
	}
	want := []counter{{1, 6}, {2, 7}, {3, 23}, {5, 7}}
	m1 := newCounterMerger()
	m2 := newCounterMerger()
	m3 := newCounterMerger()
	m1.Add(counterIter(inputs[0]))
	m1.Add(counterIter(inputs[1]))
	m2.Add(counterIter(inputs[2]))
	m2.Add(counterIter(inputs[3]))
	m3.Add(m1.Next)
	m3.Add(m2.Next)

	var got []counter
	var err error
	var c counter
	for c, err = m3.Next(); err == nil; c, err = m3.Next() {
		got = append(got, c)
	}
	if err != io.EOF {
		t.Fatalf("Next() failed: %v", err)
	}
	if !slices.EqualFunc(got, want, func(a, b counter) bool {
		return a.cnt == b.cnt && a.val == b.val
	}) {
		t.Fatalf("Next()=%v, want %v", got, want)
	}
}

func newCounterMerger() *Merger[counter] {
	return NewMerger(func(c1, c2 counter) int {
		return c1.val - c2.val
	},
		func(c1, c2 counter) counter {
			if c1.val != c2.val {
				panic(fmt.Sprintf("mismatching keys: %d, %d", c1.val, c2.val))
			}
			return counter{c1.val, c1.cnt + c2.cnt}
		},
	)
}

type counter struct {
	val int
	cnt int
}

func counterIter(s []counter) func() (counter, error) {
	i := 0
	return func() (counter, error) {
		if i >= len(s) {
			return counter{}, io.EOF
		}
		c := s[i]
		i++
		if c.cnt == 123 {
			return counter{}, fmt.Errorf("test error")
		}
		return c, nil
	}
}
