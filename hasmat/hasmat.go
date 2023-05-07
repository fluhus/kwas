// Creates a HAS matrix for the population structure script.
package main

/*
#include <stdint.h>

typedef void allocfunc(int64_t vals, int64_t kmers, int64_t k);
static void call_alloc(void* p, int64_t vals, int64_t kmers, int64_t k) {
	((allocfunc*)p)(vals, kmers,k);
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/gnum"
	"github.com/fluhus/gostuff/snm"
	"github.com/fluhus/kwas/kmr"
)

const verbose = false // Enable debug prints.

// Loads a matrix from a HAS file.
func loadMatrix(file string) ([][]byte, []kmr.Kmer, error) {
	if verbose {
		fmt.Println("Loading from:", file)
	}
	var vals [][]byte
	var kmers []kmr.Kmer
	err := kmr.IterTuplesFile(file, &kmr.HasTuple{},
		func(ht *kmr.HasTuple) error {
			kmers = append(kmers, ht.Kmer)
			if len(ht.Samples) == 0 {
				vals = append(vals, nil)
				return nil
			}
			v := make([]byte, ht.Samples[len(ht.Samples)-1]+1)
			for _, s := range ht.Samples {
				v[s] = 1
			}
			vals = append(vals, v)
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	// Make all same length.
	maxLen := gnum.Max(snm.Slice(len(vals), func(i int) int {
		return len(vals[i])
	}))
	for i := range vals {
		v := make([]byte, maxLen)
		copy(v, vals[i])
		vals[i] = v
	}

	if verbose {
		fmt.Println("Data shape:", len(vals), len(vals[0]))
	}

	return vals, kmers, nil
}

// The API function for python.
//
//export cLoadMatrix
func cLoadMatrix(file *C.char, alloc unsafe.Pointer,
	pbuf **uint8, pkmers ***byte) {
	if verbose {
		fmt.Println("Called go function")
	}
	data, kmers, _ := loadMatrix(C.GoString(file))
	bufLen := len(data) * len(data[0])
	C.call_alloc(alloc, C.int64_t(bufLen), C.int64_t(len(kmers)), kmr.K)
	buf := unsafe.Slice(*pbuf, bufLen)[:0]
	for _, row := range data {
		buf = append(buf, row...)
	}
	ckmers := unsafe.Slice(*pkmers, len(kmers))
	for i := range ckmers {
		kmerBuf := unsafe.Slice(ckmers[i], kmr.K+1)
		str := sequtil.DNAFrom2Bit(nil, kmers[i][:])[:kmr.K]
		copy(kmerBuf, str)
	}
}

func main() {}
