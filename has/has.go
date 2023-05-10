package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/slices"
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
	var streams []*util.Unreader[kmr.Kmer]
	for _, file := range files {
		s, err := newUnreader(file)
		util.Die(err)
		streams = append(streams, s)
	}

	wlr, err := newUnreader(*wlFile)
	util.Die(err)

	fout, err := aio.Create(*outFile)
	util.Die(err)
	wout := bnry.NewWriter(fout)

	fmt.Println("Reading")
	pt := ptimer.New()

	for _, cp := range kmr.Checkpoints(5000) {
		has := map[kmr.Kmer][]int{}
		err := wlr.ReadUntil(cp.Less, func(kmer kmr.Kmer) error {
			has[kmer] = nil
			return nil
		})
		if err != io.EOF {
			util.Die(err)
		}
		for i, s := range streams {
			err := s.ReadUntil(cp.Less, func(kmer kmr.Kmer) error {
				if haskmer, ok := has[kmer]; ok {
					has[kmer] = append(haskmer, idx[i])
				}
				return nil
			})
			if err != io.EOF {
				util.Die(err)
			}
		}
		var slice []kmr.HasTuple
		for k, v := range has {
			if len(v) > 0 {
				slice = append(slice, kmr.HasTuple{Kmer: k, Data: kmr.KmerHas{Samples: v}})
			}
		}
		slices.SortFunc(slice, func(a, b kmr.HasTuple) bool {
			return a.Kmer.Less(b.Kmer)
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

// Returns an unreader of kmer files.
func newUnreader(file string) (*util.Unreader[kmr.Kmer], error) {
	// TODO(amit): Close input file.
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	r := kmr.NewReader(f)
	return util.NewUnreader(r.Read), nil
}
