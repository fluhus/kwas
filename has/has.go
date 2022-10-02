package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmc"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

var (
	p    = flag.Int("p", 0, "Sample part number")
	np   = flag.Int("np", 0, "Number of sample parts")
	inf  = flag.String("i", "", "Input filtered count file")
	outf = flag.String("o", "", "Output file")
	ff   = flag.String("f", "", "File containing input file names")
)

func main() {
	flag.Parse()

	fmt.Println("Reading kmer whitelist")
	kmers, err := readKmerWhitelist()
	util.Die(err)
	fmt.Println("Found", len(kmers), "kmers")

	// TODO(amit): Allow input files in args like in count.
	files, err := util.ReadLines(aio.Open(*ff))
	util.Die(err)
	files, idx := util.ChooseStrings(files, *p-1, *np)
	fmt.Println("Found", len(files), "files to count")

	for i, file := range files {
		fmt.Printf("Opening %v/%v: %s\n", i+1, len(files), file)
		t := time.Now()
		var buf kmr.FullKmer
		util.Die(kmc.KMC2(func(kmer []byte, count int) {
			sequtil.DNATo2Bit(buf[:0], kmer)
			if _, ok := kmers[buf]; !ok {
				return
			}
			kmers[buf] = append(kmers[buf], idx[i])
		}, file, kmc.OptionK(kmr.K)))
		fmt.Println("Took", time.Since(t))
	}

	fmt.Println("Collecting")
	keys := make([]kmr.FullKmer, 0, len(kmers))
	for kmer, idx := range kmers {
		if idx == nil {
			continue
		}
		keys = append(keys, kmer)
	}
	fmt.Println(len(keys), "kmers")

	fmt.Println("Sorting")
	pt := progress.NewTimer()
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) == -1
	})
	pt.Done()

	fmt.Println("Writing")
	fout, err := aio.Create(*outf)
	pt = progress.NewTimer()
	util.Die(err)
	for _, kmer := range keys {
		tup := &kmr.HasTuple{Kmer: kmer, Samples: kmers[kmer]}
		util.Die(tup.Encode(fout))
		pt.Inc()
	}
	pt.Done()
	fout.Close()

	fmt.Println("Done")
}

// Reads the kmer whitelist into a ready map.
func readKmerWhitelist() (map[kmr.FullKmer][]int, error) {
	f, err := aio.Open(*inf)
	if err != nil {
		return nil, err
	}
	m := map[kmr.FullKmer][]int{}
	cnt := &kmr.HasCount{}
	pt := progress.NewTimer()
	for err = cnt.Decode(f); err == nil; err = cnt.Decode(f) {
		pt.Inc()
		m[cnt.Kmer] = nil
	}
	pt.Done()
	if err != io.EOF {
		return nil, err
	}
	return m, nil
}
