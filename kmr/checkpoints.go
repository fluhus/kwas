package kmr

import (
	"bytes"
	"math/rand"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/slices"
)

// Checkpoints returns n canonical kmers that divide the space into approximately
// equal buckets. The last checkpoint is all T's.
func Checkpoints(n int) []Kmer {
	kmers := make([]Kmer, n*1000)
	buf := make([]byte, K)
	rc := make([]byte, K)
	for i := range kmers {
		for j := range buf {
			buf[j] = sequtil.Iton(rand.Intn(4))
		}
		rc = sequtil.ReverseComplement(rc[:0], buf)
		if bytes.Compare(buf, rc) == 1 {
			buf, rc = rc, buf
		}
		sequtil.DNATo2Bit(kmers[i][:0], buf)
	}

	// fmt.Println("Sorting")
	slices.SortFunc(kmers, func(a, b Kmer) bool { return a.Less(b) })
	return snm.Slice(n, func(i int) Kmer {
		if i == n-1 { // Last checkpoint is the maximal kmer.
			return Kmer(snm.Slice(K, func(i int) byte { return 255 }))
		}
		return kmers[(i+1)*1000]
	})
}
