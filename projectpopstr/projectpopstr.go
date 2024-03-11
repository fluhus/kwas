// Calculates the projection of a sample onto a principal component space.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/aio"
	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/gostuff/jio"
	"github.com/fluhus/gostuff/ppln/v2"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/progress"
	"github.com/fluhus/kwas/util"
)

const (
	threads = 4
	usePPLN = false
)

var (
	fin   = flag.String("i", "", "Input sample dump file")
	fcomp = flag.String("c", "", "Input components JSON file")
	fout  = flag.String("o", "", "Output JSON file")
)

func main() {
	flag.Parse()

	fmt.Println("Reading components")
	comps, err := loadComponents(*fcomp)
	util.Die(err)
	fmt.Println("Found", len(comps), "components")
	var vals []float64
	var bufs [][]byte
	if usePPLN {
		vals = make([]float64, len(comps)*threads)
		bufs = make([][]byte, threads)
	} else {
		vals = make([]float64, len(comps))
	}

	fmt.Println("Reading kmers")
	pt := progress.NewTimer()
	if usePPLN {
		err = ppln.NonSerial[kmr.Kmer, struct{}](threads,
			kmr.IterKmersFile(*fin),
			func(kmer kmr.Kmer, g int) (struct{}, error) {
				if g < 0 || g >= threads {
					panic(fmt.Sprintf("bad g: %v", g))
				}
				bufs[g] = sequtil.DNAFrom2Bit(bufs[g][:0], kmer[:])
				for i, c := range comps {
					vals[i*threads+g] += c[string(bufs[g])]
				}
				return struct{}{}, nil
			}, func(a struct{}) error {
				pt.Inc()
				return nil
			})
		util.Die(err)
	} else {
		var buf []byte
		for kmer, err := range kmr.IterKmersFile(*fin) {
			util.Die(err)
			buf = sequtil.DNAFrom2Bit(buf[:0], kmer[:])[:kmr.K]
			for i := range vals {
				vals[i] += comps[i][string(buf)]
			}
			pt.Inc()
		}
	}
	pt.Done()
	util.Die(err)

	if usePPLN {
		vals2 := make([]float64, len(comps))
		for i := range vals2 {
			vals2[i] = gnum.Sum(vals[i*threads : (i+1)*threads])
		}
		vals = vals2
	}

	fmt.Println("Writing projection")
	util.Die(jio.Save(*fout, vals))

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
