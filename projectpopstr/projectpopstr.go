// Calculates the projection of a sample onto a principal component space.
// Takes in the output of grabforpopstr.
package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/fluhus/gostuff/sets"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/util"
)

var (
	fin   = flag.String("i", "", "Input sample file")
	fcomp = flag.String("c", "", "Input components file")
	fout  = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()
	var kmersSlice []uint64
	util.Die(aio.FromJSON(*fin, &kmersSlice))
	kmers := sets.Set[uint64]{}.Add(kmersSlice...)
	fmt.Println("Sample has", len(kmers), "kmers")

	var vals []float64
	comps, err := aio.Open(*fcomp)
	util.Die(err)
	j := json.NewDecoder(comps)
	for {
		comp := map[uint64]float64{}
		err = j.Decode(&comp)
		if err != nil {
			break
		}
		val := 0.0
		for kmer := range kmers {
			v, ok := comp[kmer]
			if !ok {
				panic(fmt.Sprintf("Kmer in sample but not in component: %d",
					kmer))
			}
			val += v
		}
		vals = append(vals, val)
	}
	fmt.Println("Calculated", len(vals), "components")
	fmt.Println(vals)
	aio.ToJSON(*fout, vals)
	fmt.Println("Done")
}
