package kmr

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
)

type Writer struct {
	w   *bnry.Writer
	cur Kmer
}

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

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: bnry.NewWriter(w)}
}

type Reader struct {
	r   io.ByteReader
	cur Kmer
}

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

func NewReader(r io.ByteReader) *Reader {
	return &Reader{r: r}
}
