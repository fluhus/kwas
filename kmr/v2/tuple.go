// KmerTuple type logic.

package kmr

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
)

// Tuple holds a kmer and some data attached to it.
type Tuple[H KmerDataHandler[T], T any] struct {
	h    H
	Kmer Kmer
	Data T
	buf  []byte
}

// KmerDataHandler implements functions for handling data in a kmer tuple.
type KmerDataHandler[T any] interface {
	encode(T, *bnry.Writer) error   // Writes the data.
	decode(*T, io.ByteReader) error // Loads data into this object.
	merge(T, T) T                   // Merges two pieces of data.
	clone(T) T                      // Deep-copies the data.
	new() T                         // Initializes an empty data.
}

// Encode writes this kmer and its data to the writer.
func (t *Tuple[H, T]) Encode(w *bnry.Writer) error {
	t.buf = t.Kmer[:]
	if err := w.Write(t.buf); err != nil {
		return err
	}
	if err := t.h.encode(t.Data, w); err != nil {
		return err
	}
	return nil
}

// Decode reads a kmer and its data and writes it to this instance.
func (t *Tuple[H, T]) Decode(r io.ByteReader) error {
	t.buf = t.Kmer[:0]
	if err := bnry.Read(r, &t.buf); err != nil {
		return err
	}
	if len(t.buf) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %v, want %v",
			len(t.buf), len(t.Kmer))
	}
	if err := t.h.decode(&t.Data, r); err != nil {
		return err
	}
	return nil
}

// Clone returns a deep copy of this instance.
func (t *Tuple[H, T]) Clone() *Tuple[H, T] {
	return &Tuple[H, T]{Kmer: t.Kmer, Data: t.h.clone(t.Data), buf: nil}
}

// Add adds the data of another kmer to this one.
func (t *Tuple[H, T]) Add(other *Tuple[H, T]) {
	if t.Kmer != other.Kmer {
		panic(fmt.Sprintf("mismatching kmers: %v %v", t.Kmer, other.Kmer))
	}
	t.Data = t.h.merge(t.Data, other.Data)
}

func NewTuple[H KmerDataHandler[T], T any]() *Tuple[H, T] {
	t := &Tuple[H, T]{}
	t.Data = t.h.new()
	return t
}
