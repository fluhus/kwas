// Functionality for merging sorted streams of kmers.

package kmr

import (
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/heaps"
	"github.com/fluhus/gostuff/ptimer"
)

// An iterator over kmer tuples from an input stream.
type kmerIter1[T any, H KmerDataHandler[T]] struct {
	r   io.ByteReader
	cur *Tuple[T, H]
}

// Creates a new iterator over the given input.
func newKmerIter1[T any, H KmerDataHandler[T]](r io.ByteReader,
	t *Tuple[T, H]) (*kmerIter1[T, H], error) {
	it := &kmerIter1[T, H]{r: r, cur: t}
	err := it.next()
	if err != nil {
		return nil, err
	}
	return it, nil
}

// Advances the iterator and sets cur to the next kmer tuple.
func (it *kmerIter1[T, H]) next() error {
	err := it.cur.Decode(it.r)
	if err != nil {
		it.cur = nil
	}
	return err
}

// Merger1 merges sorted streams of kmer tuples.
type Merger1[T any, H KmerDataHandler[T]] struct {
	h *heaps.Heap[*kmerIter1[T, H]]
	z *Tuple[T, H]
}

// NewMerger1 returns a new merger.
func NewMerger1[T any, H KmerDataHandler[T]](zero *Tuple[T, H]) *Merger1[T, H] {
	return &Merger1[T, H]{
		heaps.New(func(ki1, ki2 *kmerIter1[T, H]) bool {
			return ki1.cur.Kmer.Less(ki2.cur.Kmer)
		}), zero}
}

// Add adds an input stream to be merged by this merger.
func (m *Merger1[T, H]) Add(r io.ByteReader) error {
	it, err := newKmerIter1(r, m.z.Clone())
	if err != nil {
		return err
	}
	m.h.Push(it)
	return nil
}

// Next returns the next kmer tuple, possible merged from a several streams.
// Returned kmers are sorted.
func (m *Merger1[T, H]) Next() (*Tuple[T, H], error) {
	if m.h.Len() == 0 {
		panic("called next() on an empty heap")
	}

	result := m.h.Head().cur.Clone()
	for {
		if err := m.nextMin(); err != nil {
			return nil, err
		}
		if m.h.Len() == 0 || m.h.Head().cur.Kmer != result.Kmer {
			break
		}
		result.Add(m.h.Head().cur)
	}

	return result, nil
}

// Advances the minimal iterator and fixes the heap.
func (m *Merger1[T, H]) nextMin() error {
	err := m.h.Head().next()
	if err == io.EOF {
		m.h.Pop()
		return nil
	}
	if err != nil {
		return err
	}
	m.h.Fix(0)
	return nil
}

// Dump merges all the remaining kmer tuples and writes them to the given writer.
func (m *Merger1[T, H]) Dump(w io.Writer) error {
	bw := bnry.NewWriter(w)
	pt := ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d kmers dumped", i)
	})

	for m.h.Len() > 0 {
		tup, err := m.Next()
		if err != nil {
			return err
		}
		err = tup.Encode(bw)
		if err != nil {
			return err
		}
		pt.Inc()
		// if pt.N >= 100000000 {
		// 	break
		// } // TODO(amit): REMOVE
	}
	pt.Done()
	return nil
}
