// Reservior sasmpling implementation.

package util

import (
	"fmt"
	"math/rand"
)

type Reservoir[T any] struct {
	Sample []T
	n      int
}

func NewReservoir[T any](n int) *Reservoir[T] {
	if n < 1 {
		panic(fmt.Sprintf("bad n: %d", n))
	}
	return &Reservoir[T]{make([]T, 0, n), 0}
}

func (r *Reservoir[T]) Add(t T) {
	r.n++
	if len(r.Sample) < cap(r.Sample) {
		r.Sample = append(r.Sample, t)
		return
	}
	if i := rand.Intn(r.n); i < len(r.Sample) {
		r.Sample[i] = t
	}
}
