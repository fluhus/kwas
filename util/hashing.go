package util

import (
	"github.com/spaolacci/murmur3"
	"golang.org/x/exp/constraints"
)

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

// For encoding integers.
var hash64IntBuf = make([]byte, 8)

// Hash64Int returns the hash of the given int.
func Hash64Int[T constraints.Integer](x T) uint64 {
	u := uint64(x)
	for i := range 8 {
		hash64IntBuf[i] = byte(u >> i)
	}
	return Hash64(hash64IntBuf)
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
