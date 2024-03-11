// Merges the outputs of smfq.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/gmerge"
	"github.com/fluhus/kwas/util"
)

var (
	inGlob  = flag.String("i", "", "Input file glob `pattern`")
	outFile = flag.String("o", "", "Output file")
	part    = flag.Int("p", 1, "1-based part number")
	nparts  = flag.Int("np", 1, "Total number of parts")
	del     = flag.Bool("d", false, "Delete input files")
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

	if *part != 1 || *nparts != 1 {
		fmt.Printf("Part %d/%d\n", *part, *nparts)
	}
	sort.Strings(inFiles)
	inFiles, _ = util.ChooseStrings(inFiles, *part-1, *nparts)
	fmt.Println("Found", len(inFiles), "input files")
	time.Sleep(time.Second)

	m := gmerge.NewMerger(
		func(gk1, gk2 geneKmers) int {
			return strings.Compare(gk1.Gene, gk2.Gene)
		},
		func(gk1, gk2 geneKmers) geneKmers {
			if len(gk1.Kmers) > len(gk2.Kmers) {
				gk1.Kmers.AddSet(gk2.Kmers)
				return gk1
			}
			gk2.Kmers.AddSet(gk1.Kmers)
			return gk2
		},
	)

	fmt.Println("Opening files")
	for _, file := range inFiles {
		f, err := aio.Open(file)
		util.Die(err)
		util.Die(m.Add(geneKmersIterator(f)))
	}
	fout, err := aio.Create(*outFile)
	util.Die(err)
	defer fout.Close()
	enc := json.NewEncoder(fout)

	fmt.Println("Merging")
	pt := ptimer.New()
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

	if *del {
		fmt.Println("Deleting input files")
		for _, f := range inFiles {
			util.Die(os.Remove(f))
		}
	}

	fmt.Println("Done")
}

// A single entry from smfq.
type geneKmers struct {
	Gene  string
	Kmers sets.Set[int]
}

func geneKmersIterator(r io.ReadCloser) func() (geneKmers, error) {
	dec := json.NewDecoder(r)
	return func() (geneKmers, error) {
		var gk geneKmers
		err := dec.Decode(&gk)
		if err != nil {
			r.Close()
		}
		return gk, err
	}
}
