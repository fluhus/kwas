// Extracts kmers from a sample and saves them in a condensed format.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/kmc"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/slices"
)

var (
	inFile   = flag.String("i", "", "Input file")
	outFile  = flag.String("o", "", "Output file")
	selfTest = flag.Bool("t", false, "Make additional sanity tests, for debugging")
)

func main() {
	flag.Parse()

	fmt.Println("Reading kmers")
	pt := ptimer.New()
	var kmers []kmr.Kmer
	var buf kmr.Kmer
	err := kmc.KMC2(func(kmer []byte, count int) {
		sequtil.DNATo2Bit(buf[:0], kmer)
		kmers = append(kmers, buf)
	}, *inFile, kmc.OptionK(kmr.K), kmc.OptionThreads(2))
	pt.Done()
	util.Die(err)
	fmt.Println("Found", len(kmers), "kmers")

	fmt.Println("Sorting")
	pt = ptimer.New()
	slices.SortFunc(kmers, func(a, b kmr.Kmer) bool {
		return a.Less(b)
	})
	pt.Done()

	fmt.Println("Encoding")
	pt = ptimer.New()
	enc := encodeKmers(kmers)
	pt.Done()
	fmt.Printf("%d bytes, average %.1f bytes per kmer\n",
		len(enc), float64(len(enc))/float64(len(kmers)))

	if *selfTest {
		fmt.Println("Sanity testing (-t)")
		fmt.Print("  Uniqueness: ")
		fmt.Println(boolToOK(len(kmers) == len(sets.Set[kmr.Kmer]{}.Add(kmers...))))

		fmt.Print("  Decoded equals: ")
		dec, err := decodeKmers(enc)
		fmt.Println(boolToOK(slices.Equal(kmers, dec)))
		util.Die(err)
	}

	if *outFile != "" {
		fmt.Println("Writing")
		outf, err := aio.Create(*outFile)
		util.Die(err)
		_, err = outf.Write(enc)
		util.Die(err)
		outf.Close()
	}

	fmt.Println("Done")
}

func encodeKmers(kmers []kmr.Kmer) []byte {
	buf := bytes.NewBuffer(nil)
	w := kmr.NewWriter(buf)
	for _, kmer := range kmers {
		w.Write(kmer)
	}
	return buf.Bytes()
}

func decodeKmers(kmers []byte) ([]kmr.Kmer, error) {
	r := kmr.NewReader(bytes.NewBuffer(kmers))
	var result []kmr.Kmer
	for {
		kmer, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, kmer)
	}
	return result, nil
}

func boolToOK(ok bool) string {
	if ok {
		return "ok"
	}
	return "NOT OK"
}
