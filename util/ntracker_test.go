package util

import (
	"slices"
	"testing"
)

func TestNonNSubseqs(t *testing.T) {
	input := "ATTANGGAtagnnagtt"
	want := []string{"ATT", "TTA", "GGA", "GAt", "Ata", "tag", "agt", "gtt"}
	var got []string
	for s := range NonNSubseqs([]byte(input), 3) {
		got = append(got, string(s))
	}
	if !slices.Equal(got, want) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, got, want)
	}
	got = nil
	for s := range NonNSubseqsString(input, 3) {
		got = append(got, s)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("NonNSubseqs(%q,3)=%v, want %v", input, got, want)
	}
}
