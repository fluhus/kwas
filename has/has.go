package main

import (
	"flag"
	"fmt"
	"slices"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/iterx"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
)

var (
	p       = flag.Int("p", 1, "Sample part number")
	np      = flag.Int("np", 1, "Number of sample parts")
	wlFile  = flag.String("i", "", "Input filtered count file")
	outFile = flag.String("o", "", "Output file")
	ff      = flag.String("f", "", "File containing input file names")
)

func main() {
	flag.Parse()

	files, err := util.ReadLines(aio.Open(*ff))
	util.Die(err)
	files, idx := util.ChooseStrings(files, *p-1, *np)
	fmt.Println("Found", len(files), "files to count")

	fmt.Println("Opening files")
	var streams []*iterx.Iter[kmr.Kmer]
	for _, file := range files {
		streams = append(streams, iterx.New(kmr.IterKmersFile(file)))
	}

	wlr := iterx.New(kmr.IterKmersFile(*wlFile))

	fout, err := aio.Create(*outFile)
	util.Die(err)
	wout := bnry.NewWriter(fout)

	fmt.Println("Reading")
	pt := ptimer.New()

	for _, cp := range kmr.Checkpoints(5000) {
		has := map[kmr.Kmer][]int{}
		for kmer, err := range wlr.Until(cp.Less) {
			util.Die(err)
			has[kmer] = nil
		}
		for i, s := range streams {
			for kmer, err := range s.Until(cp.Less) {
				util.Die(err)
				if haskmer, ok := has[kmer]; ok {
					has[kmer] = append(haskmer, idx[i])
				}
			}
		}
		var slice []kmr.HasTuple
		for k, v := range has {
			if len(v) > 0 {
				slice = append(slice, kmr.HasTuple{
					Kmer: k, Data: kmr.HasData{Samples: v}})
			}
		}
		slices.SortFunc(slice, func(a, b kmr.HasTuple) int {
			return a.Kmer.Compare(b.Kmer)
		})
		for _, kmer := range slice {
			util.Die(kmer.Encode(wout))
			pt.Inc()
		}
	}

	util.Die(fout.Close())
	pt.Done()

	fmt.Println("Done")
}
