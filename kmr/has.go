package kmr

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
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
