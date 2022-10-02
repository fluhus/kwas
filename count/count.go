// Counts kmers in fastq files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"time"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmc"
	"github.com/fluhus/kwas/kmr"
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

	m := map[kmr.FullKmer]int{}
	var buf []byte

	for i, file := range files {
		fmt.Printf("Opening %v/%v: %s\n", i+1, len(files), file)
		t := time.Now()
		util.Die(kmc.KMC2(func(kmer []byte, count int) {
			if util.Hash64(kmer)%uint64(*nk) != uint64(*k-1) {
				return
			}
			buf = sequtil.DNATo2Bit(buf[:0], kmer)
			m[*(*kmr.FullKmer)(buf)]++
		}, file, kmc.OptionK(kmr.K)))
		fmt.Println("Took", time.Since(t), "len", len(m))
	}

	fmt.Println("Organizing counts")
	t := time.Now()
	tuples := make([]*kmr.HasCount, 0, len(m))
	for kmer, count := range m {
		tuples = append(tuples, &kmr.HasCount{
			Kmer: kmer, Count: uint64(count)})
	}
	fmt.Println("Took", time.Since(t))
	fmt.Println("Sorting")
	t = time.Now()
	slices.SortFunc(tuples, func(a, b *kmr.HasCount) bool {
		return bytes.Compare(a.Kmer[:], b.Kmer[:]) == -1
	})
	fmt.Println("Took", time.Since(t))
	fmt.Println("Writing")
	t = time.Now()
	fout, err := aio.Create(*out)
	util.Die(err)
	for _, tup := range tuples {
		err = tup.Encode(fout)
		if err != nil {
			break
		}
	}
	fout.Close()
	util.Die(err)
	fmt.Println("Took", time.Since(t))
	fmt.Println("Done")
}
