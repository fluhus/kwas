// Collects whitelisted kmers from a sample and outputs their hashes.
package main

import (
	"flag"
	"fmt"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmc"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
)

var (
	fin  = flag.String("i", "", "Input file")
	fwl  = flag.String("w", "", "Whitelist file")
	fout = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()
	fmt.Println("WL:", *fwl)
	fmt.Println("Input:", *fin)
	fmt.Println("Output:", *fout)

	wl, err := kmr.ReadFullKmersLines(*fwl)
	util.Die(err)
	fmt.Println(len(wl), "kmers in whitelist")

	// TODO(amit): Can I make this kmers instead of hashes?
	var hashes []uint64
	var buf []byte
	util.Die(kmc.KMC2(func(kmer []byte, count int) {
		buf = sequtil.DNATo2Bit(buf[:0], kmer)
		if !wl.Has(*(*kmr.FullKmer)(buf)) {
			return
		}
		hashes = append(hashes, util.Hash64(kmer))
	}, *fin, kmc.OptionK(kmr.K)))
	fmt.Println("Found", len(hashes), "kmers")

	aio.ToJSON(*fout, hashes)
}
