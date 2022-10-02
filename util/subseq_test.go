package util

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestSubseq(t *testing.T) {
	s := []int{1, 3, 5, 9, 11}
	want := [][]int{
		{1, 3, 5},
		{3, 5, 9},
		{5, 9, 11},
	}
	var got [][]int
	for ss, it := Subseqs(s, 3); ss != nil; ss = it.Next() {
		got = append(got, ss)
	}
	if !slices.EqualFunc(got, want, slices.Equal[int]) {
		t.Errorf("Subseqs(%v)=%v, want %v", s, got, want)
	}
}
