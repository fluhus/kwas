package kmr

import (
	"bytes"
	"io"
	"testing"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/snm"
)

func TestMerger1Has(t *testing.T) {
	bufs := make([]*bytes.Buffer, 3)
	for i := range bufs {
		bufs[i] = &bytes.Buffer{}
	}

	(&HasTuple{
		Kmer: Kmer{1},
		Data: HasData{
			Samples: []int{1, 3},
		},
	}).Encode(bnry.NewWriter(bufs[0]))
	(&HasTuple{
		Kmer: Kmer{2},
		Data: HasData{
			Samples: []int{5, 8},
		},
	}).Encode(bnry.NewWriter(bufs[0]))
	(&HasTuple{
		Kmer: Kmer{6},
		Data: HasData{
			Samples: []int{10, 14},
		},
	}).Encode(bnry.NewWriter(bufs[0]))

	(&HasTuple{
		Kmer: Kmer{0},
		Data: HasData{
			Samples: []int{0, 4},
		},
	}).Encode(bnry.NewWriter(bufs[1]))
	(&HasTuple{
		Kmer: Kmer{1},
		Data: HasData{
			Samples: []int{2, 5},
		},
	}).Encode(bnry.NewWriter(bufs[1]))
	(&HasTuple{
		Kmer: Kmer{2},
		Data: HasData{
			Samples: []int{2, 9},
		},
	}).Encode(bnry.NewWriter(bufs[1]))
	(&HasTuple{
		Kmer: Kmer{4},
		Data: HasData{
			Samples: []int{0, 4},
		},
	}).Encode(bnry.NewWriter(bufs[1]))

	(&HasTuple{
		Kmer: Kmer{2},
		Data: HasData{
			Samples: []int{1, 4},
		},
	}).Encode(bnry.NewWriter(bufs[2]))
	(&HasTuple{
		Kmer: Kmer{4},
		Data: HasData{
			Samples: []int{1, 3},
		},
	}).Encode(bnry.NewWriter(bufs[2]))

	want := []*HasTuple{
		{Kmer: Kmer{0}, Data: HasData{Samples: []int{0, 4}}},
		{Kmer: Kmer{1}, Data: HasData{Samples: []int{1, 2, 3, 5}}},
		{Kmer: Kmer{2}, Data: HasData{Samples: []int{1, 2, 4, 5, 8, 9}}},
		{Kmer: Kmer{4}, Data: HasData{Samples: []int{0, 1, 3, 4}}},
		{Kmer: Kmer{6}, Data: HasData{Samples: []int{10, 14}}},
	}

	m := NewMerger[HasHandler]()
	for i := range bufs {
		if err := m.Add(IterTuplesReader[HasHandler](bufs[i])); err != nil {
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
		if !hasTuplesEqual(ct, want[i]) {
			t.Fatalf("next()=%v, want %v", ct, want[i])
		}
	}
	if err := ct.Decode(out); err != io.EOF {
		t.Fatalf("next()=(%v, %v), want EOF", ct, err)
	}
}

func TestMerger1Count(t *testing.T) {
	bufs := snm.Slice(3, func(i int) *bytes.Buffer { return &bytes.Buffer{} })
	ws := snm.Slice(3, func(i int) *bnry.Writer { return bnry.NewWriter(bufs[i]) })

	(&CountTuple{
		Kmer: Kmer{1},
		Data: CountData{2},
	}).Encode(ws[0])
	(&CountTuple{
		Kmer: Kmer{2},
		Data: CountData{1},
	}).Encode(ws[0])
	(&CountTuple{
		Kmer: Kmer{6},
		Data: CountData{5},
	}).Encode(ws[0])

	(&CountTuple{
		Kmer: Kmer{0},
		Data: CountData{4},
	}).Encode(ws[1])
	(&CountTuple{
		Kmer: Kmer{1},
		Data: CountData{10},
	}).Encode(ws[1])
	(&CountTuple{
		Kmer: Kmer{2},
		Data: CountData{3},
	}).Encode(ws[1])
	(&CountTuple{
		Kmer: Kmer{4},
		Data: CountData{2},
	}).Encode(ws[1])

	(&CountTuple{
		Kmer: Kmer{2},
		Data: CountData{1},
	}).Encode(ws[2])
	(&CountTuple{
		Kmer: Kmer{4},
		Data: CountData{5},
	}).Encode(ws[2])

	want := []*CountTuple{
		{Kmer: Kmer{0}, Data: CountData{4}},
		{Kmer: Kmer{1}, Data: CountData{12}},
		{Kmer: Kmer{2}, Data: CountData{5}},
		{Kmer: Kmer{4}, Data: CountData{7}},
		{Kmer: Kmer{6}, Data: CountData{5}},
	}

	m := NewMerger[CountHandler]()
	for i := range bufs {
		if err := m.Add(IterTuplesReader[CountHandler](bufs[i])); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &bytes.Buffer{}
	if err := m.Dump(out); err != nil {
		t.Fatalf("dump(...) failed: %v", err)
	}

	ct := &CountTuple{}
	for i := range want {
		if err := ct.Decode(out); err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		if !countTuplesEqual(ct, want[i]) {
			t.Fatalf("next()=%v, want %v", ct, want[i])
		}
	}
	if err := ct.Decode(out); err != io.EOF {
		t.Fatalf("next()=(%v, %v), want EOF", ct, err)
	}
}
