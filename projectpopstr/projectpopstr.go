// Calculates the projection of a sample onto a principal component space.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/fluhus/kmers/aio"
	"github.com/fluhus/kmers/kmc"
	"github.com/fluhus/kmers/kmr"
	"github.com/fluhus/kmers/progress"
	"github.com/fluhus/kmers/util"
)

var (
	fin   = flag.String("i", "", "Input sample fastq file")
	fcomp = flag.String("c", "", "Input components JSON file")
	fout  = flag.String("o", "", "Output JSON file")
)

func main() {
	flag.Parse()

	fmt.Println("Reading components")
	comps, err := loadComponents(*fcomp)
	util.Die(err)
	fmt.Println("Found", len(comps), "components")
	vals := make([]float64, len(comps))

	fmt.Println("Reading fastq")
	pt := progress.NewTimer()
	kmc.KMC2(func(kmer []byte, count int) {
		for i := range vals {
			vals[i] += comps[i][string(kmer)]
		}
	}, *fin, kmc.OptionK(kmr.K), kmc.OptionThreads(2))
	pt.Done()

	fmt.Println("Writing projection")
	util.Die(aio.ToJSON(*fout, vals))

	fmt.Println("Done")
}

// Loads the projection data produced by popstr.
func loadComponents(file string) ([]map[string]float64, error) {
	f, err := aio.Open(file)
	if err != nil {
		return nil, err
	}
	j := json.NewDecoder(f)
	var result []map[string]float64
	for {
		m := map[string]float64{}
		err := j.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}
