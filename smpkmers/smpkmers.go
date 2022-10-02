// Randomly samples kmers from HAS files.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

// TODO(amit): Unite with smpkmers2.
// TODO(amit): Output HAS file and extract text kmers in a separate program?

var (
	fin  = flag.String("i", "", "Input text file glob template")
	fout = flag.String("o", "", "Output file")
	n    = flag.Int("n", 0, "Number of kmers to sample")
)

func main() {
	flag.Parse()
	r := util.NewReservoir[string](*n)
	files, err := filepath.Glob(*fin)
	util.Die(err)
	fmt.Println("Found", len(files), "files")

	pt := progress.NewTimer()
	for _, file := range files {
		lines, err := util.ReadLines(aio.Open(file))
		util.Die(err)
		for _, kmer := range lines {
			r.Add(kmer)
		}
		pt.Inc()
	}
	pt.Done()

	// Appending "" for a trailing new line.
	util.Die(os.WriteFile(*fout,
		[]byte(strings.Join(append(r.Sample, ""), "\n")), 0o644))
	fmt.Println("Done")
}
