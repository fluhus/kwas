package kmr

import (
	"io"
	"reflect"
	"testing"

	"github.com/fluhus/kwas/aio"
)

func TestKmerIter(t *testing.T) {
	buf := &aio.Buffer{}
	CountMap{
		{7}: {7, 7, 7, 7},
		{2}: {0, 0, 1, 0},
		{1}: {2, 1, 2, 1},
	}.Encode(buf)

	want := []*CountTuple{
		{Kmer: Kmer{1}, Count: Counts{2, 1, 2, 1}},
		{Kmer: Kmer{2}, Count: Counts{0, 0, 1, 0}},
		{Kmer: Kmer{7}, Count: Counts{7, 7, 7, 7}},
	}

	it, err := newKmerIter(buf, &CountTuple{})
	if err != nil {
		t.Fatalf("newKmerIter(...) failed: %v", err)
	}
	for i := range want {
		if err != nil {
			t.Fatalf("next() failed: %v", err)
		}
		if !reflect.DeepEqual(it.cur, want[i]) {
			t.Fatalf("cur=%v, want %v", it.cur, want[i])
		}
		err = it.next()
	}
	if err != io.EOF {
		t.Fatalf("next() failed: %v", err)
	}
}

func TestMergerCount(t *testing.T) {
	bufs := make([]*aio.Buffer, 4)
	for i := range bufs {
		bufs[i] = &aio.Buffer{}
	}
	CountMap{
		{7}: {7, 7, 7, 7},
		{2}: {0, 0, 1, 0},
		{1}: {2, 1, 2, 1},
	}.Encode(bufs[0])
	CountMap{
		{3}: {3, 3, 4, 4},
		{1}: {1, 0, 0, 1},
		{0}: {1, 2, 3, 4},
	}.Encode(bufs[1])
	CountMap{
		{2}: {0, 0, 1, 0},
		{5}: {1, 0, 0, 1},
	}.Encode(bufs[2])
	CountMap{
		{2}: {2, 4, 6, 8},
		{0}: {1, 2, 3, 4},
	}.Encode(bufs[3])

	want := []CountTuple{
		{Kmer: Kmer{0}, Count: Counts{2, 4, 6, 8}},
		{Kmer: Kmer{1}, Count: Counts{3, 1, 2, 2}},
		{Kmer: Kmer{2}, Count: Counts{2, 4, 8, 8}},
		{Kmer: Kmer{3}, Count: Counts{3, 3, 4, 4}},
		{Kmer: Kmer{5}, Count: Counts{1, 0, 0, 1}},
		{Kmer: Kmer{7}, Count: Counts{7, 7, 7, 7}},
	}

	m := &Merger{}
	for i := range bufs {
		if err := m.Add(bufs[i], &CountTuple{}); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &aio.Buffer{}
	if err := m.Dump(out); err != nil {
		t.Fatalf("dump(...) failed: %v", err)
	}

	var ct CountTuple
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

func TestMergerMajAl(t *testing.T) {
	bufs := make([]*aio.Buffer, 3)
	for i := range bufs {
		bufs[i] = &aio.Buffer{}
	}
	MajAlMap{
		{7}: 3,
		{2}: 1,
		{1}: 0,
	}.Encode(bufs[0])
	MajAlMap{
		{3}: 3,
		{4}: 0,
		{0}: 2,
	}.Encode(bufs[1])
	MajAlMap{
		{9}: 0,
		{5}: 1,
	}.Encode(bufs[2])

	want := []MajAlTuple{
		{Kmer: Kmer{0}, Maj: 2},
		{Kmer: Kmer{1}, Maj: 0},
		{Kmer: Kmer{2}, Maj: 1},
		{Kmer: Kmer{3}, Maj: 3},
		{Kmer: Kmer{4}, Maj: 0},
		{Kmer: Kmer{5}, Maj: 1},
		{Kmer: Kmer{7}, Maj: 3},
		{Kmer: Kmer{9}, Maj: 0},
	}

	m := &Merger{}
	for i := range bufs {
		if err := m.Add(bufs[i], &MajAlTuple{}); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &aio.Buffer{}
	if err := m.Dump(out); err != nil {
		t.Fatalf("dump(...) failed: %v", err)
	}

	var ct MajAlTuple
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

func TestMergerMAF(t *testing.T) {
	bufs := make([]*aio.Buffer, 3)
	for i := range bufs {
		bufs[i] = &aio.Buffer{}
	}

	(&MAFTuple{
		Kmer: Kmer{1},
		MAF:  []SampleMAFTuple{{1, 1}, {3, 2}},
	}).Encode(bufs[0])
	(&MAFTuple{
		Kmer: Kmer{2},
		MAF:  []SampleMAFTuple{{5, 3}, {8, 1}},
	}).Encode(bufs[0])
	(&MAFTuple{
		Kmer: Kmer{6},
		MAF:  []SampleMAFTuple{{10, 2}, {14, 3}},
	}).Encode(bufs[0])

	(&MAFTuple{
		Kmer: Kmer{0},
		MAF:  []SampleMAFTuple{{0, 1}, {4, 2}},
	}).Encode(bufs[1])
	(&MAFTuple{
		Kmer: Kmer{1},
		MAF:  []SampleMAFTuple{{2, 2}, {5, 1}},
	}).Encode(bufs[1])
	(&MAFTuple{
		Kmer: Kmer{2},
		MAF:  []SampleMAFTuple{{2, 3}, {9, 1}},
	}).Encode(bufs[1])
	(&MAFTuple{
		Kmer: Kmer{4},
		MAF:  []SampleMAFTuple{{0, 1}, {4, 3}},
	}).Encode(bufs[1])

	(&MAFTuple{
		Kmer: Kmer{2},
		MAF:  []SampleMAFTuple{{1, 2}, {4, 1}},
	}).Encode(bufs[2])
	(&MAFTuple{
		Kmer: Kmer{4},
		MAF:  []SampleMAFTuple{{1, 2}, {3, 2}},
	}).Encode(bufs[2])

	want := []MAFTuple{
		{Kmer: Kmer{0}, MAF: []SampleMAFTuple{{0, 1}, {4, 2}}},
		{Kmer: Kmer{1}, MAF: []SampleMAFTuple{{1, 1}, {2, 2}, {3, 2}, {5, 1}}},
		{Kmer: Kmer{2}, MAF: []SampleMAFTuple{
			{1, 2}, {2, 3}, {4, 1}, {5, 3}, {8, 1}, {9, 1}}},
		{Kmer: Kmer{4}, MAF: []SampleMAFTuple{{0, 1}, {1, 2}, {3, 2}, {4, 3}}},
		{Kmer: Kmer{6}, MAF: []SampleMAFTuple{{10, 2}, {14, 3}}},
	}

	m := &Merger{}
	for i := range bufs {
		if err := m.Add(bufs[i], &MAFTuple{}); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &aio.Buffer{}
	if err := m.Dump(out); err != nil {
		t.Fatalf("dump(...) failed: %v", err)
	}

	var ct MAFTuple
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

func TestMergerHas(t *testing.T) {
	bufs := make([]*aio.Buffer, 3)
	for i := range bufs {
		bufs[i] = &aio.Buffer{}
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
		if err := m.Add(bufs[i], &HasTuple{}); err != nil {
			t.Fatalf("Add(...) failed: %v", err)
		}
	}

	out := &aio.Buffer{}
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
