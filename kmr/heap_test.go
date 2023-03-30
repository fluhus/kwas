package kmr

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestMergerHas(t *testing.T) {
	bufs := make([]*bytes.Buffer, 3)
	for i := range bufs {
		bufs[i] = &bytes.Buffer{}
	}

	(&HasTuple{
		Kmer:    FullKmer{1},
		Samples: []int{1, 3},
	}).Encode(bufs[0])
	(&HasTuple{
		Kmer:    FullKmer{2},
		Samples: []int{5, 8},
	}).Encode(bufs[0])
	(&HasTuple{
		Kmer:    FullKmer{6},
		Samples: []int{10, 14},
	}).Encode(bufs[0])

	(&HasTuple{
		Kmer:    FullKmer{0},
		Samples: []int{0, 4},
	}).Encode(bufs[1])
	(&HasTuple{
		Kmer:    FullKmer{1},
		Samples: []int{2, 5},
	}).Encode(bufs[1])
	(&HasTuple{
		Kmer:    FullKmer{2},
		Samples: []int{2, 9},
	}).Encode(bufs[1])
	(&HasTuple{
		Kmer:    FullKmer{4},
		Samples: []int{0, 4},
	}).Encode(bufs[1])

	(&HasTuple{
		Kmer:    FullKmer{2},
		Samples: []int{1, 4},
	}).Encode(bufs[2])
	(&HasTuple{
		Kmer:    FullKmer{4},
		Samples: []int{1, 3},
	}).Encode(bufs[2])

	want := []*HasTuple{
		{Kmer: FullKmer{0}, Samples: []int{0, 4}},
		{Kmer: FullKmer{1}, Samples: []int{1, 2, 3, 5}},
		{Kmer: FullKmer{2}, Samples: []int{1, 2, 4, 5, 8, 9}},
		{Kmer: FullKmer{4}, Samples: []int{0, 1, 3, 4}},
		{Kmer: FullKmer{6}, Samples: []int{10, 14}},
	}

	m := &Merger{}
	for i := range bufs {
		if err := m.Add(bufs[i], &HasTuple{Sort: true}); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &bytes.Buffer{}
	if err := m.Dump(out); err != nil {
		t.Fatalf("dump(...) failed: %v", err)
	}

	ct := &HasTuple{}
	for i := range want {
		if err := ct.Decode(out); err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		if !reflect.DeepEqual(ct, want[i]) {
			t.Fatalf("next()=%v, want %v", ct, want[i])
		}
	}
	if err := ct.Decode(out); err != io.EOF {
		t.Fatalf("next()=(%v, %v), want EOF", ct, err)
	}
}
