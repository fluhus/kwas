// Package util provides common utilities.
package util

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/gzipf"
	"github.com/spaolacci/murmur3"
	"golang.org/x/exp/constraints"
)

// Die prints the error and exits, if the error is non-nil.
func Die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(2)
	}
}

// OpenOrStdin returns an open reader from the given file, or stdin of f equals
// the stdin value.
func OpenOrStdin(f string, stdin string) (io.ReadCloser, error) {
	if f == stdin {
		return os.Stdin, nil
	}
	return gzipf.Open(f)
}

var hash64 = murmur3.New64()

// Hash64 returns a 64-bit stable hash for the given bytes.
func Hash64(b ...[]byte) uint64 {
	hash64.Reset()
	for _, bb := range b {
		hash64.Write(bb)
	}
	return hash64.Sum64()
}

// Hash64String returns a 64-bit stable hash for the given string.
func Hash64String(s ...string) uint64 {
	hash64.Reset()
	for _, ss := range s {
		writeStringToHash(ss)
	}
	return hash64.Sum64()
}

// Buffer for copying strings and writing them to hash.
var hashStringBuf = make([]byte, 1<<10)

// Writes the given string to the hash.
func writeStringToHash(s string) {
	for i := 0; i < len(s); {
		n := copy(hashStringBuf, s[i:])
		hash64.Write(hashStringBuf[:n])
		i += n
	}
}

// ChooseStrings returns a sublist of the strings. p is 0-based.
func ChooseStrings(s []string, p, np int) ([]string, []int) {
	var result []string
	var idx []int
	for i, ss := range s {
		if Hash64String(ss)%uint64(np) == uint64(p) {
			result = append(result, ss)
			idx = append(idx, i)
		}
	}
	return result, idx
}

func SaveGob(file string, v interface{}) error {
	f, err := gzipf.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(v)
}

func LoadGob(file string, v interface{}) error {
	f, err := gzipf.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(v)
}

func ReadLines(r io.ReadCloser, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var result []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		result = append(result, sc.Text())
	}
	if sc.Err() != nil {
		return nil, sc.Err()
	}
	return result, nil
}

// Split splits s by sep. Slice can be recycled for performance.
func Split(s string, sep rune, slice []string) []string {
	slice = slice[:0]
	last := 0
	updateLast := false
	for i, c := range s {
		if updateLast {
			last = i
			updateLast = false
		}
		if c == sep {
			slice = append(slice, s[last:i])
			updateLast = true
		}
	}
	if updateLast {
		last = len(s)
	}
	slice = append(slice, s[last:])
	return slice
}

// SplitBytes splits s by sep and reuses slice for storing the result.
// Returns the parts.
func SplitBytes(s []byte, sep byte, slice [][]byte) [][]byte {
	slice = slice[:0]
	last := 0
	updateLast := false
	for i, c := range s {
		if updateLast {
			last = i
			updateLast = false
		}
		if c == sep {
			slice = append(slice, s[last:i])
			updateLast = true
		}
	}
	if updateLast {
		last = len(s)
	}
	slice = append(slice, s[last:])
	return slice
}

// CanonicalKmers iterates over canonical k-long subsequences of seq.
// Makes one call to ReverseComplement.
func CanonicalKmers(seq []byte, k int, foreach func([]byte)) {
	rc := sequtil.ReverseComplement(make([]byte, 0, len(seq)), seq)
	nk := len(seq) - k + 1

	lastN := -1
	for i, a := range seq[:k] {
		if isN(a) {
			lastN = i
		}
	}

	for i := 0; i < nk; i++ {
		kmer := seq[i : i+k]
		if isN(kmer[k-1]) {
			lastN = i + k - 1
		}
		if lastN >= i {
			continue
		}
		kmerRC := rc[len(rc)-i-k : len(rc)-i]
		if bytes.Compare(kmer, kmerRC) == 1 {
			kmer = kmerRC
		}
		foreach(kmer)
	}
}

// Perc returns the a/b in %.
func Perc(a, b int) float64 {
	return 100 * float64(a) / float64(b)
}

// Percf returns a/b in the format "x%" with the given precision.
func Percf(a, b, p int) string {
	return fmt.Sprintf(fmt.Sprintf("%%.%df%%%%", p), Perc(a, b))
}

func ArgMin[S ~[]E, E constraints.Ordered](s S) int {
	if len(s) == 0 {
		return -1
	}
	imin, min := 0, s[0]
	for i, v := range s {
		if v < min {
			imin, min = i, v
		}
	}
	return imin
}

func ArgMax[S ~[]E, E constraints.Ordered](s S) int {
	if len(s) == 0 {
		return -1
	}
	imax, max := 0, s[0]
	for i, v := range s {
		if v > max {
			imax, max = i, v
		}
	}
	return imax
}

func NTiles[S ~[]E, E constraints.Ordered](n int, s S) S {
	result := make(S, n+1)
	for i := 0; i <= n; i++ {
		j := int(math.Round(float64(i) / float64(n) * float64(len(s)-1)))
		result[i] = s[j]
	}
	return result
}

// Checks if a byte equals N or n.
func isN(b byte) bool {
	return b == 'N' || b == 'n'
}
