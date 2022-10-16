package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
)

var (
	in  = flag.String("i", "", "Input HAS file")
	out = flag.String("o", "", "Output JSON file")
)

func main() {
	flag.Parse()

	var fout io.WriteCloser
	if *out == "" {
		fout = os.Stdout
	} else {
		var err error
		fout, err = aio.Create(*out)
		util.Die(err)
	}
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
