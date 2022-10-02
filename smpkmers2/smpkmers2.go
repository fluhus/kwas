// Randomly samples kmers from HAS files.
package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

var (
	fin  = flag.String("i", "", "Input HAS file")
	fout = flag.String("o", "", "Output file")
	n    = flag.Int("n", 0, "Sample one out of n")
)

func main() {
	flag.Parse()

	fi, err := aio.Open(*fin)
	util.Die(err)
	defer fi.Close()

	fo, err := aio.Create(*fout)
	util.Die(err)
	defer fo.Close()

	tup := &kmr.HasTuple{}
	pt := progress.NewTimer()
	nout := 0
	for err = tup.Decode(fi); err == nil; err = tup.Decode(fi) {
		pt.Inc()
		if rand.Intn(*n) > 0 {
			continue
		}
		_, err := fmt.Fprintf(fo, "%s\n",
			sequtil.DNAFrom2Bit(nil, tup.Kmer[:])[:kmr.K])
		util.Die(err)
		nout++
	}
	pt.Done()
	if err != io.EOF {
		util.Die(err)
	}
	fmt.Printf("Printed %d/%d kmers\n", nout, pt.N)
	fmt.Println("Done")
}
