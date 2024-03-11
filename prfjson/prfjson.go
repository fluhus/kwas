// Converts kmer profiles to JSON for creating dataframes.
package main

import (
	"bufio"
	"encoding/json"
	"math"
	"os"

	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/kwas/kmr/v2"
	"github.com/fluhus/kwas/util"
)

const (
	n = 60
)

func main() {
	i := 0
	j := json.NewEncoder(os.Stdout)
	for t, err := range kmr.IterTuplesReader[kmr.ProfileHandler](
		bufio.NewReader(os.Stdin)) {
		util.Die(err)
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
}

// Converts a profile to a JSON-friendly map.
func tupleToMap(tup *kmr.ProfileTuple) map[string][]float64 {
	pos := make([]float64, len(tup.Data.P))
	for i := range pos {
		pos[i] = float64(i + 1)
	}

	fp := toFloats(tup)
	normalize(&fp)
	normalizeByEntropy(&fp)
	trsp := make([][]float64, 4)
	for i := range trsp {
		trsp[i] = make([]float64, len(tup.Data.P))
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
		"n":   gnum.Cast[[]int64, []float64](tup.Data.C[:]),
	}
}

// Like profile but with floats.
type fprofile [len(kmr.Profile{})][4]float64

// Converts a profile to floats.
func toFloats(p *kmr.ProfileTuple) fprofile {
	var result fprofile
	for i := range p.Data.P {
		for j := range p.Data.P[i] {
			result[i][j] = float64(p.Data.P[i][j])
		}
	}
	return result
}

// Normalizes each position to 1.
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

// TODO(amit): Remove in favor of gnum.Entropy?
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

// Normalizes the numbers by entropy, for visualization.
func normalizeByEntropy(p *fprofile) {
	for i := range p {
		factor := 2 - entropy(p[i])
		for j := range p[i] {
			p[i][j] *= factor
		}
	}
}
