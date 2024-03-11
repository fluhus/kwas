// CountTuple logic.

package kmr

import (
	"io"

	"github.com/fluhus/gostuff/bnry"
)

// CountTuple holds a kmer and its appearance count.
type CountTuple = Tuple[CountHandler, CountData]

type CountData struct {
	Count int
}

type CountHandler struct{}

func (h CountHandler) encode(c CountData, w *bnry.Writer) error {
	return w.Write(c.Count)
}

func (h CountHandler) decode(c *CountData, r io.ByteReader) error {
	return bnry.Read(r, &c.Count)
}

func (h CountHandler) merge(a, b CountData) CountData {
	return CountData{a.Count + b.Count}
}

func (h CountHandler) clone(c CountData) CountData {
	return c
}

func (h CountHandler) new() CountData {
	return CountData{}
}
