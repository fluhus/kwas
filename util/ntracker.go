package util

import "iter"

func NonNSubseqs(seq []byte, k int) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
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
			if !yield(seq[i : i+k]) {
				return
			}
		}
	}
}

func NonNSubseqsString(seq string, k int) iter.Seq[string] {
	return func(yield func(string) bool) {
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
			if !yield(seq[i : i+k]) {
				return
			}
		}
	}
}
