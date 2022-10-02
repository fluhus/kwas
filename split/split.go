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
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

const (
	batch = 100000
)

var (
	inFile  = flag.String("i", "", "Path to input file")
	outFile = flag.String("o", "", "Path to output files, with '*' for minimizer")
	del     = flag.Bool("d", false, "Delete existing output files")
	short   = flag.Int("n", 0, "Stop after n kmers")
	k       = flag.Int("k", 8, "Minimizer length")
)

func main() {
	util.Die(parseArgs())
	fmt.Println("Batch size:", batch)

	if *del {
		util.Die(deleteOutputFiles())
	}

	f, err := aio.Open(*inFile)
	util.Die(err)

	pt := progress.NewTimer()
	var hs []*kmr.HasTuple
	for {
		pt.Inc()
		if *short > 0 && *short < pt.N {
			break
		}
		t := &kmr.HasTuple{}
		if err = t.Decode(f); err != nil {
			break
		}
		hs = append(hs, t)
		if len(hs) >= batch {
			err = writeByMinimizer(hs)
			hs = nil
			if err != nil {
				break
			}
		}
	}
	if err != io.EOF {
		util.Die(err)
	}
	util.Die(writeByMinimizer(hs))
	pt.Done()

	fmt.Println("Done")
}

// Parses program arguments.
func parseArgs() error {
	flag.Parse()
	if *inFile == "" {
		return fmt.Errorf("empty input path")
	}
	return nil
}

// Returns the minimizer of the tuple's kmer.
func minimizer(tup *kmr.HasTuple) uint64 {
	return kmr.Minimizer(
		sequtil.DNAFrom2Bit(nil, tup.Kmer[:])[:kmr.K], *k)
}

// Returns a map from minimizer to a list of (unchanged) tuples.
func splitByMinimizer(tups []*kmr.HasTuple) map[uint64][]*kmr.HasTuple {
	m := map[uint64][]*kmr.HasTuple{}
	for _, tup := range tups {
		mnz := minimizer(tup)
		m[mnz] = append(m[mnz], tup)
	}
	return m
}

// Writes the given tuples to files according to their minimizer.
func writeByMinimizer(tups []*kmr.HasTuple) error {
	m := splitByMinimizer(tups)
	for mnz, mtups := range m {
		f, err := aio.Append(strings.ReplaceAll(*outFile, "*", fmt.Sprint(mnz)))
		if err != nil {
			return err
		}
		for _, tup := range mtups {
			if err := tup.Encode(f); err != nil {
				return err
			}
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
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
