package util

import (
	"slices"
	"testing"
)

func TestNonNSubseqs(t *testing.T) {
	input := "ATTANGGAtagnnagtt"
	want := []string{"ATT", "TTA", "GGA", "GAt", "Ata", "tag", "agt", "gtt"}
	iwant := []int{0, 1, 5, 6, 7, 8, 13, 14}
	var got []string
	var igot []int
	for i, s := range NonNSubseqs([]byte(input), 3) {
		got = append(got, string(s))
		igot = append(igot, i)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, got, want)
	}
	if !slices.Equal(igot, iwant) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, igot, iwant)
	}
	got = nil
	igot = nil
	for i, s := range NonNSubseqsString(input, 3) {
		got = append(got, s)
		igot = append(igot, i)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, got, want)
	}
	if !slices.Equal(igot, iwant) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, igot, iwant)
	}
}
