package util

type NTracker struct {
	i    int
	k    int
	last int
	seq  []byte
}

func NewNTracker(seq []byte, k int) *NTracker {
	last := -1
	for i := range seq[:k] {
		if seq[i] == 'n' || seq[i] == 'N' {
			last = i
		}
	}
	return &NTracker{0, k, last, seq}
}

func (t *NTracker) NextN() bool {
	lasti := t.i + t.k - 1
	if t.seq[lasti] == 'n' || t.seq[lasti] == 'N' {
		t.last = lasti
	}
	t.i++
	return t.last >= t.i-1
}
