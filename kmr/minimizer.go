package kmr

import "github.com/fluhus/kwas/util"

// Minimizer returns the minimal hash of an m-long canonical subsequence of
// kmer.
func Minimizer(kmer []byte, m int) uint64 {
	result := ^uint64(0)
	util.CanonicalKmers(kmer, m, func(b []byte) {
		h := util.Hash64(b)
		if h < result {
			result = h
		}
	})
	return result
}
