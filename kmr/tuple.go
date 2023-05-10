// KmerTuple type logic.

package kmr

import (
	"fmt"
	"io"
	"sort"

	"github.com/fluhus/gostuff/bnry"
	"golang.org/x/exp/slices"
)

type KmerTuple[T any, H KmerDataHandler[T]] struct {
	Kmer    Kmer
	Data    T
	Handler H
	pkmer   []byte
}

type KmerDataHandler[T any] interface {
	encode(T, *bnry.Writer) error   // Writes the data
	decode(*T, io.ByteReader) error // Loads data into this object
	merge(T, T) T                   // Merges two pieces of data
	clone(T) T                      // Deep-copies the data
}

type CountTuple = KmerTuple[KmerCount, KmerCountHandler]
type HasTuple = KmerTuple[KmerHas, KmerHasHandler]

func (t *KmerTuple[T, H]) Encode(w *bnry.Writer) error {
	t.pkmer = t.Kmer[:]
	if err := w.Write(t.pkmer); err != nil {
		return err
	}
	if err := t.Handler.encode(t.Data, w); err != nil {
		return err
	}
	return nil
}

func (t *KmerTuple[T, H]) Decode(r io.ByteReader) error {
	t.pkmer = t.Kmer[:0]
	if err := bnry.Read(r, &t.pkmer); err != nil {
		return err
	}
	if len(t.pkmer) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %v, want %v",
			len(t.pkmer), len(t.Kmer))
	}
	if err := t.Handler.decode(&t.Data, r); err != nil {
		return err
	}
	return nil
}

func (t *KmerTuple[T, H]) Copy() *KmerTuple[T, H] {
	return &KmerTuple[T, H]{t.Kmer, t.Handler.clone(t.Data), t.Handler, nil}
}

func (t *KmerTuple[T, H]) Add(other *KmerTuple[T, H]) {
	if t.Kmer != other.Kmer {
		panic(fmt.Sprintf("mismatching kmers: %v %v", t.Kmer, other.Kmer))
	}
	t.Data = t.Handler.merge(t.Data, other.Data)
}

type KmerCount struct {
	Count int
}

type KmerCountHandler struct{}

func (h KmerCountHandler) encode(c KmerCount, w *bnry.Writer) error {
	return w.Write(c.Count)
}

func (h KmerCountHandler) decode(c *KmerCount, r io.ByteReader) error {
	return bnry.Read(r, &c.Count)
}

func (h KmerCountHandler) merge(a, b KmerCount) KmerCount {
	return KmerCount{a.Count + b.Count}
}

func (h KmerCountHandler) clone(c KmerCount) KmerCount {
	return c
}

type KmerHas struct {
	Samples      []int
	SortOnEncode bool
}

type KmerHasHandler struct{}

func (h KmerHasHandler) encode(c KmerHas, w *bnry.Writer) error {
	if c.SortOnEncode {
		sort.Ints(c.Samples)
	}
	toDiffs(c.Samples)
	err := w.Write(c.Samples)
	fromDiffs(c.Samples)
	return err
}

func (h KmerHasHandler) decode(c *KmerHas, r io.ByteReader) error {
	s := c.Samples[:0]
	if err := bnry.Read(r, &s); err != nil {
		return err
	}
	fromDiffs(s)
	c.Samples = s
	return nil
}

func (h KmerHasHandler) merge(a, b KmerHas) KmerHas {
	if a.SortOnEncode != b.SortOnEncode {
		panic(fmt.Sprintf("inputs disagree on SortOnEncode: %v, %v",
			a.SortOnEncode, b.SortOnEncode))
	}
	return KmerHas{append(a.Samples, b.Samples...), a.SortOnEncode}
}

func (h KmerHasHandler) clone(c KmerHas) KmerHas {
	return KmerHas{slices.Clone(c.Samples), c.SortOnEncode}
}

func fromDiffs(a []int) {
	if len(a) == 0 {
		return
	}
	for i := range a[1:] {
		a[i+1] = a[i] + a[i+1]
	}
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
