package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/gostuff/ppln"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/graphs"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

var (
	input   = flag.String("i", "", "Input file glob pattern")
	output  = flag.String("o", "", "Output file")
	joutput = flag.String("j", "", "Optional output JSON file for cluster")
	nt      = flag.Int("t", 1, "Number of threads")
)

func main() {
	flag.Parse()

	fmt.Println("Threads:", *nt)

	kmers, err := loadKmersGlob(*input)
	util.Die(err)
	fmt.Println(len(kmers), "kmers")
	runtime.GC()

	type istring struct {
		i int
		s string
	}

	var expanded []istring
	for i, kmer := range kmers {
		e := string(sequtil.DNAFrom2Bit(nil, kmer.Kmer[:])[:kmr.K])
		expanded = append(expanded, istring{i, e})
		expanded = append(expanded, istring{i, sequtil.ReverseComplementString(e)})
	}
	sort.Slice(expanded, func(i, j int) bool {
		return expanded[i].s < expanded[j].s
	})

	graph := graphs.New(len(kmers))
	pt := progress.NewTimerMessasge(fmt.Sprint("{} out of ", len(expanded)))

	ppln.NonSerial(*nt,
		func(push func(istring), _ func() bool) error {
			for _, e := range expanded {
				push(e)
				pt.Inc()
			}
			return nil
		},
		func(e istring, push func([2]int), g int) error {
			const thr = 0.1
			const n = kmr.K / 2
			ei := e.i
			for p := 1; p <= n; p++ {
				ep := e.s[p:]
				s := sort.Search(len(expanded), func(i int) bool {
					return expanded[i].s >= ep
				})
				for ; s < len(expanded) && strings.HasPrefix(expanded[s].s, ep); s++ {
					es := expanded[s]
					si := es.i
					if e.i == es.i {
						continue
					}
					if util.JaccardDist(kmers[ei].Samples, kmers[si].Samples) < thr {
						push([2]int{ei, si})
					}
				}
			}
			return nil
		},
		func(a [2]int) error {
			graph.AddEdge(a[0], a[1])
			return nil
		},
	)
	pt.Done()
	fmt.Println(graph.NumEdges(), "edges (pairs that are close enough)")

	comps := graph.ConnectedComponents()
	fmt.Println(len(comps), "connected components,",
		util.Percf(len(comps), len(kmers), 0),
		"of kmers")
	slices.SortFunc(comps, func(a, b []int) bool {
		return len(a) > len(b)
	})
	var lens []int
	for _, v := range comps {
		lens = append(lens, len(v))
	}
	if len(lens) > 10 {
		fmt.Println("Biggest component sizes:", lens[:10])
	} else {
		fmt.Println("Component sizes:", lens)
	}
	fmt.Println("Component size quantiles:", ntiles(20, lens))

	fmt.Println("Finding centers")
	centers := make([]*kmr.HasTuple, 0, len(comps))
	pt = progress.NewTimer()
	ppln.Serial(*nt,
		func(push func([]int), _ func() bool) error {
			for _, comp := range comps {
				push(comp)
				pt.Inc()
			}
			return nil
		},
		func(a []int, i, g int) (*kmr.HasTuple, error) {
			compKmers := at(kmers, a)
			return compKmers[util.ArgMin(sqDistances(compKmers))], nil
		},
		func(a *kmr.HasTuple) error {
			centers = append(centers, a)
			return nil
		},
	)
	pt.Done()
	fmt.Println(len(centers), "centers")

	fmt.Println("Saving centers")
	fout, err := aio.Create(*output)
	util.Die(err)
	for _, c := range centers {
		util.Die(c.Encode(fout))
	}
	fout.Close()

	// Print to JSON for minimizer plots.
	if *joutput != "" {
		fmt.Println("Converting to JSON")
		var toJSON [][]string
		for _, comp := range comps {
			var c []string
			for _, kmer := range at(kmers, comp) {
				e := string(sequtil.DNAFrom2Bit(nil, kmer.Kmer[:])[:kmr.K])
				c = append(c, e)
			}
			toJSON = append(toJSON, c)
		}

		fmt.Println("Saving")
		aio.ToJSON(*joutput, toJSON)
	}
}

// Loads kmers from a HAS file.
func loadKmers(file string) ([]*kmr.HasTuple, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var result []*kmr.HasTuple
	for {
		tup := &kmr.HasTuple{}
		err = tup.Decode(f)
		if err != nil {
			break
		}
		result = append(result, tup)
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

// Loads kmer from a HAS file glob pattern.
func loadKmersGlob(file string) ([]*kmr.HasTuple, error) {
	files, err := filepath.Glob(file)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fs.ErrNotExist
	}
	fmt.Println("Reading kmers from", len(files), "files")
	var result []*kmr.HasTuple
	pt := progress.NewTimer()
	for _, f := range files {
		tups, err := loadKmers(f)
		if err != nil {
			return nil, err
		}
		result = append(result, tups...)
		pt.Inc()
	}
	pt.Done()
	return result, nil
}

// Returns the elements at the given indices.
func at[S ~[]E, E any, X ~[]D, D constraints.Integer](s S, idx X) S {
	result := make(S, 0, len(idx))
	for _, i := range idx {
		result = append(result, s[i])
	}
	return result
}

// Returns the sum of square distances from each tuple to the rest.
func sqDistances(kmers []*kmr.HasTuple) []float64 {
	result := make([]float64, len(kmers))
	for i, ki := range kmers {
		for j, kj := range kmers[i+1:] {
			d := util.JaccardDist(ki.Samples, kj.Samples)
			d *= d
			result[i] += d
			result[j+i+1] += d
		}
	}
	gnum.Mul1(result, 1.0/float64(len(kmers)-1))
	for i := range result {
		result[i] = math.Sqrt(result[i])
	}
	return result
}

// Returns the n-tiles of the given sorted slice.
func ntiles[S ~[]E, E constraints.Ordered](n int, s S) S {
	result := make(S, n+1)
	for i := 0; i <= n; i++ {
		j := int(math.Round(float64(i) / float64(n) * float64(len(s)-1)))
		result[i] = s[j]
	}
	return result
}
