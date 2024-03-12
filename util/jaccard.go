package util

import (
	"fmt"
	"slices"
)

const jaccardStrict = true

func jaccardCommon(a, b []int) int {
	if jaccardStrict {
		if !slices.IsSorted(a) {
			panic(fmt.Sprintf("a is not sorted: %v", a))
		}
		if !slices.IsSorted(b) {
			panic(fmt.Sprintf("b is not sorted: %v", b))
		}
	}

	j := 0
	common := 0
loop:
	for _, aa := range a {
		for _, bb := range b[j:] {
			switch {
			case aa < bb:
				continue loop
			case aa > bb:
				j++
			default:
				j++
				common++
				continue loop
			}
		}
	}
	return common
}

// JaccardDist returns the Jaccard distance (1-similarity) between
// two sorted lists.
func JaccardDist(a, b []int) float64 {
	if len(a) == 0 && len(b) == 0 { // Avoid 0/0.
		return 0
	}
	common := jaccardCommon(a, b)
	union := len(a) + len(b) - common
	return 1 - float64(common)/float64(union)
}

// JaccardDualDist returns the average Jaccard distance (1-similarity) between
// the two sorted lists and their complements.
// n is the number of possible elements (0 to n-1).
func JaccardDualDist(a, b []int, n int) float64 {
	common := jaccardCommon(a, b)
	union := len(a) + len(b) - common
	common2 := n - union
	union2 := n - common
	if union == 0 || union2 == 0 { // Avoid 0/0.
		return 0
	}
	return 1 - float64(common)/float64(union)/2 - float64(common2)/float64(union2)/2
}
