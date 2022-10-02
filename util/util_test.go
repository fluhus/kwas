package util

import (
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestChooseStrings(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	var got []string
	for i := 0; i < 4; i++ {
		s, _ := ChooseStrings(in, i, 4)
		if len(s) > 3 {
			t.Fatalf("ChooseStrings(\"...\", %v, 4)=%v, want <= 3", i, s)
		}
		got = append(got, s...)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("ChooseStrings(%v)=%v, want %v", in, got, in)
	}
}

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

func TestHashString(t *testing.T) {
	inputs := []string{"", "a", "blablabla", "blublublu"}
	for _, input := range inputs {
		got := Hash64String(input)
		want := Hash64([]byte(input))
		if got != want {
			t.Errorf("Hash64String(%q)=%v, want %v", input, got, want)
		}
	}
}
