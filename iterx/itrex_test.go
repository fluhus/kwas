package iterx

import (
	"iter"
	"slices"
	"testing"
)

func TestUntil(t *testing.T) {
	input := []int{2, 1, 3, 5, 8, 13, 66, 10, 7}
	conds := []int{3, 4, 5, 6}
	want := [][]int{{2, 1}, {3, 5}, {8, 13, 66}, {10, 7}}
	var got [][]int
	it := New(iterSlice(input))
	for _, c := range conds {
		var gott []int
		for x, _ := range it.Until(func(i int) bool { return i%c == 0 }) {
			gott = append(gott, x)
		}
		got = append(got, gott)
	}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("Until(...)=%v, want %v", got, want)
	}
}

func iterSlice[T any](s []T) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, x := range s {
			if !yield(x, nil) {
				return
			}
		}
	}
}
