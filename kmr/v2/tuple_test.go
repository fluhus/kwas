package kmr

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/slices"
)

func TestCountTuple_add(t *testing.T) {
	a := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	b := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{321}}
	a.Add(b)
	want := 444
	if a.Data.Count != want {
		t.Fatalf("Add(...)=%v, want %v", a.Data.Count, want)
	}
}

func TestCountTuple_bad(t *testing.T) {
	defer func() { recover() }()
	a := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	b := &CountTuple{Kmer: Kmer{1, 2, 3}, Data: CountData{321}}
	a.Add(b)
	t.Fatalf("Add(...) succeeded, want fail")
}

func TestCountTuple_copy(t *testing.T) {
	a := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	b := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	c := b.Clone()
	if !countTuplesEqual(a, b) {
		t.Fatalf("Copy() changed receiver %v, want %v", b, a)
	}
	if !countTuplesEqual(a, c) {
		t.Fatalf("Copy()=%v, want %v", c, a)
	}
}

func TestCountTuple_encode(t *testing.T) {
	a := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	b := &CountTuple{Kmer: Kmer{1, 2, 3, 4}, Data: CountData{123}}
	c := &CountTuple{}
	buf := bytes.NewBuffer(nil)
	fmt.Println(buf.Len())
	if err := b.Encode(bnry.NewWriter(buf)); err != nil {
		t.Fatalf("%v.Encode() failed: %v", b, err)
	}
	fmt.Println(buf.Len())
	if !countTuplesEqual(a, b) {
		t.Fatalf("Encode() changed receiver %v, want %v", b, a)
	}
	if err := c.Decode(buf); err != nil {
		t.Fatalf("%v.Decode() failed: %v", c, err)
	}
	if !countTuplesEqual(a, c) {
		t.Fatalf("Decode()=%v, want %v", c, a)
	}
}

func TestHasTuple_encode(t *testing.T) {
	tup := &HasTuple{
		Kmer: Kmer{},
		Data: HasData{
			Samples: []int{5, 8, 0, 7, 1},
		},
	}
	want := &HasTuple{
		Kmer: Kmer{},
		Data: HasData{
			Samples: []int{5, 8, 0, 7, 1},
		},
	}

	buf := &bytes.Buffer{}
	if err := tup.Encode(bnry.NewWriter(buf)); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := &HasTuple{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !hasTuplesEqual(got, want) {
		t.Fatalf("Decode()=%v, want %v", got, want)
	}
}

func TestDiffs(t *testing.T) {
	input := []int{5, 10, 13, 27, 100}
	want := []int{5, 5, 3, 14, 73}
	slice := slices.Clone(input)

	toDiffs(slice)
	if !slices.Equal(slice, want) {
		t.Fatalf("toDiffs(%v)=%v, want %v", input, slice, want)
	}
	fromDiffs(slice)
	if !slices.Equal(slice, input) {
		t.Fatalf("fromDiffs(%v)=%v, want %v", input, slice, input)
	}
}

func TestDiffs_negative(t *testing.T) {
	input := []int{13, 5, 27, 100, 10}
	want := []int{13, -8, 22, 73, -90}
	slice := slices.Clone(input)

	toDiffs(slice)
	if !slices.Equal(slice, want) {
		t.Fatalf("toDiffs(%v)=%v, want %v", input, slice, want)
	}
	fromDiffs(slice)
	if !slices.Equal(slice, input) {
		t.Fatalf("fromDiffs(%v)=%v, want %v", input, slice, input)
	}
}

func countTuplesEqual(a, b *CountTuple) bool {
	return a.Kmer == b.Kmer && a.Data.Count == b.Data.Count
}

func hasTuplesEqual(a, b *HasTuple) bool {
	return a.Kmer == b.Kmer && slices.Equal(a.Data.Samples, b.Data.Samples)
}

func TestLinearSort(t *testing.T) {
	tests := []struct {
		input []int
		want  []int
	}{
		{nil, nil}, {[]int{4, 3, 0, 1}, []int{0, 1, 3, 4}},
		{[]int{5, 3, 8, 7}, []int{3, 5, 7, 8}},
	}
	for _, test := range tests {
		linearSort(test.input)
		if !slices.Equal(test.input, test.want) {
			t.Errorf("linearSort=%v, want %v", test.input, test.want)
		}
	}
}

func BenchmarkLinearSort(b *testing.B) {
	s := snm.Slice(50000, func(i int) int { return i })
	rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	buf := make([]int, len(s))
	for _, n := range []int{1000, 2000, 3000, 4000, 5000, 10000} {
		b.Run(fmt.Sprint("lin.", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf[:n], s)
				linearSort(buf[:n])
			}
		})
		b.Run(fmt.Sprint("slices.", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf[:n], s)
				slices.Sort(buf[:n])
			}
		})
		b.Run(fmt.Sprint("sort.", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				copy(buf[:n], s)
				sort.Ints(buf[:n])
			}
		})
	}
}
