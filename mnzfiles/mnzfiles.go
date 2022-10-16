// Sorts minimizer split files by their sizes, large first.
// This is the step before clustering.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	inFiles = flag.String("i", "", "Input file glob pattern")
	outFile = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()

	fmt.Println("Looking up files")
	files, err := filepath.Glob(*inFiles)
	util.Die(err)
	fmt.Println("Found", len(files), "files")

	fmt.Println("Reading sizes")
	pt := progress.NewTimer()
	sizes := map[string]int64{}
	for _, file := range files {
		info, err := os.Stat(file)
		util.Die(err)
		sizes[info.Name()] += info.Size()
		pt.Inc()
	}
	pt.Done()

	fmt.Println(len(sizes), "minimizers")
	fmt.Println("Sorting")
	t := time.Now()
	names := maps.Keys(sizes)
	slices.SortFunc(names, func(a, b string) bool {
		return sizes[a] > sizes[b]
	})
	fmt.Println("Took", time.Since(t))

	fmt.Println("Writing output")
	t = time.Now()
	names = append(names, "")
	outf, err := aio.Create(*outFile)
	util.Die(err)
	outf.WriteString(strings.Join(names, "\n"))
	outf.Close()
	fmt.Println("Took", time.Since(t))
	fmt.Println("Wrote to:", *outFile)
	fmt.Println("Done")
}
