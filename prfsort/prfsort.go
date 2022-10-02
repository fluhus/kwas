// Sorts kmer profiles by their entropy.
package main

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
)

const (
	fin  = "final.prf"
	fout = "sorted.prf"
	fprc = "percentiles.prf.zst"

	tosefet = 1
)

func main() {
	fmt.Println("Reading kmers")
	t := time.Now()
	ps, err := loadProfiles(fin)
	util.Die(err)
	fmt.Println("Took", time.Since(t))
	fmt.Println(len(ps), "kmers")

	fmt.Println("Calculating entropy")
	t = time.Now()
	ents := map[*kmr.ProfileTuple]float64{}
	for _, p := range ps {
		ents[p] = avgEntropy(&p.P)
	}
	fmt.Println("Took", time.Since(t))

	fmt.Println("Sorting")
	t = time.Now()
	sort.Slice(ps, func(i, j int) bool {
		return ents[ps[i]] < ents[ps[j]]
	})
	fmt.Println("Took", time.Since(t))

	fmt.Println("Saving profiles")
	t = time.Now()
	util.Die(saveProfiles(fout, ps))

	var prc []*kmr.ProfileTuple
	for i := range make([]struct{}, 11) {
		idx := i * len(ps) / 10
		if idx == len(ps) {
			idx--
		}
		prc = append(prc, ps[idx])
	}
	util.Die(saveProfiles(fprc, prc))
	fmt.Println("Took", time.Since(t))

	// var entss []float64
	// for _, p := range ps {
	// 	entss = append(entss, ents[p])
	// }
	// aio.ToJSON("/tmp/amitmit/entropies.json", entss)
}

func loadProfiles(file string) ([]*kmr.ProfileTuple, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var result []*kmr.ProfileTuple
	for {
		tup := &kmr.ProfileTuple{}
		err := tup.Decode(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		result = append(result, tup)
	}
	return result, nil
}

func saveProfiles(file string, ps []*kmr.ProfileTuple) error {
	f, err := aio.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, p := range ps {
		if err := p.Encode(f); err != nil {
			return err
		}
	}
	return nil
}

func entropy(f [4]int64) float64 {
	sum := 0.0
	for _, v := range f {
		sum += float64(v + tosefet)
	}
	if sum == 0 {
		return 0
	}
	ent := 0.0
	for _, v := range f {
		if v == 0 {
			continue
		}
		p := float64(v+tosefet) / sum
		ent -= p * math.Log2(p)
	}
	return ent
}

func avgEntropy(p *kmr.Profile) float64 {
	ent := 0.0
	for i := range p {
		ent += entropy(p[i])
	}
	return (ent) / float64(len(p))
}
