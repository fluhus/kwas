// Package kmr provides common utilities for kmer handling.
package kmr

import (
	"fmt"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/util"
)

const (
	K   = 20          // Kmer length.
	K2B = (K + 3) / 4 // 2-bit kmer length.
)

// Kmer is a 2-bit kmer.
type Kmer [K2B]byte

// KmerSet is a set of unique full kmers.
type KmerSet = sets.Set[Kmer]

// ReadKmersLines reads a set of kmers from a file.
func ReadKmersLines(file string) (KmerSet, error) {
	kmers, err := util.ReadLines(aio.Open(file))
	if err != nil {
		return nil, err
	}

	m := make(KmerSet, len(kmers))
	var buf, buf2 []byte
	for _, kmer := range kmers {
		buf2 = append(buf2[:0], kmer...) // Efficiently convert string to bytes.
		buf = sequtil.DNATo2Bit(buf[:0], buf2)
		m.Add(*(*Kmer)(buf))
	}
	if len(m) != len(kmers) {
		return nil, fmt.Errorf("bad map length: %v, want %v",
			len(m), len(kmers))
	}
	return m, nil
}

// Less compares the receiver to the argument lexicographically.
func (a Kmer) Less(b Kmer) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// Compare returns -1 if a is lexicographically less than b, 1 if b is less than
// a, or 0 if they are equal.
func (a Kmer) Compare(b Kmer) int {
	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
