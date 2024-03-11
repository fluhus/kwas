// Unites sequences from fastq with rnames from SAM.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"iter"
	"regexp"
	"runtime/debug"

	"github.com/fluhus/biostuff/formats/bioiter/v2"
	"github.com/fluhus/biostuff/formats/sam"
	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/ppln/v2"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/gostuff/snm"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
)

var (
	samFile   = flag.String("i", "", "Input file `path` (sam or diamond)")
	kmersFile = flag.String("k", "", "Kmers file `path`")
	outFile   = flag.String("o", "", "Output file `path`")
	nameRE    = flag.String("x", "", "Name `regex` to capture")
	nThreads  = flag.Int("t", 1, "Number of threads")
	isDiamond = flag.Bool("d", false, "Input is a diamond file")
)

func main() {
	debug.SetGCPercent(33)
	flag.Parse()

	var re *regexp.Regexp
	if *nameRE != "" {
		var err error
		re, err = regexp.Compile(*nameRE)
		util.Die(err)
	}

	geneSets := snm.NewDefaultMap(func(s string) sets.Set[int] {
		return sets.Set[int]{}
	})

	const nk = 25000000 // Batch size.

	round := 0
	for wl, err := range iterKmersBatch(*kmersFile, nk) {
		util.Die(err)
		round++
		if *isDiamond {
			fmt.Printf("Reading input (round %d)\n", round)
			pt := ptimer.NewFunc(func(i int) string {
				return fmt.Sprint(i, " reads")
			})
			type geneidx struct {
				gene string
				idx  int
			}
			err = ppln.NonSerial(*nThreads,
				iterDiamondFile(*samFile),
				func(a diamondLine, g int) ([]geneidx, error) {
					if re != nil {
						mch := re.FindString(a.rid)
						if mch == "" {
							return nil, fmt.Errorf(
								"rname %q does not match name pattern %v",
								a.rid, re)
						}
						a.rid = mch
					}
					seq := []byte(a.qseq)
					var result []geneidx
					for i := range seq[kmr.K-1:] {
						kmer := seq[i : i+kmr.K]
						if idx, ok := wl[*(*kmert)(kmer)]; ok {
							result = append(result, geneidx{a.rid, idx})
						}
					}
					return result, nil
				},
				func(a []geneidx) error {
					for _, gi := range a {
						geneSets.Get(gi.gene).Add(gi.idx)
					}
					pt.Inc()
					return nil
				},
			)
			util.Die(err)
			pt.Done()
		} else {
			fmt.Printf("Reading input (round %d)\n", round)
			pt := ptimer.NewFunc(func(i int) string {
				return fmt.Sprint(i, " reads")
			})
			type geneidx struct {
				gene string
				idx  int
			}
			err = ppln.NonSerial(*nThreads,
				bioiter.SAM(*samFile),
				func(sm *sam.SAM, g int) ([]geneidx, error) {
					if sm.Rname == "*" || sm.Mapq < 30 {
						return nil, nil
					}
					if re != nil {
						mch := re.FindString(sm.Rname)
						if mch == "" {
							return nil, fmt.Errorf(
								"rname %q does not match name pattern %v",
								sm.Rname, re)
						}
						sm.Rname = mch
					}
					seq := []byte(sm.Seq)
					var result []geneidx
					for i := range seq[kmr.K-1:] {
						kmer := seq[i : i+kmr.K]
						if idx, ok := wl[*(*kmert)(kmer)]; ok {
							result = append(result, geneidx{sm.Rname, idx})
						}
					}
					return result, nil
				},
				func(a []geneidx) error {
					for _, ig := range a {
						geneSets.Get(ig.gene).Add(ig.idx)
					}
					pt.Inc()
					return nil
				},
			)
			util.Die(err)
			pt.Done()
		}
	}
	fmt.Println("Mapped to", len(geneSets.M), "groups")
	fout, err := aio.Create(*outFile)
	util.Die(err)
	enc := json.NewEncoder(fout)
	for _, k := range sortedKeys(geneSets.M) {
		j := struct {
			Gene  string
			Kmers []int
		}{k, sortedKeys(geneSets.Get(k))}
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

func iterKmersBatch(file string, n int) iter.Seq2[map[kmert]int, error] {
	return func(yield func(map[kmert]int, error) bool) {
		f, err := aio.Open(file)
		if err != nil {
			yield(nil, err)
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		i := 0
		for {
			m := make(map[kmert]int, n*2)
			for range n {
				if !sc.Scan() {
					break
				}
				m[kmert(sc.Bytes())] = i
				rc := sequtil.ReverseComplement(nil, sc.Bytes()[:kmr.K])
				m[kmert(rc)] = i
				i++
			}
			if sc.Err() != nil {
				yield(nil, sc.Err())
				return
			}
			if len(m) == 0 {
				return
			}
			if !yield(m, nil) {
				return
			}
		}
	}
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	return snm.Sorted(maps.Keys(m))
}

type diamondLine struct {
	qid, rid, qseq string
}

func iterDiamondFile(file string) iter.Seq2[diamondLine, error] {
	return func(yield func(diamondLine, error) bool) {
		f, err := aio.Open(file)
		if err != nil {
			yield(diamondLine{}, err)
			return
		}
		defer f.Close()
		r := csv.NewReader(f)
		r.Comma = '\t'
		r.Comment = '#'
		for {
			line, err := r.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(diamondLine{}, err)
				return
			}
			if len(line) != 3 {
				yield(diamondLine{}, fmt.Errorf("bad number of fields in: %v", line))
				return
			}
			if !yield(diamondLine{line[0], line[1], line[2]}, nil) {
				return
			}
		}
	}
}
