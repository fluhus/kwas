package kmr

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
)

func TestPrabCountAdd(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 321}
	a.Add(b)
	want := uint64(444)
	if a.Count != want {
		t.Fatalf("Add(...)=%v, want %v", a.Count, want)
	}
}

func TestPrabCountAdd_bad(t *testing.T) {
	defer func() { recover() }()
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3}, 321}
	a.Add(b)
	t.Fatalf("Add(...) succeeded, want fail")
}

func TestPrabCountCopy(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	c := b.Copy().(*HasCount)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Copy() changed receiver %v, want %v", b, a)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("Copy()=%v, want %v", c, a)
	}
}

func TestPrabCountEncode(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	c := &HasCount{}
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
		Samples: []int{0, 1, 5, 7, 8},
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
	a := []int{5, 10, 13, 27, 100}
	b := []uint64{5, 5, 3, 14, 73}

	if got := toDiffs(slices.Clone(a)); !slices.Equal(b, got) {
		t.Fatalf("toDiffs(%v)=%v, want %v", a, got, b)
	}
	if got := fromDiffs(b); !slices.Equal(got, a) {
		t.Fatalf("fromDiffs(%v)=%v, want %v", b, got, a)
	}
}
