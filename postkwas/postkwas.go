// Goes over KWAS results and retains only results with low p-value.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/slices"
)

var (
	thr    = flag.Float64("p", 0, "P-value threshold")
	fin    = flag.String("i", "", "Input file")
	fout   = flag.String("o", "", "Output file")
	invert = flag.Bool("n", false,
		"Invert, retain only results with p-values above the threshold")
)

func main() {
	flag.Parse()

	fmt.Fprintln(os.Stderr, "Starting")
	fmt.Fprintln(os.Stderr, "Pval threshold:", *thr)

	fi, err := aio.Open(*fin)
	util.Die(err)
	r := csv.NewReader(fi)
	head, err := r.Read()
	util.Die(err)

	ip := slices.Index(head, "has_pval")
	if ip == -1 {
		util.Die(fmt.Errorf("did not find has_pval"))
	}

	fo, err := aio.Create(*fout)
	util.Die(err)
	w := csv.NewWriter(fo)
	w.Write(head)

	sig := 0
	var pvals []float64
	var row []string
	for row, err = r.Read(); err == nil; row, err = r.Read() {
		pval, err := strconv.ParseFloat(row[ip], 64)
		util.Die(err)
		pvals = append(pvals, pval)
		if (!*invert && pval <= *thr) || (*invert && pval > *thr) {
			util.Die(w.Write(row))
			sig++
		}
	}
	if err != io.EOF {
		util.Die(err)
	}
	w.Flush()
	util.Die(w.Error())
	util.Die(fo.Close())
	fmt.Fprintln(os.Stderr, "Read", len(pvals), "rows")

	slices.Sort(pvals)
	ip, _ = slices.BinarySearch(pvals, *thr)
	fmt.Fprintln(os.Stderr, ip, "significant",
		util.Percf(ip, len(pvals), 0))

	nt := util.NTiles(10, pvals)
	for _, n := range nt {
		fmt.Fprintf(os.Stderr, "%.2g ", n)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Done")
}
