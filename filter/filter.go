// Filters out low-count kmers from KMC extraction.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

var (
	in  = flag.String("i", "", "Path to input file")
	out = flag.String("o", "", "Path to output file")
	min = flag.Uint64("n", 0, "Minimal count to leave a kmer in")
	del = flag.Bool("d", false, "Delete input file")
)

func main() {
	fmt.Println("Opening files")
	flag.Parse()
	fin, err := aio.Open(*in)
	util.Die(err)
	fout, err := aio.Create(*out)
	util.Die(err)

	fmt.Println("Filtering")
	cnt := &kmr.HasCount{}
	kept := 0
	var last kmr.FullKmer
	pt := progress.NewTimerFunc(func(i int) string {
		return fmt.Sprintf("Read %d, wrote %d (%d%%)\n", i, kept, kept*100/i)
	})
	for err = cnt.Decode(fin); err == nil; err = cnt.Decode(fin) {
		pt.Inc()
		if pt.N != 1 && bytes.Compare(last[:], cnt.Kmer[:]) != -1 {
			util.Die(fmt.Errorf("kmers not in order: %v %v", last, cnt.Kmer[:]))
		}
		last = cnt.Kmer
		if cnt.Count < *min {
			continue
		}
		kept++
		err = cnt.Encode(fout)
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
		util.Die(os.Remove(*in))
	}

	fmt.Println("Done")
}
