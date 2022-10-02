package util

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestNTracker(t *testing.T) {
	seq := []byte("ATTANGGAtagnnagttgnanttan")
	want := []bool{
		false, false, true, true, true, false, false, false, false,
		true, true, true, true, false, false, false,
		true, true, true, true, true, false, true,
	}
	var got []bool
	nt := NewNTracker(seq, 3)
	for range seq[2:] {
		got = append(got, nt.NextN())
	}
	if !slices.Equal(got, want) {
		t.Fatalf("NextN()=%v, want %v", got, want)
	}

	defer func() { recover() }()
	nt.NextN()
	t.Fatal("NextN() succeeded, want panic")
}
