package kmr

import (
	"bytes"
	"encoding/gob"
	"io"
	"reflect"
	"testing"

	"github.com/fluhus/kwas/aio"
	"golang.org/x/exp/slices"
)

func TestCountMapEncodeSingle(t *testing.T) {
	cm := CountMap{
		{0: 0}: {1, 2, 3, 4},
		{0: 1}: {5, 6, 7, 8},
	}
	buf := &aio.Buffer{}
	if err := cm.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := CountMap{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !reflect.DeepEqual(got, cm) {
		t.Fatalf("Decode()=%v, want %v", got, cm)
	}
}

func TestCountMapEncodeAdd(t *testing.T) {
	cm1 := CountMap{
		{0: 0}: {1, 2, 3, 4},
		{0: 1}: {5, 6, 7, 8},
	}
	buf := &aio.Buffer{}
	if err := cm1.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := CountMap{
		{0: 1}: {5, 4, 3, 2},
		{0: 2}: {13, 13, 13, 15},
	}
	want := CountMap{
		{0: 0}: {1, 2, 3, 4},
		{0: 1}: {10, 10, 10, 10},
		{0: 2}: {13, 13, 13, 15},
	}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode()=%v, want %v", got, want)
	}
}

func TestMajAlMap(t *testing.T) {
	m := MajAlMap{
		Kmer{0}: 1,
		Kmer{1}: 3,
		Kmer{2}: 2,
	}
	want := MajAlMap{
		Kmer{0}: 1,
		Kmer{1}: 3,
		Kmer{2}: 2,
	}
	buf := &aio.Buffer{}
	if err := m.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := MajAlMap{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode()=%v, want %v", got, want)
	}
}

func TestMAFTupleEncode(t *testing.T) {
	tup := &MAFTuple{
		Kmer: Kmer{7},
		MAF: []SampleMAFTuple{
			{2, 1},
			{3, 2},
			{1, 3},
		},
	}
	want := &MAFTuple{
		Kmer: Kmer{7},
		MAF: []SampleMAFTuple{
			{1, 3},
			{2, 1},
			{3, 2},
		},
	}

	buf := &aio.Buffer{}
	if err := tup.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}
	got := &MAFTuple{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode()=%v, want %v", got, want)
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

	buf := &aio.Buffer{}
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

func BenchmarkEncodeCountTuple(b *testing.B) {
	b.Run("gob", func(b *testing.B) {
		enc := gob.NewEncoder(io.Discard)
		ct := CountTuple{Kmer: Kmer{3}, Count: Counts{1, 2, 3, 4}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enc.Encode(ct)
		}
	})
	b.Run("aio", func(b *testing.B) {
		var w aio.Discard
		ct := CountTuple{Kmer: Kmer{3}, Count: Counts{1, 2, 3, 4}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ct.Encode(w)
		}
	})
}

func BenchmarkDecodeCountTuple(b *testing.B) {
	b.Run("gob", func(b *testing.B) {
		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)
		ct := CountTuple{Kmer: Kmer{3}, Count: Counts{1, 2, 3, 4}}
		for i := 0; i < b.N; i++ {
			enc.Encode(ct)
		}
		dec := gob.NewDecoder(buf)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := dec.Decode(&ct); err != nil {
				b.Fatalf("Decode() failed: %v", err)
			}
		}
	})
	b.Run("aio", func(b *testing.B) {
		buf := &aio.Buffer{}
		ct := CountTuple{Kmer: Kmer{3}, Count: Counts{1, 2, 3, 4}}
		for i := 0; i < b.N; i++ {
			aio.WriteBytes(buf, ct.Kmer[:])
			for _, c := range ct.Count {
				aio.WriteUvarint(buf, uint64(c))
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ct.Decode(buf)
		}
	})
}
