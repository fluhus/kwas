// Package gofisher provides an implementation of Fisher's exact test.
package gofisher

import (
	"fmt"
	"math"

	"github.com/fluhus/gostuff/gnum"
	"golang.org/x/exp/slices"
)

// Log-factorial cache.
var facs = []float64{0}

// Fill the log-factorial cache up to i.
func fillFacs(i int) {
	if i > len(facs) {
		facs = slices.Grow(facs, i-len(facs))
	}
	for j := len(facs); j <= i; j++ {
		facs = append(facs, facs[j-1]+math.Log(float64(j)))
	}
}

// Returns the probability of a single contingency table.
func fisherSingleTable(a, b, c, d, n int) float64 {
	return math.Exp(facs[a+b] + facs[c+d] + facs[a+c] + facs[b+d] - facs[n] -
		facs[a] - facs[b] - facs[c] - facs[d])
}

// Panics if one of the inputs is negative.
func checkInput(a, b, c, d int) {
	if a < 0 {
		panic(fmt.Sprintf("a is negative: %d", a))
	}
	if b < 0 {
		panic(fmt.Sprintf("b is negative: %d", b))
	}
	if c < 0 {
		panic(fmt.Sprintf("c is negative: %d", c))
	}
	if d < 0 {
		panic(fmt.Sprintf("d is negative: %d", d))
	}
}

// Returns the odds-ratio.
func odr(a, b, c, d int) float64 {
	if b*c == 0 {
		if a*d == 0 {
			return math.NaN() // 0/0
		} else {
			return math.Inf(1) // X/0
		}
	} else {
		return float64(a) * float64(d) / float64(b) / float64(c)
	}
}

// Greater returns the probability that a is at least as big.
func Greater(a, b, c, d int) (float64, float64) {
	checkInput(a, b, c, d)
	n := a + b + c + d
	or := odr(a, b, c, d)
	fillFacs(n)
	p := 0.0
	for b >= 0 && c >= 0 {
		p += fisherSingleTable(a, b, c, d, n)
		a++
		b--
		c--
		d++
	}
	return or, p
}

// Less returns the probability that a is at most as big.
func Less(a, b, c, d int) (float64, float64) {
	checkInput(a, b, c, d)
	n := a + b + c + d
	or := odr(a, b, c, d)
	fillFacs(n)
	p := 0.0
	for a >= 0 && d >= 0 {
		p += fisherSingleTable(a, b, c, d, n)
		a--
		b++
		c++
		d--
	}
	return or, p
}

// TwoSided returns the probability of a's at most as likely as a.
func TwoSided(a, b, c, d int) (float64, float64) {
	checkInput(a, b, c, d)
	n := a + b + c + d
	or := odr(a, b, c, d)
	fillFacs(n)
	aOriginal := a
	pOriginal := fisherSingleTable(a, b, c, d, n)

	p := pOriginal

	// First side, make a low.
	diff := gnum.Min2(a, d)
	a -= diff
	b += diff
	c += diff
	d -= diff
	for a < aOriginal {
		pCurrent := fisherSingleTable(a, b, c, d, n)
		if pCurrent > pOriginal { // Event is more probable than input.
			break
		}
		p += pCurrent
		a++
		b--
		c--
		d++
	}

	// Second side, make a high.
	diff = gnum.Min2(b, c)
	a += diff
	b -= diff
	c -= diff
	d += diff
	for a > aOriginal {
		pCurrent := fisherSingleTable(a, b, c, d, n)
		if pCurrent > pOriginal { // Event is more probable than input.
			break
		}
		p += pCurrent
		a--
		b++
		c++
		d--
	}

	return or, p
}

// Clear clears the log-factorial cache.
func Clear() {
	facs = []float64{0}
}
