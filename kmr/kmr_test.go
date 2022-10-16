package kmr

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
)

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

func TestKmerFromBytes(t *testing.T) {
	input := []byte("AAAACCCCGGGGTTTTGTTTTGGGGCCCCAAAA")
	want := Kmer{0b00000000, 0b01010101, 0b10101010, 0b11111111,
		0b11111111, 0b10101010, 0b01010101, 0b00000000}
	got := KmerFromBytes(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("KmerFromByte(%q)=%v, want %v", input, got, want)
	}
}

func TestKmerToBytes(t *testing.T) {
	input := Kmer{0b00000000, 0b01010101, 0b10101010, 0b11111111,
		0b11111111, 0b10101010, 0b01010101, 0b00000000}
	want := []byte("AAAACCCCGGGGTTTTATTTTGGGGCCCCAAAA")
	got := KmerToBytes(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("KmerToByte(%q)=%v, want %v", input, got, want)
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
