package util

import "testing"

func TestReservior(t *testing.T) {
	const (
		k = 10000 // Number of tests
		m = 10    // Number of elements
		n = 3     // Size of sample
	)

	// Run sampling.
	counts := map[int]int{}
	for i := 0; i < k; i++ {
		r := NewReservoir[int](3)
		for i := 1; i <= m; i++ {
			r.Add(i)
		}
		for _, val := range r.Sample {
			counts[val]++
		}
	}

	// Check distribution.
	const exp = k * n / m
	for k, v := range counts {
		if k < 1 || k > m {
			t.Errorf("unexpected value: %d, want 1-%d", k, m)
			continue
		}
		// Check that count is within 10% around expected.
		if v < exp*9/10 || v > exp*11/10 {
			t.Errorf("unexpected count: %d, want ~%d", v, exp)
		}
	}
}
