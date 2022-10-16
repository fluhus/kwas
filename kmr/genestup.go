package kmr

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/sets"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type GeneSetTuple struct {
	Kmer  FullKmer
	Genes sets.Set[string]
}

func (t *GeneSetTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *GeneSetTuple) Encode(w io.Writer) error {
	return bnry.Write(w, t.Kmer[:], t.Genes)
}

func (t *GeneSetTuple) Decode(r io.ByteReader) error {
	var b []byte
	var s []string
	if err := bnry.Read(r, &b, &s); err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %d, want %d",
			len(b), len(t.Kmer))
	}
	t.Kmer = *(*FullKmer)(b)
	t.Genes = sets.Set[string]{}.Add(s...)
	return nil
}

func (t *GeneSetTuple) Add(other Tuple) {
	o := other.(*GeneSetTuple)
	if !slices.Equal(t.Kmer[:], o.Kmer[:]) {
		panic(fmt.Sprintf("mismatching kmers: %q %q", t.Kmer, o.Kmer))
	}
	t.Genes.AddSet(o.Genes)
}

func (t *GeneSetTuple) Copy() Tuple {
	return &GeneSetTuple{Kmer: t.Kmer, Genes: maps.Clone(t.Genes)}
}
