// Counts kmers in fastq files.
package main

import (
	"flag"
	"fmt"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/iterx"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/slices"
)

var (
	p   = flag.Int("p", 1, "Sample part number")
	np  = flag.Int("np", 1, "Number of sample parts")
	k   = flag.Int("k", 1, "Kmer part number")
	nk  = flag.Int("nk", 1, "Number of kmer parts")
	out = flag.String("o", "", "Output file")
	ff  = flag.String("f", "", "File with input files "+
		"(if omitted, inupt files are expected as arguments)")
)

// TODO(amit): Make input file a glob pattern?

func main() {
	flag.Parse()

	var files []string
	if *ff != "" {
		var err error
		files, err = util.ReadLines(aio.Open(*ff))
		util.Die(err)
	} else {
		files = flag.Args()
	}
	if len(files) == 0 {
		util.Die(fmt.Errorf("got no input files"))
	}
	files, _ = util.ChooseStrings(files, *p-1, *np)
	fmt.Println("Found", len(files), "files to count")

	fmt.Println("Opening files")
	var streams []*iterx.Iter[kmr.Kmer]
	for _, file := range files {
		streams = append(streams, iterx.New(kmr.IterKmersFile(file)))
	}

	fout, err := aio.Create(*out)
	util.Die(err)
	wout := bnry.NewWriter(fout)

	fmt.Println("Creating checkpoints")
	checkpoints := kmr.Checkpoints(1000)

	fmt.Println("Counting")
	pt := ptimer.NewMessage("{} kmers")

	for icp, cp := range checkpoints {
		counts := map[kmr.Kmer]int{}
		if pt.N > 0 {
			counts = make(map[kmr.Kmer]int, pt.N*3/icp/2)
		}
		for _, s := range streams {
			if s == nil {
				continue
			}
			for kmer, err := range s.Until(cp.Less) {
				util.Die(err)
				if *nk != 1 {
					if util.Hash64(kmer[:])%uint64(*nk) != uint64(*k-1) {
						continue
					}
				}
				counts[kmer]++
			}
		}
		tuples := make([]kmr.CountTuple, 0, len(counts))
		for k, v := range counts {
			tuples = append(tuples,
				kmr.CountTuple{Kmer: k, Data: kmr.CountData{Count: v}})
		}
		slices.SortFunc(tuples, func(a, b kmr.CountTuple) int {
			return a.Kmer.Compare(b.Kmer)
		})
		for _, t := range tuples {
			t.Encode(wout)
			pt.Inc()
		}
	}
	fout.Close()
	pt.Done()

	fmt.Println("Done")
}
