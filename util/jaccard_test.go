package util

import (
	"testing"

	"github.com/fluhus/gostuff/gnum"
)

func TestJaccard(t *testing.T) {
	tests := []struct {
		a    []int
		b    []int
		want float64
	}{
		{[]int{1}, nil, 1},
		{nil, []int{1}, 1},
		{[]int{1}, []int{1}, 0},
		{[]int{1, 2}, []int{1, 3}, 2.0 / 3.0},
		{[]int{1, 2, 3}, []int{1, 3, 4}, 0.5},
	}
	for _, test := range tests {
		if got := JaccardDist(test.a, test.b); gnum.Abs(got-test.want) > 0.00001 {
			t.Errorf("JaccardDist(%v,%v)=%v, want %v",
				test.a, test.b, got, test.want)
		}
	}
}

func TestJaccardDual(t *testing.T) {
	tests := []struct {
		a     []int
		b     []int
		n     int
		acomp []int
		bcomp []int
	}{
		{[]int{1}, nil, 1, nil, []int{1}},
		{nil, []int{1}, 1, []int{1}, nil},
		{[]int{1}, []int{1}, 1, nil, nil},
		{[]int{1, 2}, []int{1, 3}, 3, []int{3}, []int{2}},
		{[]int{1, 2, 3}, []int{1, 3, 4}, 4, []int{4}, []int{2}},
	}
	for _, test := range tests {
		want := (JaccardDist(test.a, test.b) + JaccardDist(test.acomp, test.bcomp)) / 2
		if got := JaccardDualDist(test.a, test.b, test.n); gnum.Abs(got-want) > 0.00001 {
			t.Errorf("JaccardDualDist(%v,%v)=%v, want %v",
				test.a, test.b, got, want)
		}
	}
}
