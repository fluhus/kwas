package graphs

import (
	"reflect"
	"testing"
)

func TestComponents(t *testing.T) {
	edges := [][2]int{
		{0, 1}, {1, 2}, {5, 7}, {6, 9}, {9, 10}, {8, 10}, {7, 8},
	}
	want := [][]int{
		{0, 1, 2}, {3}, {4}, {5, 6, 7, 8, 9, 10}, {11},
	}
	g := New(12)
	for _, e := range edges {
		g.AddEdge(e[0], e[1])
	}
	got := g.ConnectedComponents()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("components(...)=%v, want %v", got, want)
	}
}
