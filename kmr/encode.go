package kmr

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/fluhus/gostuff/bnry"
)

type Writer struct {
	w   *bnry.Writer
	cur FullKmer
}

func (w *Writer) Write(kmer FullKmer) error {
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
	cur FullKmer
}

func (r *Reader) Read() (FullKmer, error) {
	n, err := binary.ReadUvarint(r.r)
	if err != nil {
		return FullKmer{}, err
	}
	if n > uint64(len(r.cur)) {
		return FullKmer{}, fmt.Errorf("bad kmer piece length: %d", n)
	}
	for i := range r.cur[:n] {
		r.cur[len(r.cur)-int(n)+i], err = r.r.ReadByte()
		if err != nil {
			return FullKmer{}, err
		}
	}
	return r.cur, nil
}

func NewReader(r io.ByteReader) *Reader {
	return &Reader{r: r}
}
