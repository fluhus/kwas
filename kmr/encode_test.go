package kmr

import (
	"bytes"
	"io"
	"testing"

	"golang.org/x/exp/slices"
)

func TestWriteRead(t *testing.T) {
	inputs := [][]Kmer{
		{},
		{{5, 4, 3, 2, 1}},
		{{0, 0, 0, 0, 0}, {0, 0, 0, 0, 1}, {0, 0, 0, 0, 3}, {0, 0, 0, 1, 5}},
		{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 1}, {3, 4, 5, 1, 2}, {4, 5, 1, 2, 3},
			{5, 1, 2, 3, 4}},
		{{5, 1, 2, 3, 4}, {4, 5, 1, 2, 3}, {3, 4, 5, 1, 2}, {2, 3, 4, 5, 1},
			{1, 2, 3, 4, 5}},
	}
	for _, input := range inputs {
		buf := bytes.NewBuffer(nil)
		w := NewWriter(buf)
		for _, kmer := range input {
			if err := w.Write(kmer); err != nil {
				t.Fatalf("Write(%v) failed: %v", kmer, err)
			}
		}
		var got []Kmer
		r := NewReader(buf)
		for {
			kmer, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Read(%v) failed: %v", input, err)
			}
			got = append(got, kmer)
		}

		if !slices.Equal(got, input) {
			t.Fatalf("Write+Read(%v)=%v", input, got)
		}
	}
}
