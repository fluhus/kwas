// Sorts kmer profiles by their entropy.
package main

import (
	"flag"
	"fmt"
	"math"
	"runtime/debug"
	"sort"

	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/ptimer"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
	"golang.org/x/exp/constraints"
)

const (
	additiveSmoothing = 1
)

var (
	fin  = flag.String("i", "", "Input profiles")
	fout = flag.String("o", "", "Sorted output")
	fprc = flag.String("p", "", "Percentile output")
)

func main() {
	debug.SetGCPercent(33)
	flag.Parse()

	fmt.Println("Reading kmers")
	ps, err := loadProfiles(*fin)
	util.Die(err)

	fmt.Println("Calculating entropy")
	pt := ptimer.New()
	ents := map[*kmr.ProfileTuple]float64{}
	for _, p := range ps {
		ents[p] = avgEntropy(&p.Data.P)
		pt.Inc()
	}
	pt.Done()

	fmt.Println("Sorting")
	pt = ptimer.New()
	sort.Slice(ps, func(i, j int) bool {
		return ents[ps[i]] < ents[ps[j]]
	})
	pt.Done()

	fmt.Println("Saving profiles")
	pt = ptimer.New()
	if *fout != "" {
		util.Die(saveProfiles(*fout, ps))
	}

	if *fprc != "" {
		var prc []*kmr.ProfileTuple
		for i := range 11 {
			idx := idiv(i*(len(ps)-1), 10)
			prc = append(prc, ps[idx])
		}
		util.Die(saveProfiles(*fprc, prc))
	}
	pt.Done()
	fmt.Println("Done")
}

// Loads profiles from a file.
func loadProfiles(file string) ([]*kmr.ProfileTuple, error) {
	pt := ptimer.NewMessage("{} kmers")
	defer pt.Done()

	var result []*kmr.ProfileTuple
	for tup, err := range kmr.IterTuplesFile[kmr.ProfileHandler](file) {
		if err != nil {
			return nil, err
		}
		result = append(result, tup.Clone())
		pt.Inc()
	}
	return result, nil
}

// Saves profiles to a file.
func saveProfiles(file string, ps []*kmr.ProfileTuple) error {
	f, err := aio.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bnry.NewWriter(f)
	for _, p := range ps {
		if err := p.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

// TODO(amit): Remove in favor of gnum.Entropy.
func entropy(f [4]int64) float64 {
	sum := 0.0
	for _, v := range f {
		sum += float64(v + additiveSmoothing)
	}
	if sum == 0 {
		return 0
	}
	ent := 0.0
	for _, v := range f {
		if v == 0 {
			continue
		}
		p := float64(v+additiveSmoothing) / sum
		ent -= p * math.Log2(p)
	}
	return ent
}

// Returns the average entropy with additive smoothing.
func avgEntropy(p *kmr.Profile) float64 {
	ent := 0.0
	for i := range p {
		ent += entropy(p[i])
	}
	return (ent) / float64(len(p))
}

// An integer division with rounding.
func idiv[T constraints.Integer](a, b T) T {
	return T(math.Round(float64(a) / float64(b)))
}
