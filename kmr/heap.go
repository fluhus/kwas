// Functionality for merging sorted streams of kmers.

package kmr

import (
	"container/heap"
	"fmt"
	"io"

	"github.com/fluhus/kwas/progress"
)

// Tuple is a type that has a kmer and some data related to that kmer.
type Tuple interface {
	GetKmer() Kmer              // Returns the tuple's kmer
	Encode(io.Writer) error     // Writes the tuple
	Decode(io.ByteReader) error // Loads a tuple into this object
	Add(Tuple)                  // Adds another tuple of the same kmer to this
	Copy() Tuple                // Deep-copies the tuple
}

// TupleFromString returns a Tuple that matches the given string.
func TupleFromString(s string) Tuple {
	switch s {
	case "has":
		return &HasTuple{Sort: true}
	case "cnt":
		return &CountTuple{}
	case "prf":
		return &ProfileTuple{}
	case "genes":
		return &GeneSetTuple{}
	default:
		return nil
	}
}

// An iterator over kmer tuples from an input stream.
type kmerIter struct {
	r   io.ByteReader
	cur Tuple
}

// Creates a new iterator over the given input.
func newKmerIter(r io.ByteReader, t Tuple) (*kmerIter, error) {
	it := &kmerIter{r: r, cur: t}
	err := it.next()
	if err != nil {
		return nil, err
	}
	return it, nil
}

// Advances the iterator and sets cur to the next kmer tuple.
func (it *kmerIter) next() error {
	err := it.cur.Decode(it.r)
	if err != nil {
		it.cur = nil
	}
	return err
}

func (it *kmerIter) String() string {
	return fmt.Sprintf("{cur:%v}", it.cur.GetKmer())
}

// Merger merges sorted streams of kmer tuples.
type Merger []*kmerIter

// Add adds an input stream to be merged by this merger.
func (m *Merger) Add(r io.ByteReader, t Tuple) error {
	it, err := newKmerIter(r, t)
	if err != nil {
		return err
	}
	heap.Push(m, it)
	return nil
}

// Next returns the next kmer tuple, possible merged from a several streams.
// Returned kmers are sorted.
func (m *Merger) Next() (Tuple, error) {
	if m.Len() == 0 {
		panic("called next() on an empty heap")
	}

	result := (*m)[0].cur.Copy()
	for {
		if err := m.nextMin(); err != nil {
			return nil, err
		}
		if m.Len() == 0 || (*m)[0].cur.GetKmer() != result.GetKmer() {
			break
		}
		result.Add((*m)[0].cur)
	}

	return result, nil
}

// Advances the minimal iterator and fixes the heap.
func (m *Merger) nextMin() error {
	err := (*m)[0].next()
	if err == io.EOF {
		heap.Pop(m)
		return nil
	}
	if err != nil {
		return err
	}
	heap.Fix(m, 0)
	return nil
}

// Dump merges all the remaining kmer tuples and writes them to the given writer.
func (m *Merger) Dump(w io.Writer) error {
	n := 0
	pt := progress.NewTimerFunc(func(i int) string {
		return fmt.Sprintf("%d kmers dumped", i)
	})

	for m.Len() > 0 {
		tup, err := m.Next()
		if err != nil {
			return err
		}
		err = tup.Encode(w)
		if err != nil {
			return err
		}
		n++
		pt.Inc()
	}
	pt.Done()
	return nil
}

// Heap interface

func (m *Merger) Len() int {
	return len(*m)
}
func (m *Merger) Less(i, j int) bool {
	icur := (*m)[i].cur
	jcur := (*m)[j].cur
	return icur.GetKmer().Less(jcur.GetKmer())
}
func (m *Merger) Swap(i, j int) {
	(*m)[i], (*m)[j] = (*m)[j], (*m)[i]
}
func (m *Merger) Push(a interface{}) {
	*m = append(*m, a.(*kmerIter))
}
func (m *Merger) Pop() interface{} {
	result := (*m)[len(*m)-1]
	*m = (*m)[:len(*m)-1]
	return result
}
