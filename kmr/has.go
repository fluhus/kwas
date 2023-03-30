package kmr

import (
	"fmt"
	"io"

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
	Sort    bool
}

func (t *HasTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *HasTuple) Copy() Tuple {
	cp := &HasTuple{
		Kmer:    t.Kmer,
		Samples: make([]int, len(t.Samples)),
		Sort:    t.Sort,
	}
	copy(cp.Samples, t.Samples)
	return cp
}

func (t *HasTuple) Add(other Tuple) {
	othert := other.(*HasTuple)
	t.Samples = append(t.Samples, othert.Samples...)
}

func (t *HasTuple) Encode(w io.Writer) error {
	if t.Sort {
		slices.Sort(t.Samples)
	}
	toDiffs(t.Samples)
	err := bnry.Write(w, t.Kmer[:], t.Samples)
	fromDiffs(t.Samples)
	if err != nil {
		return nil
	}
	return nil
}

func (t *HasTuple) Decode(r io.ByteReader) error {
	var b []byte
	if err := bnry.Read(r, &b, &t.Samples); err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %v, want %v", len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	fromDiffs(t.Samples)
	return nil
}

func toDiffs(a []int) {
	if len(a) == 0 {
		return
	}
	last := a[0]
	for i := range a[1:] {
		lastt := a[i+1]
		a[i+1] = a[i+1] - last
		last = lastt
	}
}

func fromDiffs(a []int) {
	if len(a) == 0 {
		return
	}
	for i := range a[1:] {
		a[i+1] = a[i] + a[i+1]
	}
}
