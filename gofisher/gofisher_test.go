package gofisher

import (
	"math"
	"testing"

	"github.com/fluhus/gostuff/gnum"
)

func TestGreater(t *testing.T) {
	t.Cleanup(Clear)
	tests := []struct {
		input   [4]int
		wantODR float64
		wantP   float64
	}{
		{[4]int{1, 1, 1, 1}, 1, 0.833333333333333},
		{[4]int{1, 0, 0, 1}, math.Inf(1), 0.5},
		{[4]int{0, 1, 1, 0}, 0, 1},
		{[4]int{1, 9, 11, 3}, 0.0303030303030303, 0.9999663480953115},
		{[4]int{0, 10, 12, 2}, 0, 1},
		{[4]int{60, 10, 30, 25}, 5.0, 0.0001220652942110203},
	}
	for _, test := range tests {
		odr, p := Greater(test.input[0], test.input[1], test.input[2], test.input[3])
		if !equalFloat64(odr, test.wantODR) || !equalFloat64(p, test.wantP) {
			t.Errorf("Greater(%v)=%f,%f want %f,%f",
				test.input, odr, p, test.wantODR, test.wantP)
		}
	}
}

func TestLess(t *testing.T) {
	t.Cleanup(Clear)
	tests := []struct {
		input   [4]int
		wantODR float64
		wantP   float64
	}{
		{[4]int{1, 1, 1, 1}, 1, 0.8333333333333335},
		{[4]int{1, 0, 0, 1}, math.Inf(1), 1},
		{[4]int{0, 1, 1, 0}, 0, 0.5},
		{[4]int{1, 9, 11, 3}, 0.0303030303030303, 0.001379728092610052},
		{[4]int{0, 10, 12, 2}, 0, 3.365190469780604e-05},
		{[4]int{60, 10, 30, 25}, 5, 0.9999777586419631},
	}
	for _, test := range tests {
		odr, p := Less(test.input[0], test.input[1], test.input[2], test.input[3])
		if !equalFloat64(odr, test.wantODR) || !equalFloat64(p, test.wantP) {
			t.Errorf("Less(%v)=%f,%f want %f,%f",
				test.input, odr, p, test.wantODR, test.wantP)
		}
	}
}

func TestTwoSided(t *testing.T) {
	t.Cleanup(Clear)
	tests := []struct {
		input   [4]int
		wantODR float64
		wantP   float64
	}{
		{[4]int{1, 1, 1, 1}, 1, 1},
		{[4]int{1, 0, 0, 1}, math.Inf(1), 1},
		{[4]int{0, 1, 1, 0}, 0, 1},
		{[4]int{1, 9, 11, 3}, 0.0303030303030303, 0.002759456185220104},
		{[4]int{0, 10, 12, 2}, 0.0, 6.730380939561209e-05},
		{[4]int{60, 10, 30, 25}, 5, 0.00023568647189489372},
	}
	for _, test := range tests {
		odr, p := TwoSided(
			test.input[0], test.input[1], test.input[2], test.input[3])
		if !equalFloat64(odr, test.wantODR) || !equalFloat64(p, test.wantP) {
			t.Errorf("TwoSided(%v)=%f,%f want %f,%f",
				test.input, odr, p, test.wantODR, test.wantP)
		}
	}
}

func TestBadInput(t *testing.T) {
	tests := [][4]int{
		{-1, 1, 1, 1},
		{1, -1, 1, 1},
		{1, 1, -1, 1},
		{1, 1, 1, -1},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Cleanup(Clear)
			defer func() { recover() }()
			Greater(test[0], test[1], test[2], test[3])
			t.Errorf("Greater(%v) succeeded, want fail", test)
		})
		t.Run("", func(t *testing.T) {
			t.Cleanup(Clear)
			defer func() { recover() }()
			Less(test[0], test[1], test[2], test[3])
			t.Errorf("Less(%v) succeeded, want fail", test)
		})
		t.Run("", func(t *testing.T) {
			t.Cleanup(Clear)
			defer func() { recover() }()
			TwoSided(test[0], test[1], test[2], test[3])
			t.Errorf("TwoSided(%v) succeeded, want fail", test)
		})
	}
}

func equalFloat64(a, b float64) bool {
	const delta = 0.00000000001
	if math.IsInf(a, 1) || math.IsInf(b, 1) {
		return math.IsInf(a, 1) && math.IsInf(b, 1)
	}
	if math.IsInf(a, -1) || math.IsInf(b, -1) {
		return math.IsInf(a, -1) && math.IsInf(b, -1)
	}
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.IsNaN(a) && math.IsNaN(b)
	}
	return gnum.Diff(a, b) < delta
}
