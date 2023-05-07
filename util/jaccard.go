package util

import "fmt"

const jaccardStrict = true

func jaccardCommon(a, b []int) int {
	i, j := 0, 0
	common := 0
	for i < len(a) && j < len(b) {
		if jaccardStrict {
			if i > 0 && a[i] <= a[i-1] {
				panic(fmt.Sprintf("a[%d] <= a[%d]: %d <= %d",
					i, i-1, a[i], a[i-1]))
			}
			if j > 0 && b[j] <= b[j-1] {
				panic(fmt.Sprintf("b[%d] <= b[%d]: %d <= %d",
					j, j-1, b[j], b[j-1]))
			}
		}
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			common++
			i++
			j++
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
