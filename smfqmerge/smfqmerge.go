// Merges the outputs of smfq.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/gmerge"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

var (
	inGlob  = flag.String("i", "", "Input file glob `pattern`")
	outFile = flag.String("o", "", "Output file")
	part    = flag.Int("p", 1, "1-based part number")
	nparts  = flag.Int("np", 1, "Total number of parts")
)

func main() {
	flag.Parse()

	fmt.Println("Input:", *inGlob)
	fmt.Println("Output:", *outFile)

	inFiles, err := filepath.Glob(*inGlob)
	util.Die(err)
	if len(inFiles) == 0 {
		util.Die(fmt.Errorf("found 0 files"))
	}
	fmt.Println("Found", len(inFiles), "files")
	sort.Strings(inFiles)
	inFiles, _ = util.ChooseStrings(inFiles, *part-1, *nparts)
	fmt.Println("Working on", len(inFiles))
	time.Sleep(time.Second)

	m := gmerge.NewMerger(
		func(r gkReader) (geneKmers, error) {
			return r.next()
		}, func(gk1, gk2 geneKmers) geneKmers {
			gk1.Kmers.AddSet(gk2.Kmers)
			return gk1
		}, func(gk1, gk2 geneKmers) int {
			return strings.Compare(gk1.Gene, gk2.Gene)
		},
	)

	fmt.Println("Opening files")
	for _, file := range inFiles {
		f, err := aio.Open(file)
		util.Die(err)
		j := json.NewDecoder(f)
		util.Die(m.Add(gkReader{j}))
	}
	fout, err := aio.Create(*outFile)
	util.Die(err)
	defer fout.Close()
	enc := json.NewEncoder(fout)

	fmt.Println("Merging")
	pt := progress.NewTimer()
	for {
		gk, err := m.Next()
		if err == io.EOF {
			break
		}
		util.Die(err)
		util.Die(enc.Encode(gk))
		pt.Inc()
	}
	pt.Done()
	fmt.Println("Done")
}

// A single entry from smfq.
type geneKmers struct {
	Gene  string
	Kmers sets.Set[int]
}

// A stream of geneKmers instances.
type gkReader struct {
	dec *json.Decoder
}

// Returns the next entry.
func (r gkReader) next() (geneKmers, error) {
	var gk geneKmers
	err := r.dec.Decode(&gk)
	return gk, err
}
