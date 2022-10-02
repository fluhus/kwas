package kmr

import (
	"fmt"

	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/aio"
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

func (t *GeneSetTuple) Encode(w aio.Writer) error {
	if err := aio.WriteBytes(w, t.Kmer[:]); err != nil {
		return err
	}
	if err := aio.WriteUvarint(w, uint64(len(t.Genes))); err != nil {
		return err
	}
	for g := range t.Genes {
		if err := aio.WriteString(w, g); err != nil {
			return err
		}
	}
	return nil
}

func (t *GeneSetTuple) Decode(r aio.Reader) error {
	b, err := aio.ReadBytes(r)
	if err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %d, want %d",
			len(b), len(t.Kmer))
	}
	t.Kmer = *(*FullKmer)(b)
	n, err := aio.ReadUvarint(r)
	if err != nil {
		return err
	}
	t.Genes = make(sets.Set[string], n)
	for i := uint64(0); i < n; i++ {
		g, err := aio.ReadString(r)
		if err != nil {
			return err
		}
		t.Genes.Add(g)
	}
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
