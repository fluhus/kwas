// Command merge merges sorted streams of kmer tuples.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
)

var (
	p   = flag.Int("p", 1, "Part number, 1-based")
	np  = flag.Int("np", 1, "Total number of parts")
	del = flag.Bool("d", false, "Delete input files when done")
	in  = flag.String("i", "", "Input file pattern")
	out = flag.String("o", "", "Output file")
	typ = flag.String("t", "", "Type of files being merged")
)

func main() {
	util.Die(parseArgs())
	fmt.Printf("Running part %v/%v\n", *p, *np)
	fmt.Println("Input file: ", *in)
	fmt.Println("Output file:", *out)

	files, err := filepath.Glob(*in)
	util.Die(err)
	nfiles := len(files)
	sort.Strings(files)
	files, _ = util.ChooseStrings(files, *p-1, *np)

	fmt.Printf("Reading %v files out of %v\n", len(files), nfiles)

	switch *typ {
	case "cnt":
		err = merge[kmr.CountHandler](files)
	case "has":
		err = merge[kmr.HasHandler](files)
	case "prf":
		err = merge[kmr.ProfileHandler](files)
	default:
		err = fmt.Errorf("unsupported type: %q", *typ)
	}
	util.Die(err)

	if *del {
		fmt.Println("Removing input files")
		for _, file := range files {
			util.Die(os.Remove(file))
		}
		fmt.Println(len(files), "files deleted")
	}

	fmt.Println("Done")
}

// Merges the given files using the given zero tuple.
func merge[H kmr.KmerDataHandler[T], T any](files []string) error {
	m := kmr.NewMerger[H]()
	for _, file := range files {
		if err := m.Add(kmr.IterTuplesFile[H](file)); err != nil {
			return err
		}
	}

	fmt.Println("Writing to:", *out)
	fout, err := aio.Create(*out)
	if err != nil {
		return err
	}
	defer fout.Close()
	return m.Dump(fout)
}

func parseArgs() error {
	flag.Parse()
	if *in == "" {
		return fmt.Errorf("empty input path")
	}
	if *out == "" {
		return fmt.Errorf("empty output path")
	}
	return nil
}
