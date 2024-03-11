// Splits HAS files by minimizer.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/lazy"
	"github.com/fluhus/kwas/util"
)

var (
	inFile  = flag.String("i", "", "Path to input file")
	outFile = flag.String("o", "", "Path to output files, with '*' for minimizer")
	del     = flag.Bool("d", false, "Delete existing output files")
	short   = flag.Int("n", 0, "Stop after n kmers (for debugging)")
	k       = flag.Int("k", 8, "Minimizer length")
	bufSize = flag.Int("b", 1<<17, "Write buffer size, higher means more RAM but faster")
)

func main() {
	util.Die(parseArgs())

	if *del {
		util.Die(deleteOutputFiles())
	}

	f, err := aio.Open(*inFile)
	util.Die(err)

	ws := map[uint64]*lazy.Writer{}
	bws := map[uint64]*bnry.Writer{}

	pt := ptimer.New()
	t := &kmr.HasTuple{}
	for {
		if *short > 0 && pt.N >= *short {
			break
		}
		if err = t.Decode(f); err != nil {
			break
		}
		mnz := minimizer(t)
		w := bws[mnz]
		if w == nil {
			ww := lazy.NewWriter(strings.ReplaceAll(*outFile, "*",
				fmt.Sprint(mnz)), *bufSize)
			ws[mnz] = ww
			w = bnry.NewWriter(ww)
			bws[mnz] = w
		}
		err = t.Encode(w)
		if err != nil {
			break
		}
		pt.Inc()
	}
	if err != io.EOF {
		util.Die(err)
	}
	for _, w := range ws {
		util.Die(w.Flush())
	}
	pt.Done()

	fmt.Println("Done")
}

// Parses program arguments.
func parseArgs() error {
	flag.Parse()
	if *inFile == "" {
		return fmt.Errorf("empty input path")
	}
	if *outFile == "" {
		return fmt.Errorf("empty output path")
	}
	if *bufSize < 4096 {
		return fmt.Errorf("bad buffer size: %d, want at least 4096", *bufSize)
	}
	return nil
}

// Returns the minimizer of the tuple's kmer.
func minimizer(tup *kmr.HasTuple) uint64 {
	return kmr.Minimizer(
		sequtil.DNAFrom2Bit(nil, tup.Kmer[:])[:kmr.K], *k)
}

// Removes all the files that match the input file pattern.
func deleteOutputFiles() error {
	files, err := filepath.Glob(*outFile)
	if err != nil {
		return err
	}
	fmt.Println("Deleting", len(files), "old files")
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}
