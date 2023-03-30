// Creates k-mer profiles from fastq files.
package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/fluhus/biostuff/formats/fastq"
	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	fin     = flag.String("i", "", "Input file")
	fout    = flag.String("o", "", "Output file")
	fwl     = flag.String("w", "", "Whitelist file")
	flatten = flag.Bool("flatten", true, "If true, make all counts 0 or 1")
)

func main() {
	flag.Parse()

	fmt.Println("Reading whitelist")
	lines, err := util.ReadLines(aio.Open(*fwl))
	util.Die(err)
	wl := sets.Set[string]{}.Add(lines...)
	fmt.Println(len(wl))

	f, err := aio.Open(*fin)
	util.Die(err)
	r := fastq.NewReader(f)

	fmt.Println("Reading fastq")
	pt := progress.NewTimer()
	ps := kmr.ProfileSet[string]{}
	var fq *fastq.Fastq
	for fq, err = r.Read(); err == nil; fq, err = r.Read() {
		seq := fq.Sequence
		rc := sequtil.ReverseComplement(nil, seq)
		seqs, rcs := string(seq), string(rc)
		nt, rnt := util.NewNTracker(seq, kmr.K), util.NewNTracker(rc, kmr.K)
		for i := range seq[:len(seq)-kmr.K+1] {
			if !nt.NextN() && wl.Has(seqs[i:i+kmr.K]) {
				ps.Get(seqs[i:i+kmr.K]).Fill(seq, i)
			}
			if !rnt.NextN() && wl.Has(rcs[i:i+kmr.K]) {
				ps.Get(rcs[i:i+kmr.K]).Fill(rc, i)
			}
		}
		pt.Inc()
	}
	f.Close()
	pt.Done()

	if *flatten {
		fmt.Println("Flattening counts")
		pt := progress.NewTimer()
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
	pt = progress.NewTimer()
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
	pt = progress.NewTimer()
	keys := maps.Keys(ps)
	keys2bit := map[string]kmr.Kmer{}
	for _, k := range keys {
		keys2bit[k] = stringToKmer(k)
	}
	slices.SortFunc(keys, func(a, b string) bool {
		aa, bb := keys2bit[a], keys2bit[b]
		return bytes.Compare(aa[:], bb[:]) == -1
	})
	pt.Done()

	fmt.Println("Writing")
	pt = progress.NewTimer()
	out, err := aio.Create(*fout)
	util.Die(err)
	for _, key := range keys {
		util.Die((&kmr.ProfileTuple{
			Kmer: keys2bit[key],
			P:    *ps.Get(key),
			C:    ps.Get(key).SingleSampleCount(),
		}).Encode(out))
		pt.Inc()
	}
	out.Close()
	pt.Done()

	fmt.Println("Done")
}

func stringToKmer(s string) kmr.Kmer {
	kmer2bit := sequtil.DNATo2Bit(nil, []byte(s))
	kmer := kmr.Kmer(kmer2bit)
	return kmer
}
