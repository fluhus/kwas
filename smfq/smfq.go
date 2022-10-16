// Unites sequences from fastq with rnames from SAM.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"

	"github.com/fluhus/biostuff/formats/fastq"
	"github.com/fluhus/biostuff/formats/sam"
	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ppln"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	kmerIterGC = false // Run GC occasionally when loading kmers.
)

var (
	fqFile    = flag.String("f", "", "Fastq file `path`")
	samFile   = flag.String("s", "", "SAM file `path`")
	kmersFile = flag.String("k", "", "Kmers file `path`")
	outFile   = flag.String("o", "", "Output file `path`")
	nameRE    = flag.String("x", "", "Name `regex` to capture")
	nThreads  = flag.Int("t", 1, "Number of threads")
)

func main() {
	flag.Parse()

	var re *regexp.Regexp
	if *nameRE != "" {
		var err error
		re, err = regexp.Compile(*nameRE)
		util.Die(err)
	}

	genes := map[string]int{}
	geneSets := map[string]sets.Set[int]{}
	var wl map[kmert]int

	kit, err := newKmerIter(*kmersFile)
	util.Die(err)

	const nk = 25000000 // Batch size.

	for wl, err = kit.next(nk); err == nil; wl, err = kit.next(nk) {
		if *fqFile != "" {
			// For diamond files that don't include the query sequence.
			fmt.Println("Reading fastq and sam")
			ff, err := aio.Open(*fqFile)
			util.Die(err)
			fs, err := aio.Open(*samFile)
			util.Die(err)

			rf, rs := fastq.NewReader(ff), sam.NewReader(fs)
			pt := progress.NewTimerFunc(func(i int) string {
				return fmt.Sprint(i, " reads, ", len(genes), " genes")
			})
			type fqsm struct {
				fq *fastq.Fastq
				sm *sam.SAM
			}
			type geneidx struct {
				gene string
				idx  int
			}
			err = ppln.NonSerial(*nThreads,
				func(push func(a fqsm), stop func() bool) error {
					for {
						if stop() {
							break
						}
						fq, ef := rf.Read()
						sm, es := rs.Read()
						if ef != nil || es != nil {
							if ef == io.EOF && es == io.EOF {
								break
							}
							if err := util.NotExpectingEOF(ef); err != nil {
								return err
							}
							if err := util.NotExpectingEOF(es); err != nil {
								return err
							}
						}
						if !strings.HasPrefix(string(fq.Name), sm.Qname) {
							return fmt.Errorf("mismatching names: %q, %q",
								fq.Name, sm.Qname)
						}
						push(fqsm{fq, sm})
						pt.Inc()
					}
					return nil
				},
				func(a fqsm, push func(geneidx), g int) error {
					if a.sm.Rname == "*" {
						return nil
					}
					if re != nil {
						mch := re.FindString(a.sm.Rname)
						if mch == "" {
							return fmt.Errorf(
								"rname %q does not match name pattern %v",
								a.sm.Rname, re)
						}
						a.sm.Rname = mch
					}
					seq := a.fq.Sequence
					for i := range seq[kmr.K-1:] {
						kmer := seq[i : i+kmr.K]
						if idx, ok := wl[*(*kmert)(kmer)]; ok {
							push(geneidx{a.sm.Rname, idx})
						}
					}
					return nil
				},
				func(a geneidx) error {
					s, ok := geneSets[a.gene]
					if !ok {
						s = sets.Set[int]{}
						geneSets[a.gene] = s
					}
					s.Add(a.idx)
					return nil
				},
			)
			util.Die(err)
			pt.Done()
		} else {
			// For bowtie files that do include the query sequence.
			fmt.Println("Reading sam")
			fs, err := aio.Open(*samFile)
			util.Die(err)

			rs := sam.NewReader(fs)
			pt := progress.NewTimerFunc(func(i int) string {
				return fmt.Sprint(i, " reads, ", len(genes), " genes")
			})
			type fqsm struct {
				fq *fastq.Fastq
				sm *sam.SAM
			}
			type geneidx struct {
				gene string
				idx  int
			}
			err = ppln.NonSerial(*nThreads,
				func(push func(a fqsm), stop func() bool) error {
					for {
						if stop() {
							break
						}
						sm, err := rs.Read()
						if err != nil {
							if err == io.EOF {
								break
							}
							return err
						}
						push(fqsm{nil, sm})
						pt.Inc()
					}
					return nil
				},
				func(a fqsm, push func(geneidx), g int) error {
					if a.sm.Rname == "*" || a.sm.Mapq < 30 {
						return nil
					}
					if re != nil {
						mch := re.FindString(a.sm.Rname)
						if mch == "" {
							return fmt.Errorf(
								"rname %q does not match name pattern %v",
								a.sm.Rname, re)
						}
						a.sm.Rname = mch
					}
					seq := []byte(a.sm.Seq)
					for i := range seq[kmr.K-1:] {
						kmer := seq[i : i+kmr.K]
						if idx, ok := wl[*(*kmert)(kmer)]; ok {
							push(geneidx{a.sm.Rname, idx})
						}
					}
					return nil
				},
				func(a geneidx) error {
					s, ok := geneSets[a.gene]
					if !ok {
						s = sets.Set[int]{}
						geneSets[a.gene] = s
					}
					s.Add(a.idx)
					return nil
				},
			)
			util.Die(err)
			pt.Done()
		}
	}
	if err != io.EOF {
		util.Die(err)
	}
	fout, err := aio.Create(*outFile)
	util.Die(err)
	enc := json.NewEncoder(fout)
	keys := maps.Keys(geneSets)
	slices.Sort(keys)
	for _, k := range keys {
		j := struct {
			Gene  string
			Kmers []int
		}{k, sortSet(geneSets[k])}
		util.Die(enc.Encode(j))
	}
	fout.Close()
}

type kmert [kmr.K]byte

func (k kmert) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", k[:])), nil
}

func (k kmert) MarshalText() ([]byte, error) {
	return k[:], nil
}

type kmerIter struct {
	r io.ReadCloser
	s *bufio.Scanner
	i int
}

func newKmerIter(file string) (*kmerIter, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	return &kmerIter{f, bufio.NewScanner(f), 0}, nil
}

func (k *kmerIter) next(n int) (map[kmert]int, error) {
	if kmerIterGC {
		runtime.GC()
	}
	m := map[kmert]int{}
	for i := 0; i < n; i++ {
		if !k.s.Scan() {
			break
		}
		m[*(*kmert)(k.s.Bytes())] = k.i
		rc := sequtil.ReverseComplement(nil, k.s.Bytes()[:kmr.K])
		m[*(*kmert)(rc)] = k.i
		if kmerIterGC && i == n/2 {
			runtime.GC()
		}
		k.i++
	}
	if k.s.Err() != nil {
		return nil, k.s.Err()
	}
	if len(m) == 0 {
		return nil, io.EOF
	}
	if kmerIterGC {
		runtime.GC()
	}
	return m, nil
}

func sortSet[T constraints.Ordered](s sets.Set[T]) []T {
	k := maps.Keys(s)
	slices.Sort(k)
	return k
}
