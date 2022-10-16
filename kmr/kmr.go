// Package kmr provides common utilities for kmer handling.
package kmr

import (
	"fmt"
	"io"
	"sort"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/slices"
)

const (
	K       = 33
	K2BFull = (K + 3) / 4     // Kmer including SNP
	K2B     = (K - 1 + 3) / 4 // Neutralized kmer
	SNPPos  = K / 2

	Prefix = 2
)

// Kmer is a 2-bit kmer excluding its SNP.
type Kmer [K2B]byte

// FullKmer is a 2-bit kmer including its SNP.
type FullKmer [K2BFull]byte

// KmerPrefix is a kmer's prefix, for indexing.
type KmerPrefix [Prefix]byte

// KmerSuffix is a kmer's suffix, for indexing.
type KmerSuffix [K2BFull - Prefix]byte

// Counts is the alleles counts of a single kmer in its SNP position.
type Counts [4]uint16

// Sum returns the sum of counts.
func (c Counts) Sum() int {
	result := 0
	for _, a := range c {
		result += int(a)
	}
	return result
}

// HasTuple represents a kmer with the samples that have it.
type HasTuple struct {
	Kmer    FullKmer
	Samples []int // Indexes of samples that have this kmer.
}

func (t *HasTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *HasTuple) Copy() Tuple {
	cp := &HasTuple{
		Kmer:    t.Kmer,
		Samples: make([]int, len(t.Samples)),
	}
	copy(cp.Samples, t.Samples)
	return cp
}

func (t *HasTuple) Add(other Tuple) {
	othert := other.(*HasTuple)
	t.Samples = append(t.Samples, othert.Samples...)
}

func (t *HasTuple) Encode(w io.Writer) error {
	if !sort.IntsAreSorted(t.Samples) {
		sort.Ints(t.Samples)
	}
	if err := bnry.Write(w, t.Kmer[:], toDiffs(t.Samples)); err != nil {
		return nil
	}
	return nil
}

func (t *HasTuple) Decode(r io.ByteReader) error {
	var b []byte
	var diffs []uint64
	if err := bnry.Read(r, &b, &diffs); err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("bad kmer length: %v, want %v", len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	t.Samples = fromDiffs(diffs)
	if !slices.IsSorted(t.Samples) {
		return fmt.Errorf("samples are not sorted")
	}
	return nil
}

func toDiffs(a []int) []uint64 {
	if len(a) == 0 {
		return nil
	}
	diffs := make([]uint64, len(a))
	diffs[0] = uint64(a[0])
	for i := range a[1:] {
		diffs[i+1] = uint64(a[i+1] - a[i])
	}
	return diffs
}

func fromDiffs(diffs []uint64) []int {
	if len(diffs) == 0 {
		return nil
	}
	a := make([]int, len(diffs))
	a[0] = int(diffs[0])
	for i := range diffs[1:] {
		a[i+1] = a[i] + int(diffs[i+1])
	}
	return a
}

// KmerFromBytes creates a neutralized 2-bit kmer from a full kmer.
func KmerFromBytes(b []byte) Kmer {
	if len(b) != K {
		panic(fmt.Sprintf("bad length: %v, want %v %q",
			len(b), K, b))
	}
	cp := make([]byte, len(b)-1)
	copy(cp, b[:SNPPos])
	copy(cp[SNPPos:], b[SNPPos+1:])
	var tb Kmer
	sequtil.DNATo2Bit(tb[:0], cp)
	return tb
}

// KmerToBytes returns an expanded kmer from a neutralized 2-bit kmer. The SNP
// position has an 'A'.
func KmerToBytes(kmer Kmer) []byte {
	b := make([]byte, K)
	sequtil.DNAFrom2Bit(b[1:1], kmer[:])
	copy(b[:SNPPos], b[1:])
	b[SNPPos] = 'A'
	return b
}

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
