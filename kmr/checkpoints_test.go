package kmr

import "testing"

func Benchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Checkpoints(1)
	}
}
