package util

import "iter"

// NonNSubseqs iterates over k-long subsequences of seq that do not contain
// an n/N character. Yields the index and the subsequence.
func NonNSubseqs(seq []byte, k int) iter.Seq2[int, []byte] {
	return func(yield func(int, []byte) bool) {
		lastn := -1
		for i, b := range seq[:k-1] {
			if b == 'n' || b == 'N' {
				lastn = i
			}
		}
		for i, b := range seq[k-1:] {
			if b == 'n' || b == 'N' {
				lastn = i + k - 1
			}
			if lastn >= i {
				continue
			}
			if !yield(i, seq[i:i+k]) {
				return
			}
		}
	}
}

// NonNSubseqsString iterates over k-long subsequences of seq that do not
// contain an n/N character. Yields the index and the subsequence.
func NonNSubseqsString(seq string, k int) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		lastn := -1
		for i, b := range seq[:k-1] {
			if b == 'n' || b == 'N' {
				lastn = i
			}
		}
		for i, b := range seq[k-1:] {
			if b == 'n' || b == 'N' {
				lastn = i + k - 1
			}
			if lastn >= i {
				continue
			}
			if !yield(i, seq[i:i+k]) {
				return
			}
		}
	}
}
