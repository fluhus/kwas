package util

import (
	"cmp"
	"slices"
	"sort"
	"testing"

	"golang.org/x/exp/maps"
)

func TestHashString(t *testing.T) {
	inputs := []string{"", "a", "blablabla", "blublublu"}
	for _, input := range inputs {
		got := Hash64String(input)
		want := Hash64([]byte(input))
		if got != want {
			t.Errorf("Hash64String(%q)=%v, want %v", input, got, want)
		}
	}
}

func TestChooseStrings(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	var got []string
	for i := 0; i < 4; i++ {
		s, _ := ChooseStrings(in, i, 4)
		if len(s) > 3 {
			t.Fatalf("ChooseStrings(\"...\", %v, 4)=%v, want <= 3", i, s)
		}
		got = append(got, s...)
	}
	sort.Strings(got)
	if !slices.Equal(got, in) {
		t.Fatalf("ChooseStrings(%v)=%v, want %v", in, got, in)
	}
}

func TestHash64Int(t *testing.T) {
	input := []int{4, 2, 3, 1, 4, 2, 3, 3, 1, 1, 1, 2, 3, 4}
	want := [][]int{
		{0, 4, 13}, {1, 5, 11}, {2, 6, 7, 12}, {3, 8, 9, 10},
	}
	m := map[uint64][]int{}
	for i, x := range input {
		h := Hash64Int(x)
		m[h] = append(m[h], i)
	}
	got := maps.Values(m)
	slices.SortFunc[[][]int](got, func(a, b []int) int {
		return cmp.Compare(a[0], b[0])
	})
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("Hash64Int(%d) => %v, want %v", input, got, want)
	}
}
