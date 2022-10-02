// Converts kmer profiles to JSON for creating dataframes.
package main

import (
	"encoding/json"
	"io"
	"math"
	"os"

	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/kwas/aio"
	"github.com/fluhus/kwas/kmr"
	"github.com/fluhus/kwas/util"
)

const (
	n = 60
)

func main() {
	t := &kmr.ProfileTuple{}
	var err error
	i := 0

	r := aio.WrapReader(os.Stdin)
	j := json.NewEncoder(os.Stdout)

	for err = t.Decode(r); err == nil; err = t.Decode(r) {
		m := tupleToMap(t)
		if jerr := j.Encode(m); jerr != nil {
			err = jerr
			break
		}
		i++
		if i == n {
			break
		}
	}
	if err != io.EOF {
		util.Die(err)
	}
}

func tupleToMap(tup *kmr.ProfileTuple) map[string][]float64 {
	pos := make([]float64, len(tup.P))
	for i := range pos {
		pos[i] = float64(i + 1)
	}

	fp := toFloats(tup)
	normalize(&fp)
	normalizeByEntropy(&fp)
	trsp := make([][]float64, 4)
	for i := range trsp {
		trsp[i] = make([]float64, len(tup.P))
	}

	for i := range fp {
		for j := range fp[i] {
			trsp[j][i] = fp[i][j]
		}
	}

	return map[string][]float64{
		"pos": pos,
		"A":   trsp[0],
		"C":   trsp[1],
		"G":   trsp[2],
		"T":   trsp[3],
		"n":   gnum.Cast[[]int64, []float64](tup.C[:]),
	}
}

type fprofile [len(kmr.Profile{})][4]float64

func toFloats(p *kmr.ProfileTuple) fprofile {
	var result fprofile
	for i := range p.P {
		for j := range p.P[i] {
			result[i][j] = float64(p.P[i][j])
		}
	}
	return result
}

func normalize(p *fprofile) {
	for i := range p {
		sm := gnum.Sum(p[i][:])
		if sm == 0 {
			continue
		}
		for j := range p[i] {
			p[i][j] /= sm
		}
	}
}

func entropy(f [4]float64) float64 {
	ent := 0.0
	for _, ff := range f {
		if ff == 0 {
			continue
		}
		ent -= ff * math.Log2(ff)
	}
	return ent
}

func normalizeByEntropy(p *fprofile) {
	for i := range p {
		factor := 2 - entropy(p[i])
		for j := range p[i] {
			p[i][j] *= factor
		}
	}
}
