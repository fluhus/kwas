package kmr

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
)

func TestPrabCountAdd(t *testing.T) {
	a := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	b := &CountTuple{FullKmer{1, 2, 3, 4}, 321}
	a.Add(b)
	want := uint64(444)
	if a.Count != want {
		t.Fatalf("Add(...)=%v, want %v", a.Count, want)
	}
}

func TestCountAdd_bad(t *testing.T) {
	defer func() { recover() }()
	a := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	b := &CountTuple{FullKmer{1, 2, 3}, 321}
	a.Add(b)
	t.Fatalf("Add(...) succeeded, want fail")
}

func TestCountCopy(t *testing.T) {
	a := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	b := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	c := b.Copy().(*CountTuple)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Copy() changed receiver %v, want %v", b, a)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("Copy()=%v, want %v", c, a)
	}
}

func TestCountEncode(t *testing.T) {
	a := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	b := &CountTuple{FullKmer{1, 2, 3, 4}, 123}
	c := &CountTuple{}
	buf := bytes.NewBuffer(nil)
	if err := b.Encode(buf); err != nil {
		t.Fatalf("%v.Encode() failed: %v", b, err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Encode() changed receiver %v, want %v", b, a)
	}
	if err := c.Decode(buf); err != nil {
		t.Fatalf("%v.Decode() failed: %v", c, err)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("Decode()=%v, want %v", c, a)
	}
}

func TestHasTupleEncode(t *testing.T) {
	tup := &HasTuple{
		Kmer:    FullKmer{},
		Samples: []int{5, 8, 0, 7, 1},
	}
	want := &HasTuple{
		Kmer:    FullKmer{},
		Samples: []int{5, 8, 0, 7, 1},
	}

	buf := &bytes.Buffer{}
	if err := tup.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := &HasTuple{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
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
