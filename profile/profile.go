// Creates k-mer profiles from fastq files.
package main

import (
	"flag"
	"fmt"
	"runtime/debug"

	"github.com/fluhus/biostuff/formats/bioiter/v2"
	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/maps"
)

var (
	fin     = flag.String("i", "", "Input file")
	fout    = flag.String("o", "", "Output file")
	fwl     = flag.String("w", "", "Whitelist file")
	flatten = flag.Bool("flatten", true, "If true, make all counts 0 or 1")
)

func main() {
	debug.SetGCPercent(33)
	flag.Parse()

	fmt.Println("Reading whitelist")
	lines, err := util.ReadLines(aio.Open(*fwl))
	util.Die(err)
	wl := sets.Set[string]{}.Add(lines...)
	fmt.Println(len(wl))

	fmt.Println("Reading fastq")
	pt := ptimer.New()
	ps := kmr.ProfileSet[string]{}
	for fq, err := range bioiter.Fastq(*fin) {
		util.Die(err)
		seq := fq.Sequence
		for i, ss := range util.NonNSubseqsString(string(seq), kmr.K) {
			if wl.Has(ss) {
				ps.Get(ss).Fill(seq, i)
			}
		}
		rc := sequtil.ReverseComplement(nil, seq)
		for i, ss := range util.NonNSubseqsString(string(rc), kmr.K) {
			if wl.Has(ss) {
				ps.Get(ss).Fill(rc, i)
			}
		}
		pt.Inc()
	}
	pt.Done()

	if *flatten {
		fmt.Println("Flattening counts")
		pt := ptimer.New()
		for _, p := range ps {
			for i := range p {
				for j := range p[i] {
					if p[i][j] > 1 {
						p[i][j] = 1
					}
				}
			}
			pt.Inc()
		}
		pt.Done()
	}

	fmt.Println("Validating")
	pt = ptimer.New()
	for kmer, p := range ps {
		const from = (len(p) - kmr.K) / 2
		for i, pos := range p[from : from+kmr.K] {
			zeros := 0
			for _, x := range pos {
				if x == 0 {
					zeros++
				}
			}
			if zeros != 3 {
				panic(fmt.Sprintf("position %d in kmer %q has differing counts: %v",
					i, kmer, pos))
			}
		}
		pt.Inc()
	}
	pt.Done()

	fmt.Println("Sorting")
	pt = ptimer.New()
	keys := snm.Sorted(maps.Keys(ps))
	pt.Done()

	fmt.Println("Writing")
	pt = ptimer.New()
	out, err := aio.Create(*fout)
	util.Die(err)
	w := bnry.NewWriter(out)
	for _, key := range keys {
		util.Die((&kmr.ProfileTuple{
			Kmer: stringToKmer(key),
			Data: &kmr.ProfileData{
				P: *ps.Get(key),
				C: ps.Get(key).SingleSampleCount(),
			},
		}).Encode(w))
		pt.Inc()
	}
	out.Close()
	pt.Done()

	fmt.Println("Done")
}

// Turns a string into a 2-bit kmer.
func stringToKmer(s string) kmr.Kmer {
	if len(s) != kmr.K {
		panic(fmt.Sprintf("bad string length: %v, want %v",
			len(s), kmr.K))
	}
	kmer2bit := sequtil.DNATo2Bit(nil, []byte(s))
	kmer := kmr.Kmer(kmer2bit)
	return kmer
}
