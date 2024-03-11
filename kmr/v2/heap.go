// Logic for merging sorted streams of kmers.

package kmr

import (
	"fmt"
	"io"
	"iter"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/heaps"
	"github.com/fluhus/gostuff/ptimer"
)

// An iterator over kmer tuples from an input stream.
type kmerStream[H KmerDataHandler[T], T any] struct {
	next func() (*Tuple[H, T], error, bool)
	stop func()
	cur  *Tuple[H, T]
}

// Creates a new iterator over the given input.
func newKmerStream[H KmerDataHandler[T], T any](
	seq iter.Seq2[*Tuple[H, T], error]) (*kmerStream[H, T], error) {
	read, stop := iter.Pull2(seq)
	it := &kmerStream[H, T]{next: read, stop: stop, cur: &Tuple[H, T]{}}
	err, ok := it.advance()
	if err != nil {
		return nil, err
	}
	if !ok { // TODO(amit): Reconsider this.
		return nil, io.ErrUnexpectedEOF
	}
	return it, nil
}

// Advances the iterator and sets cur to the next kmer tuple.
func (it *kmerStream[H, T]) advance() (error, bool) {
	t, err, ok := it.next()
	if err != nil || !ok {
		it.cur = nil
		return err, ok
	} else {
		it.cur = t
	}
	return err, ok
}

// Merger merges sorted streams of kmer tuples.
type Merger[H KmerDataHandler[T], T any] struct {
	h *heaps.Heap[*kmerStream[H, T]]
}

// NewMerger returns a new merger.
func NewMerger[H KmerDataHandler[T], T any]() *Merger[H, T] {
	return &Merger[H, T]{
		heaps.New(func(ki1, ki2 *kmerStream[H, T]) bool {
			return ki1.cur.Kmer.Less(ki2.cur.Kmer)
		})}
}

// Add adds an input stream to be merged by this merger.
func (m *Merger[H, T]) Add(seq iter.Seq2[*Tuple[H, T], error]) error {
	it, err := newKmerStream(seq)
	if err != nil {
		return err
	}
	m.h.Push(it)
	return nil
}

// Next returns the next kmer tuple, possible merged from a several streams.
// Returned kmers are sorted.
func (m *Merger[H, T]) Next() (*Tuple[H, T], error) {
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
func (m *Merger[H, T]) nextMin() error {
	err, ok := m.h.Head().advance()
	if !ok {
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
func (m *Merger[H, T]) Dump(w io.Writer) error {
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
	}
	pt.Done()
	return nil
}
