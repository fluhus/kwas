// Counts kmers in fastq files.
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
	p   = flag.Int("p", 1, "Sample part number")
	np  = flag.Int("np", 1, "Number of sample parts")
	out = flag.String("o", "", "Output file")
	ff  = flag.String("f", "", "File with input files "+
		"(if omitted, inupt files are expected as arguments)")

	// Ignored for now.
	k  = flag.Int("k", 1, "Kmer part number")
	nk = flag.Int("nk", 1, "Number of kmer parts")
)

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
	var streams []*util.Unreader[kmr.Kmer]
	for _, file := range files {
		s, err := newUnreader(file)
		util.Die(err)
		streams = append(streams, s)
	}

	fout, err := aio.Create(*out)
	util.Die(err)
	wout := bnry.NewWriter(fout)

	fmt.Println("Creating checkpoints")
	checkpoints := kmr.Checkpoints(1000)

	fmt.Println("Counting")
	pt := ptimer.NewMessasge("{} kmers")

	for icp, cp := range checkpoints {
		counts := map[kmr.Kmer]int{}
		if pt.N > 0 {
			counts = make(map[kmr.Kmer]int, pt.N*3/icp/2)
		}
		for i, s := range streams {
			if s == nil {
				continue
			}
			err := s.ReadUntil(cp.Less, func(kmer kmr.Kmer) error {
				counts[kmer]++
				return nil
			})
			if err == io.EOF {
				streams[i] = nil
				continue
			}
			util.Die(err)
		}
		tuples := make([]kmr.CountTuple, 0, len(counts))
		for k, v := range counts {
			tuples = append(tuples,
				kmr.CountTuple{Kmer: k, Data: kmr.KmerCount{Count: v}})
		}
		slices.SortFunc(tuples, func(a, b kmr.CountTuple) bool {
			return a.Kmer.Less(b.Kmer)
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

func newUnreader(file string) (*util.Unreader[kmr.Kmer], error) {
	// TODO(amit): Close input file.
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	r := kmr.NewReader(f)
	return util.NewUnreader(r.Read), nil
}
