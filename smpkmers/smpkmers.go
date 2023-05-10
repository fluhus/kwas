// Randomly samples kmers and sample IDs from HAS files.
package main

import (
	"flag"
	"fmt"
	"math/rand"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

var (
	fin  = flag.String("i", "", "Input HAS file glob template")
	fout = flag.String("o", "", "Output file")
	n    = flag.Int("n", 0, "Subsample exactly n kmers")
	r    = flag.Float64("r", 0, "Subsample kmers with probability 1/r")
	s    = flag.Uint("s", 0, "Subsample sample IDs with probability 1/s")
)

func main() {
	flag.Parse()
	if sumBools(*n != 0, *r != 0, *s != 0) != 1 {
		util.Die(fmt.Errorf("please use either -n, -r or -s"))
	}
	if *r != 0 {
		ratio := 1.0 / *r
		kept := 0
		out, err := aio.Create(*fout)
		util.Die(err)
		w := bnry.NewWriter(out)
		pt := progress.NewTimerFunc(func(i int) string {
			return fmt.Sprintf("%d (kept %d)", i, kept)
		})
		util.Die(
			kmr.IterTuplesFiles(*fin, &kmr.HasTuple{}, func(t *kmr.HasTuple) error {
				pt.Inc()
				if rand.Float64() < ratio {
					kept++
					return t.Encode(w)
				}
				return nil
			}))
		util.Die(out.Close())
		pt.Done()
	} else if *n != 0 {
		r := util.NewReservoir[*kmr.HasTuple](*n)
		pt := progress.NewTimer()
		util.Die(
			kmr.IterTuplesFiles(*fin, &kmr.HasTuple{}, func(t *kmr.HasTuple) error {
				r.Add(t.Copy())
				pt.Inc()
				return nil
			}))
		pt.Done()
		out, err := aio.Create(*fout)
		util.Die(err)
		w := bnry.NewWriter(out)
		for _, t := range r.Sample {
			util.Die(t.Encode(w))
		}
		util.Die(out.Close())
	} else if *s != 0 {
		s := uint64(*s)
		out, err := aio.Create(*fout)
		util.Die(err)
		w := bnry.NewWriter(out)
		pt := progress.NewTimer()
		util.Die(
			kmr.IterTuplesFiles(*fin, &kmr.HasTuple{}, func(t *kmr.HasTuple) error {
				var samples []int
				for _, i := range t.Data.Samples {
					if hashInt(i)%s == 0 {
						samples = append(samples, i)
					}
				}
				t.Data.Samples = samples
				pt.Inc()
				return t.Encode(w)
			}))
		util.Die(out.Close())
		pt.Done()
	}
	fmt.Println("Done")
}

// Buffer for int hashing.
var hashBuf = make([]byte, 8)

// Returns the hash value of the given int.
func hashInt(n int) uint64 {
	if n < 0 {
		panic(fmt.Sprintf("invalid i: %d", n))
	}
	for i := range hashBuf {
		hashBuf[i] = byte((n >> (i * 8)) & 255)
	}
	return util.Hash64(hashBuf)
}

// Returns the number of true's.
func sumBools(b ...bool) int {
	sum := 0
	for _, bb := range b {
		if bb {
			sum++
		}
	}
	return sum
}
