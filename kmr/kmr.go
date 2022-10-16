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
	K       = 33
	K2BFull = (K + 3) / 4
)

// FullKmer is a 2-bit kmer including its SNP.
type FullKmer [K2BFull]byte

// FullKmerSet is a set of unique full kmers.
type FullKmerSet = sets.Set[FullKmer]

// ReadFullKmersLines reads a set of kmers from a file.
func ReadFullKmersLines(file string) (FullKmerSet, error) {
	kmers, err := util.ReadLines(aio.Open(file))
	if err != nil {
		return nil, err
	}

	m := make(FullKmerSet, len(kmers))
	var buf, buf2 []byte
	for _, kmer := range kmers {
		buf2 = append(buf2[:0], kmer...) // Efficiently convert string to bytes.
		buf = sequtil.DNATo2Bit(buf[:0], buf2)
		m.Add(*(*FullKmer)(buf))
	}
	if len(m) != len(kmers) {
		return nil, fmt.Errorf("bad map length: %v, want %v",
			len(m), len(kmers))
	}
	return m, nil
}
