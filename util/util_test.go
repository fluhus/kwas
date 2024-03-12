package util

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	input := "a\nc\nd\nc\n"
	want := []string{"a", "c", "d", "c"}
	got, err := ReadLines(io.NopCloser(strings.NewReader(input)), nil)
	if err != nil {
		t.Fatalf("ReadLines(%q) failed: %v", input, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadLines(%q)=%v, want %v", input, got, want)
	}

	got, err = ReadLines(io.NopCloser(strings.NewReader(input)), io.EOF)
	if err != io.EOF {
		t.Fatalf("ReadLines(%q)=(%v,%v), want EOF", input, got, err)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		input string
		sep   rune
		want  []string
	}{
		{"", ',', []string{""}},
		{",", ',', []string{"", ""}},
		{"a,b,c", ',', []string{"a", "b", "c"}},
		{"a,b,c,", ',', []string{"a", "b", "c", ""}},
		{",a,b,c", ',', []string{"", "a", "b", "c"}},
		{"a,b,,c", ',', []string{"a", "b", "", "c"}},
		{"hello", ' ', []string{"hello"}},
		{"hello world yeah", ' ', []string{"hello", "world", "yeah"}},
	}
	var slice []string
	for _, test := range tests {
		got := Split(test.input, test.sep, slice)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Split(%q,%q)=%v, want %v",
				test.input, test.sep, got, test.want)
		}
	}
}

func TestCanonicalKmers(t *testing.T) {
	input := []byte("ATTAGGCAC")
	want := []string{"AAT", "TAA", "CTA", "AGG", "GCC", "GCA", "CAC"}
	var got []string
	CanonicalKmers(input, 3, func(b []byte) {
		got = append(got, string(b))
	})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CanonicalKmers(%q,3)=%v, want %v", input, got, want)
	}
}
