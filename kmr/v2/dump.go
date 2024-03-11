// Dump format logic.

package kmr

import (
	"encoding/binary"
	"fmt"
	"io"
	"iter"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
)

// A Writer writes kmers in a condensed format.
// Sorted kmers make a smaller output.
type Writer struct {
	w   *bnry.Writer
	cur Kmer
}

// Write writes the given kmer to the underlying writer.
func (w *Writer) Write(kmer Kmer) error {
	i := 0
	for i = range kmer {
		if kmer[i] != w.cur[i] {
			break
		}
	}
	if err := w.w.Write(kmer[i:]); err != nil {
		return err
	}
	w.cur = kmer
	return nil
}

// NewWriter returns a new kmer writer that writes to the given reader.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: bnry.NewWriter(w)}
}

// A Reader reads kmers from a stream.
type Reader struct {
	r   io.ByteReader
	cur Kmer
}

// Read reads the next kmer.
// It is generally better to use the iterator functions.
func (r *Reader) Read() (Kmer, error) {
	n, err := binary.ReadUvarint(r.r)
	if err != nil {
		return Kmer{}, err
	}
	if n > uint64(len(r.cur)) {
		return Kmer{}, fmt.Errorf("bad kmer piece length: %d", n)
	}
	for i := range r.cur[:n] {
		r.cur[len(r.cur)-int(n)+i], err = r.r.ReadByte()
		if err != nil {
			return Kmer{}, err
		}
	}
	return r.cur, nil
}

// NewReader returns a new reader from the given stream.
func NewReader(r io.ByteReader) *Reader {
	return &Reader{r: r}
}

// IterKmersFile iterates over kmers in a dump file.
func IterKmersFile(file string) iter.Seq2[Kmer, error] {
	return func(yield func(Kmer, error) bool) {
		f, err := aio.Open(file)
		if err != nil {
			yield(Kmer{}, err)
			return
		}
		defer f.Close()
		r := NewReader(f)
		for {
			kmer, err := r.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(Kmer{}, err)
				return
			}
			if !yield(kmer, nil) {
				return
			}
		}
	}
}

// IterKmersReader iterates over kmers in a reader.
func IterKmersReader(r io.ByteReader) iter.Seq2[Kmer, error] {
	return func(yield func(Kmer, error) bool) {
		rr := NewReader(r)
		for {
			kmer, err := rr.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(Kmer{}, err)
				return
			}
			if !yield(kmer, nil) {
				return
			}
		}
	}
}
