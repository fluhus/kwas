// HasTuple logic.

package kmr

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/gostuff/snm"
	"github.com/fluhus/kwas/regress"
)

const (
	useLinSort     = false
	useRegressSort = false
)

// HasTuple holds a kmer and the sample IDs that have it.
type HasTuple = Tuple[HasHandler, HasData]

type HasData struct {
	Samples      []int
	SortOnEncode bool // If true, will sort before encoding.
}

type HasHandler struct{}

func (h HasHandler) encode(c HasData, w *bnry.Writer) error {
	if c.SortOnEncode {
		if useLinSort && len(c.Samples) > 2000 {
			linearSort(c.Samples)
		} else if useRegressSort {
			regressSort(c.Samples)
		} else {
			slices.Sort(c.Samples)
		}
	}
	toDiffs(c.Samples)
	err := w.Write(c.Samples)
	fromDiffs(c.Samples)
	return err
}

func (h HasHandler) decode(c *HasData, r io.ByteReader) error {
	s := c.Samples[:0]
	if err := bnry.Read(r, &s); err != nil {
		return err
	}
	fromDiffs(s)
	c.Samples = s
	return nil
}

func (h HasHandler) merge(a, b HasData) HasData {
	if a.SortOnEncode != b.SortOnEncode {
		panic(fmt.Sprintf("inputs disagree on SortOnEncode: %v, %v",
			a.SortOnEncode, b.SortOnEncode))
	}
	return HasData{append(a.Samples, b.Samples...), a.SortOnEncode}
}

func (h HasHandler) clone(c HasData) HasData {
	return HasData{slices.Clone(c.Samples), c.SortOnEncode}
}

func (h HasHandler) new() HasData {
	return HasData{SortOnEncode: true}
}

func fromDiffs(a []int) {
	if len(a) == 0 {
		return
	}
	for i := range a[1:] {
		a[i+1] = a[i] + a[i+1]
	}
}

func toDiffs(a []int) {
	if len(a) == 0 {
		return
	}
	last := a[0]
	for i := range a[1:] {
		lastt := a[i+1]
		a[i+1] = a[i+1] - last
		last = lastt
	}
}

func linearSort(a []int) {
	if len(a) == 0 {
		return
	}
	mn, mx := gnum.Min(a), gnum.Max(a)
	b := make([]bool, mx-mn+1)
	for _, s := range a {
		b[s-mn] = true
	}
	a = a[:0]
	for i := range b {
		if b[i] {
			a = append(a, i+mn)
		}
	}
}

var (
	rsLinX  [][]float64
	rsLinY  []float64
	rsLinB  []float64
	rsSortX []float64
	rsSortY []float64
	rsSortB float64

	logmap = snm.NewDefaultMap(func(i int) float64 {
		return math.Log(float64(i))
	})
)

func regressSort(a []int) {
	n := len(a)
	if n == 0 {
		return
	}
	mn, mx := gnum.Min(a), gnum.Max(a)
	// nlogn := float64(n) * logmap.Get(n) // math.Log(float64(n))
	nlogn := float64(n * logint(n))
	span := mx - mn + 1

	pLin, pSort := 1.0, 1.0
	if len(rsLinY) >= 10 && len(rsSortY) >= 10 { // Enough data to make a prediction.
		pLin = 1 / (float64(n)*rsLinB[0] + float64(span)*rsLinB[1])
		pSort = 1 / (nlogn * rsSortB)
	}

	if rand.Float64() < pSort/(pSort+pLin) {
		// t := time.Now()
		slices.Sort(a)
		// d := time.Since(t).Seconds()
		// rsSortX = append(rsSortX, nlogn)
		// rsSortY = append(rsSortY, d)
		// if len(rsSortY) == cap(rsSortY) {
		// 	rsSortB = regress.Regression1(rsSortX, rsSortY)
		// }
		return
	}

	// t := time.Now()
	b := make([]bool, span)
	for _, s := range a {
		b[s-mn] = true
	}
	a = a[:0]
	for i := range b {
		if b[i] {
			a = append(a, i+mn)
		}
	}
	// d := time.Since(t).Seconds()
	// rsLinX = append(rsLinX, []float64{float64(n), float64(span)})
	// rsLinY = append(rsLinY, d)
	// if len(rsLinY) == cap(rsLinY) {
	// 	rsLinB = regress.Regression2(rsLinX, rsLinY)
	// }
}

func regressSortTrain(a []int) {
	n := len(a)
	if n == 0 {
		return
	}
	mn, mx := gnum.Min(a), gnum.Max(a)
	nlogn := float64(n * logint(n))
	span := mx - mn + 1

	aa := slices.Clone(a)
	t := time.Now()
	slices.Sort(aa)
	d := time.Since(t).Seconds()
	rsSortX = append(rsSortX, nlogn)
	rsSortY = append(rsSortY, d)
	if len(rsSortY) == cap(rsSortY) {
		rsSortB = regress.Regression1(rsSortX, rsSortY)
	}

	a = slices.Clone(a)
	t = time.Now()
	b := make([]bool, span)
	for _, s := range a {
		b[s-mn] = true
	}
	a = a[:0]
	for i := range b {
		if b[i] {
			a = append(a, i+mn)
		}
	}
	d = time.Since(t).Seconds()
	rsLinX = append(rsLinX, []float64{float64(n), float64(span)})
	rsLinY = append(rsLinY, d)
	if len(rsLinY) == cap(rsLinY) {
		rsLinB = regress.Regression2(rsLinX, rsLinY)
	}
}

func init() {
	if useRegressSort {
		t := time.Now()
		for _, n := range []int{10, 100, 1000, 10000, 100000, 1000000} {
			a := rand.Perm(n)
			for i := 1; i <= 10; i++ {
				regressSortTrain(a[:len(a)*i/10])
			}
		}
		fmt.Println("Init took", time.Since(t))
	}
}

func logint(n int) int {
	a := 0
	for n > 0 {
		a++
		n /= 2
	}
	return a
}
