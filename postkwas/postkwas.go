// Goes over KWAS results and retains only results with low p-value.
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/util"
)

var (
	fin = flag.String("i", "", "Input files glob")
	// fout   = flag.String("o", "", "Output file")
	// invert = flag.Bool("n", false,
	// 	"Invert, retain only results with p-values above the threshold")
	fsig  = flag.String("s", "", "Significant kmers output file")
	fnsig = flag.String("n", "", "Non-significant kmers output file")
)

func main() {
	flag.Parse()
	files, err := filepath.Glob(*fin)
	util.Die(err)

	fmt.Println("Counting kmers")
	n, err := countLinesFiles(files)
	util.Die(err)
	fmt.Println("Found", n, "kmers")
	pval := 0.05 / float64(n)
	fmt.Println("P-value significance threshold:", pval)

	fmt.Println("Filtering kmers")
	fs, err := aio.Create(*fsig)
	util.Die(err)
	fn, err := aio.Create(*fnsig)
	util.Die(err)
	util.Die(filterByPvalFiles(files, pval, fs, fn))
	fs.Close()
	fn.Close()
}

// Writes a CSV (including a header) to out, with the lines where kmer p-value
// is at most maxPval.
func filterByPvalFiles(files []string, maxPval float64,
	outSig, outNSig io.Writer) error {
	header, err := readHeader(files[0])
	if err != nil {
		return err
	}
	ip := slices.Index(header, "kmer_pval")
	if ip == -1 {
		util.Die(fmt.Errorf("did not find kmer_pval column"))
	}

	ws := csv.NewWriter(outSig)
	wn := csv.NewWriter(outNSig)
	if err := ws.Write(header); err != nil {
		return err
	}
	if err := wn.Write(header); err != nil {
		return err
	}

	all, wrote := 0, 0
	pt := ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d/%d files done (%.1f%% significant)",
			i, len(files), util.Perc(wrote, all))
	})
	for _, f := range files {
		if err := iterCSV(f, header, func(row []string) error {
			pval, err := strconv.ParseFloat(row[ip], 64)
			if err != nil {
				return err
			}
			if pval <= maxPval {
				if err := ws.Write(row); err != nil {
					return err
				}
				wrote++
			} else {
				if err := wn.Write(row); err != nil {
					return err
				}
			}
			all++
			return nil
		}); err != nil {
			return err
		}
		pt.Inc()
	}
	ws.Flush()
	wn.Flush()
	pt.Done()
	return nil
}

// Returns the header of the given CSV file.
func readHeader(file string) ([]string, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).Read()
}

// Calls forEach for each line in a CSV file.
// If the file's header does not match the given header, returns an error.
func iterCSV(file string, header []string, forEach func([]string) error) error {
	f, err := aio.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check that header matches.
	r := csv.NewReader(f)
	h, err := r.Read()
	if err != nil {
		return err
	}
	if !slices.Equal(h, header) {
		return fmt.Errorf("mismatching headers: %v %v", h, header)
	}

	var row []string
	for row, err = r.Read(); err == nil; row, err = r.Read() {
		if err := forEach(row); err != nil {
			return err
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

// Counts the lines in the given files, minus the headers.
func countLinesFiles(files []string) (int, error) {
	pt := ptimer.NewMessasge(fmt.Sprintf("{}/%d files done", len(files)))
	n := 0
	for _, f := range files {
		nf, err := countLines(f)
		if err != nil {
			return 0, err
		}
		if nf == 0 {
			return 0, fmt.Errorf("found 0 lines in: %s", f)
		}
		n += nf - 1
		pt.Inc()
	}
	pt.Done()
	return n, nil
}

// Counts lines in a single file.
func countLines(file string) (int, error) {
	f, err := aio.Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n++
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	return n, nil
}
