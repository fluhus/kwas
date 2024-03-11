package kmr

import (
	"fmt"
	"testing"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/snm"
	"golang.org/x/exp/slices"
)

func TestCheckpoints(t *testing.T) {
	want := []byte("AAAACCCGGT")
	n := 100
	failed := 0
	for i := 0; i < n; i++ {
		cp := Checkpoints(len(want))
		got := snm.Slice(len(cp), func(i int) byte {
			return sequtil.DNAFrom2Bit(nil, cp[i][:])[0]
		})
		if !slices.Equal(got, want) {
			failed++
		}
	}
	if failed > n/50 {
		t.Errorf("checkpoints(%d) prefixes!=%s %d times, want %d",
			len(want), want, failed, n/50)
	}
}

func BenchmarkCheckpoints(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Checkpoints(n)
			}
		})
	}
}
