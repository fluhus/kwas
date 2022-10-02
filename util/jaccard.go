package util

import "fmt"

const jaccardStrict = true

func JaccardDist(a, b []int) float64 {
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
	return 1 - float64(common)/float64(len(a)+len(b)-common)
}
