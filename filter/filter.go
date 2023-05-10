// Filters out low-count kmers from KMC extraction.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
)

var (
	inFile  = flag.String("i", "", "Path to input file")
	outFile = flag.String("o", "", "Path to output file")
	min     = flag.Int("n", 0, "Minimal count to leave a kmer in")
	del     = flag.Bool("d", false, "Delete input file")
)

func main() {
	fmt.Println("Opening files")
	flag.Parse()
	fin, err := aio.Open(*inFile)
	util.Die(err)
	fout, err := aio.Create(*outFile)
	util.Die(err)
	kw := kmr.NewWriter(fout)

	fmt.Println("Filtering")
	cnt := &kmr.CountTuple{}
	kept := 0
	var last kmr.Kmer
	pt := ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("read %d, wrote %d (%d%%)", i, kept, kept*100/i)
	})
	for err = cnt.Decode(fin); err == nil; err = cnt.Decode(fin) {
		pt.Inc()
		if cnt.Kmer.Less(last) {
			util.Die(fmt.Errorf("kmers not in order: %v %v", last, cnt.Kmer))
		}
		last = cnt.Kmer
		if cnt.Data.Count < *min {
			continue
		}
		kept++
		err = kw.Write(last)
		if err != nil {
			break
		}
	}
	if err != io.EOF {
		util.Die(err)
	}
	fin.Close()
	fout.Close()
	pt.Done()

	if *del {
		fmt.Println("Deleting input file")
		util.Die(os.Remove(*inFile))
	}

	fmt.Println("Done")
}
