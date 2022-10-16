package kmr

import (
	"fmt"
	"io"
	"sort"

	"github.com/fluhus/gostuff/bnry"
	"golang.org/x/exp/slices"
)

// HasCount counts how many samples have a kmer.
type HasCount struct {
	Kmer  FullKmer
	Count uint64
}

func (p *HasCount) GetKmer() []byte {
	return p.Kmer[:]
}

func (p *HasCount) Add(other Tuple) {
	o := other.(*HasCount)
	if p.Kmer != o.Kmer {
		panic(fmt.Sprintf("mismatching kmers: %q %q", p.Kmer, o.Kmer))
	}
	p.Count += o.Count
}

func (p *HasCount) Copy() Tuple {
	return &HasCount{p.Kmer, p.Count}
}

func (p *HasCount) Encode(w io.Writer) error {
	return bnry.Write(w, p.Kmer[:], p.Count)
}

func (p *HasCount) Decode(r io.ByteReader) error {
	var b []byte
	if err := bnry.Read(r, &b, &p.Count); err != nil {
		return err
	}
	if len(b) != len(p.Kmer) {
		return fmt.Errorf("unexpected length: %d, want %d",
			len(b), len(p.Kmer))
	}
	copy(p.Kmer[:], b)
	return nil
}

// HasTuple represents a kmer with the samples that have it.
type HasTuple struct {
	Kmer    FullKmer
	Samples []int // Indexes of samples that have this kmer.
}

func (t *HasTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *HasTuple) Copy() Tuple {
	cp := &HasTuple{
		Kmer:    t.Kmer,
		Samples: make([]int, len(t.Samples)),
	}
	copy(cp.Samples, t.Samples)
	return cp
}

func (t *HasTuple) Add(other Tuple) {
	othert := other.(*HasTuple)
	t.Samples = append(t.Samples, othert.Samples...)
}

func (t *HasTuple) Encode(w io.Writer) error {
	if !sort.IntsAreSorted(t.Samples) {
		sort.Ints(t.Samples)
	}
	if err := bnry.Write(w, t.Kmer[:], toDiffs(t.Samples)); err != nil {
		return nil
	}
	return nil
}

func (t *HasTuple) Decode(r io.ByteReader) error {
	var b []byte
	var diffs []uint64
	if err := bnry.Read(r, &b, &diffs); err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %v, want %v", len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	t.Samples = fromDiffs(diffs)
	if !slices.IsSorted(t.Samples) {
		return fmt.Errorf("samples are not sorted")
	}
	return nil
}

func toDiffs(a []int) []uint64 {
	if len(a) == 0 {
		return nil
	}
	diffs := make([]uint64, len(a))
	diffs[0] = uint64(a[0])
	for i := range a[1:] {
		diffs[i+1] = uint64(a[i+1] - a[i])
	}
	return diffs
}

func fromDiffs(diffs []uint64) []int {
	if len(diffs) == 0 {
		return nil
	}
	a := make([]int, len(diffs))
	a[0] = int(diffs[0])
	for i := range diffs[1:] {
		a[i+1] = a[i] + int(diffs[i+1])
	}
	return a
}
