// Goes over KWAS results and splits kmers into significant and non-significant.
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"iter"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/util"
)

var (
	fin   = flag.String("i", "", "Input files glob")
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
	ikey := slices.Index(header, "key")
	if ikey == -1 {
		util.Die(fmt.Errorf("did not find key column"))
	}

	all, wrote := 0, 0
	pt := ptimer.NewFunc(func(i int) string {
		return fmt.Sprintf("%d/%d files done (%.1f%% significant)",
			i, len(files), util.Perc(wrote, all))
	})
	for _, f := range files {
		for row, err := range iterCSV(f, header) {
			if err != nil {
				return err
			}
			pval, err := strconv.ParseFloat(row[ip], 64)
			if err != nil {
				return err
			}
			if pval <= maxPval {
				if _, err := fmt.Fprintf(outSig, "%s\n", row[ikey]); err != nil {
					return err
				}
				wrote++
			} else {
				if _, err := fmt.Fprintf(outNSig, "%s\n", row[ikey]); err != nil {
					return err
				}
			}
			all++
			return nil
		}
		pt.Inc()
	}
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

// Iterates over lines in a CSV file.
// If the file's header does not match the given header, yields an error.
func iterCSV(file string, header []string) iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		f, err := aio.Open(file)
		if err != nil {
			yield(nil, err)
			return
		}
		defer f.Close()

		r := csv.NewReader(f)

		// Check that header matches.
		h, err := r.Read()
		if err != nil {
			yield(nil, err)
			return
		}
		if !slices.Equal(h, header) {
			yield(nil, fmt.Errorf("mismatching headers: %v %v", h, header))
			return
		}

		var row []string
		for row, err = r.Read(); err == nil; row, err = r.Read() {
			if !yield(row, err) {
				return
			}
		}
		if err != io.EOF {
			yield(nil, err)
		}
	}
}

// Counts the lines in the given files, minus the headers.
func countLinesFiles(files []string) (int, error) {
	pt := ptimer.NewMessage(fmt.Sprintf("{}/%d files done", len(files)))
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
