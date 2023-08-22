package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"path/filepath"
	"sync"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/gostuff/jio"
	"github.com/fluhus/gostuff/minhash"
	"github.com/fluhus/gostuff/ppln"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/gostuff/snm"
	"github.com/fluhus/kwas/graphs"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

const (
	assertSamplesSorted = false // For debugging.

	indexK       = 50
	indexSearchK = 37

	// For int-hashing.
	prime uint64 = 10089886811898868001
	mask  uint64 = 7544360184296396679
)

var (
	input    = flag.String("i", "", "Input file glob pattern")
	output   = flag.String("o", "", "Output file")
	joutput  = flag.String("j", "", "Optional output JSON file for cluster")
	nt       = flag.Int("t", 1, "Number of threads")
	nSamples = flag.Int("n", 0, "Total number of samples")
)

func main() {
	flag.Parse()

	kmers, err := loadKmersGlob(*input)
	util.Die(err)

	var mhs [][]uint64
	var idx mhindex

	pt := ptimer.NewMessasge("{} kmers indexed")
	idx = mhindex{}
	ppln.Serial[int, []uint64](*nt,
		func(push func(int), stop func() bool) error {
			for i := range kmers {
				push(i)
			}
			return nil
		},
		func(a, i, g int) ([]uint64, error) {
			return intsMinHash(kmers[a].Data.Samples, indexK), nil
		},
		func(mh []uint64) error {
			mhs = append(mhs, slices.Clone(mh))
			idx.add(pt.N, mh)
			pt.Inc()
			return nil
		})
	pt.Done()

	graph := graphs.New(len(kmers))
	pt = ptimer.NewMessasge("{} kmers done")
	ptl := &sync.Mutex{}

	ppln.NonSerial[int, [2]int](*nt,
		func(push func(int), stop func() bool) error {
			for i := range kmers {
				push(i)
			}
			return nil
		},
		func(a int, push func([2]int), g int) error {
			const thr = 0.05
			for _, i := range idx.search(a, mhs[a], indexSearchK) {
				if util.JaccardDualDist(kmers[a].Data.Samples,
					kmers[i].Data.Samples, *nSamples) < thr {
					push([2]int{a, i})
				}
			}
			ptl.Lock()
			pt.Inc()
			ptl.Unlock()
			return nil
		}, func(a [2]int) error {
			graph.AddEdge(a[0], a[1])
			return nil
		})
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
	fmt.Println("Component size quantiles:", util.NTiles(20, lens))

	centers := make([]*kmr.HasTuple, 0, len(comps))
	pt = ptimer.NewMessasge("{} centers calculated")
	ppln.Serial(*nt,
		func(push func([]int), _ func() bool) error {
			for _, comp := range comps {
				push(comp)
				pt.Inc()
			}
			return nil
		},
		func(a []int, i, g int) (*kmr.HasTuple, error) {
			compKmers := snm.At(kmers, a)
			return compKmers[util.ArgMin(sqDistances(compKmers))], nil
		},
		func(a *kmr.HasTuple) error {
			centers = append(centers, a)
			return nil
		},
	)
	pt.Done()

	fmt.Println("Saving centers")
	fout, err := aio.Create(*output)
	util.Die(err)
	w := bnry.NewWriter(fout)
	for _, c := range centers {
		util.Die(c.Encode(w))
	}
	fout.Close()

	// Print to JSON for minimizer plots.
	if *joutput != "" {
		fmt.Println("Converting to JSON")
		var toJSON [][]string
		for _, comp := range comps {
			var c []string
			for _, kmer := range snm.At(kmers, comp) {
				e := string(sequtil.DNAFrom2Bit(nil, kmer.Kmer[:])[:kmr.K])
				c = append(c, e)
			}
			toJSON = append(toJSON, c)
		}

		fmt.Println("Saving")
		jio.Save(*joutput, toJSON)
	}

	fmt.Println("Done")
}

// Loads kmers from a HAS file.
func loadKmers(file string, pt *ptimer.Timer) ([]*kmr.HasTuple, error) {
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
		if len(tup.Data.Samples) > *nSamples {
			return nil, fmt.Errorf("kmer has %d samples but #samples (-n) is %d",
				len(tup.Data.Samples), *nSamples)
		}
		if assertSamplesSorted {
			for i := range tup.Data.Samples[1:] {
				if tup.Data.Samples[i] >= tup.Data.Samples[i+1] {
					return nil, fmt.Errorf(
						"kmer #%d: samples[%d] >= samples[%d]: %d >= %d",
						len(result), i, i+1,
						tup.Data.Samples[i], tup.Data.Samples[i+1])
				}
			}
		}
		result = append(result, tup)
		pt.Inc()
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
	pt := ptimer.NewMessasge("{} kmers loaded")
	for _, f := range files {
		tups, err := loadKmers(f, pt)
		if err != nil {
			return nil, err
		}
		result = append(result, tups...)
	}
	pt.Done()
	return result, nil
}

// Returns the sum of square distances from each tuple to the rest.
func sqDistances(kmers []*kmr.HasTuple) []float64 {
	result := make([]float64, len(kmers))
	for i, ki := range kmers {
		for j, kj := range kmers[i+1:] {
			d := util.JaccardDualDist(ki.Data.Samples, kj.Data.Samples, *nSamples)
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

// Min-hash index, for quick set lookup.
type mhindex map[uint64][]int

// Adds the given min-hashes and associates them with the given ID i.
func (m mhindex) add(i int, mh []uint64) {
	for _, h := range mh {
		m[h] = append(m[h], i)
	}
}

// Searches for ID's that share at least min min-hashes with i.
func (m mhindex) search(i int, mh []uint64, min int) []int {
	// TODO(amit): make the min = min(min, len)
	cnt := map[int]int{}
	for _, h := range mh {
		for _, ii := range m[h] {
			if ii <= i { // Avoid pair repetition and self.
				continue
			}
			cnt[ii]++
		}
	}
	var result []int
	for k, v := range cnt {
		if v >= min {
			result = append(result, k)
		}
	}
	return result
}

// Returns the k min-hashes for the ints in a.
func intsMinHash(a []int, k int) []uint64 {
	mh := minhash.New[uint64](k)
	for _, i := range a {
		mh.Push(intHash(i))
	}
	return mh.View()
}

// Hashes an integer.
func intHash[I constraints.Integer](i I) uint64 {
	return (uint64(i) * prime) ^ mask
}
