// Package kmr provides common utilities for kmer handling.
package kmr

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/maps"
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

// CountTuple is a kmer with its allele counts.
type CountTuple struct {
	Kmer  Kmer
	Count Counts
}

func (t *CountTuple) Copy() Tuple {
	tup := *t
	return &tup
}

func (t *CountTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *CountTuple) Add(other Tuple) {
	otherc := other.(*CountTuple)
	for i, cnt := range otherc.Count {
		t.Count[i] += cnt
	}
}

func (t *CountTuple) Encode(w aio.Writer) error {
	if err := bnry.Write(w, t.Kmer[:],
		t.Count[0], t.Count[1], t.Count[2], t.Count[3]); err != nil {
		return err
	}
	return nil
}

func (t *CountTuple) Decode(r aio.Reader) error {
	var b []byte
	var c Counts
	if err := bnry.Read(r, &b, &c[0], &c[1], &c[2], &c[3]); err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("mismatching length: %v, want %v",
			len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	t.Count = c
	return nil
}

// CountMap maps a kmer to its count.
type CountMap map[Kmer]Counts

// Encode writes this map to the given writer.
func (c CountMap) Encode(w aio.Writer) error {
	kmers := maps.Keys(c)
	slices.SortFunc(kmers, func(a, b Kmer) bool {
		return bytes.Compare(a[:], b[:]) == -1
	})
	for _, kmer := range kmers {
		ct := CountTuple{Kmer: kmer, Count: c[kmer]}
		if err := ct.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

// Decode adds the counts of the encoded map to this one.
func (c CountMap) Decode(r aio.Reader) error {
	tup := &CountTuple{}
	var err error
	for err = tup.Decode(r); err == nil; err = tup.Decode(r) {
		cnt, ok := c[tup.Kmer]
		if !ok {
			c[tup.Kmer] = tup.Count
			continue
		}
		for i := range cnt {
			cnt[i] += tup.Count[i]
		}
		c[tup.Kmer] = cnt
	}
	if err != io.EOF {
		return err
	}
	return nil
}

// MajAlTuple is a kmer along with its major allele.
type MajAlTuple struct {
	Kmer Kmer
	Maj  byte
}

func (t *MajAlTuple) Copy() Tuple {
	tup := *t
	return &tup
}

func (t *MajAlTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *MajAlTuple) Encode(w aio.Writer) error {
	if err := aio.WriteBytes(w, t.Kmer[:]); err != nil {
		return err
	}
	if err := aio.WriteUvarint(w, uint64(t.Maj)); err != nil {
		return err
	}
	return nil
}

func (t *MajAlTuple) Decode(r aio.Reader) error {
	b, err := aio.ReadBytes(r)
	if err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("mismatching length: %v, want %v",
			len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	u, err := aio.ReadUvarint(r)
	if err != nil {
		return aio.NotExpectingEOF(err)
	}
	t.Maj = byte(u)
	return nil
}

func (t *MajAlTuple) Add(other Tuple) {
	panic("add is unsupported for MajAlTuple")
}

// MajAlMap maps a kmer to its major allele.
type MajAlMap map[Kmer]byte

// Encode writes this map to the given writer.
func (m MajAlMap) Encode(w aio.Writer) error {
	kmers := make([]Kmer, 0, len(m))
	for kmer := range m {
		kmers = append(kmers, kmer)
	}
	sort.Slice(kmers, func(i, j int) bool {
		return bytes.Compare(kmers[i][:], kmers[j][:]) == -1
	})
	for _, kmer := range kmers {
		tup := MajAlTuple{Kmer: kmer, Maj: m[kmer]}
		if err := tup.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

// Decode adds the counts of the encoded map to this one.
func (m MajAlMap) Decode(r aio.Reader) error {
	tup := &MajAlTuple{}
	var err error
	for err = tup.Decode(r); err == nil; err = tup.Decode(r) {
		_, ok := m[tup.Kmer]
		if ok {
			return fmt.Errorf("recurring kmer in majal map: %v", tup.Kmer)
		}
		m[tup.Kmer] = tup.Maj
	}
	if err != io.EOF {
		return err
	}
	return nil
}

type MAFValue byte

const (
	MAFMajor MAFValue = 1
	MAFMinor MAFValue = 2
	MAFBoth  MAFValue = 3
)

// SampleMAFTuple is a sample along with its MAF for a specific kmer.
type SampleMAFTuple struct {
	SampleID int // Index in the samples list
	MAF      MAFValue
}

// MAFTuple is a kmer with the MAFs of the samples that have it.
type MAFTuple struct {
	Kmer Kmer
	MAF  []SampleMAFTuple
}

func (t *MAFTuple) Copy() Tuple {
	tup := MAFTuple{Kmer: t.Kmer, MAF: make([]SampleMAFTuple, len(t.MAF))}
	copy(tup.MAF, t.MAF)
	return &tup
}

func (t *MAFTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *MAFTuple) Add(other Tuple) {
	othert := other.(*MAFTuple)
	t.MAF = append(t.MAF, othert.MAF...)
}

func (t *MAFTuple) Encode(w aio.Writer) error {
	sort.Slice(t.MAF, func(i, j int) bool {
		return t.MAF[i].SampleID < t.MAF[j].SampleID
	})
	if err := aio.WriteBytes(w, t.Kmer[:]); err != nil {
		return err
	}
	if err := aio.WriteUvarint(w, uint64(len(t.MAF))); err != nil {
		return err
	}
	last := -1 // Indicator that no value was written yet.
	for _, smaf := range t.MAF {
		if smaf.SampleID <= last {
			panic(fmt.Sprintf("sample ID (%v) is unexpectedly not greater than "+
				"last (%v)", smaf.SampleID, last))
		}
		diff := smaf.SampleID - last
		if last == -1 {
			diff--
		}
		if err := aio.WriteUvarint(w, uint64(diff)); err != nil {
			return err
		}
		if err := w.WriteByte(byte(smaf.MAF)); err != nil {
			return err
		}
		last = smaf.SampleID
	}
	return nil
}

func (t *MAFTuple) Decode(r aio.Reader) error {
	b, err := aio.ReadBytes(r)
	if err != nil {
		return err
	}
	if len(b) != len(t.Kmer) {
		return fmt.Errorf("mismatching length: %v, want %v",
			len(b), len(t.Kmer))
	}
	copy(t.Kmer[:], b)
	n, err := aio.ReadUvarint(r)
	if err != nil {
		return aio.NotExpectingEOF(err)
	}
	t.MAF = make([]SampleMAFTuple, n)
	last := 0
	for i := range t.MAF {
		u, err := aio.ReadUvarint(r)
		if err != nil {
			return aio.NotExpectingEOF(err)
		}
		b, err := r.ReadByte()
		if err != nil {
			return aio.NotExpectingEOF(err)
		}
		t.MAF[i] = SampleMAFTuple{SampleID: last + int(u), MAF: MAFValue(b)}
		last += int(u)
	}
	return nil
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

func (t *HasTuple) Encode(w aio.Writer) error {
	if !sort.IntsAreSorted(t.Samples) {
		sort.Ints(t.Samples)
	}
	if err := bnry.Write(w, t.Kmer[:], toDiffs(t.Samples)); err != nil {
		return nil
	}
	return nil
}

func (t *HasTuple) Decode(r aio.Reader) error {
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
