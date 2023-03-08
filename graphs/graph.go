package graphs

import (
	"fmt"
	"sort"

	"github.com/fluhus/gostuff/sets"
	"golang.org/x/exp/maps"
)

type Graph struct {
	v int
	e sets.Set[[2]int]
}

func New(n int) *Graph {
	if n < 0 {
		panic(fmt.Sprintf("bad n: %d", n))
	}
	return &Graph{n, sets.Set[[2]int]{}}
}

func (g *Graph) NumVertices() int {
	return g.v
}

func (g *Graph) NumEdges() int {
	return len(g.e)
}

func (g *Graph) Edges() [][2]int {
	return maps.Keys(g.e)
}

func (g *Graph) HasEdge(v1, v2 int) bool {
	v1, v2 = order(v1, v2)
	return g.e.Has([2]int{v1, v2})
}

func (g *Graph) AddEdge(v1, v2 int) {
	v1, v2 = order(v1, v2)
	g.e.Add([2]int{v1, v2})
}

func (g *Graph) DeleteEdge(v1, v2 int) {
	v1, v2 = order(v1, v2)
	delete(g.e, [2]int{v1, v2})
}

func (g *Graph) ConnectedComponents() [][]int {
	edges := map[int][]int{}
	for e := range g.e {
		edges[e[0]] = append(edges[e[0]], e[1])
		edges[e[1]] = append(edges[e[1]], e[0])
	}

	m := map[int]int{}
	for i := 0; i < g.v; i++ {
		if _, ok := m[i]; ok {
			continue
		}
		m[i] = i
		queue := []int{i}
		for len(queue) > 0 {
			for _, j := range edges[queue[0]] {
				if _, ok := m[j]; !ok {
					m[j] = i
					queue = append(queue, j)
				}
			}
			queue = queue[1:]
		}
	}

	comps := map[int][]int{}
	for k, v := range m {
		comps[v] = append(comps[v], k)
	}
	var poncs [][]int
	for _, v := range comps {
		sort.Ints(v)
		poncs = append(poncs, v)
	}
	sort.Slice(poncs, func(i, j int) bool {
		return poncs[i][0] < poncs[j][0]
	})
	return poncs
}

func order(a, b int) (int, int) {
	if a > b {
		return b, a
	}
	return a, b
}
