package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/kmers/aio"
	"github.com/fluhus/kmers/kmr"
	"github.com/fluhus/kmers/util"
)

var (
	in  = flag.String("i", "", "Input file")
	out = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()

	fout, err := aio.Create(*out)
	util.Die(err)
	j := json.NewEncoder(fout)

	util.Die(kmr.IterTuplesFile(*in, &kmr.HasTuple{}, func(t *kmr.HasTuple) error {
		return j.Encode(map[string]any{
			"kmer":    string(sequtil.DNAFrom2Bit(nil, t.Kmer[:])[:kmr.K]),
			"samples": t.Samples,
		})
	}))
	util.Die(fout.Close())

	fmt.Println("Done")
}
