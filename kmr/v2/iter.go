package kmr

import (
	"fmt"
	"io"
	"iter"
	"path/filepath"

	"github.com/fluhus/gostuff/aio"
)

// IterTuplesFile iterates the given file.
func IterTuplesFile[H KmerDataHandler[T], T any](
	file string) iter.Seq2[*Tuple[H, T], error] {
	return func(yield func(*Tuple[H, T], error) bool) {
		t := NewTuple[H]()
		f, err := aio.Open(file)
		if err != nil {
			yield(nil, err)
			return
		}
		defer f.Close()
		for err = t.Decode(f); err == nil; err = t.Decode(f) {
			if !yield(t, nil) {
				break
			}
		}
		if err != io.EOF {
			yield(t, err)
		}
	}
}

// IterTuplesReader iterates the given reader.
func IterTuplesReader[H KmerDataHandler[T], T any](
	r io.ByteReader) iter.Seq2[*Tuple[H, T], error] {
	return func(yield func(*Tuple[H, T], error) bool) {
		t := NewTuple[H]()
		var err error
		for err = t.Decode(r); err == nil; err = t.Decode(r) {
			if !yield(t, nil) {
				break
			}
		}
		if err != io.EOF {
			yield(t, err)
		}
	}
}

// IterTuplesFiles iterates the files matching the given glob pattern.
func IterTuplesFiles[H KmerDataHandler[T], T any](
	glob string) iter.Seq2[*Tuple[H, T], error] {
	return func(yield func(*Tuple[H, T], error) bool) {
		files, err := filepath.Glob(glob)
		if err != nil {
			yield(nil, err)
			return
		}
		if len(files) == 0 {
			yield(nil, fmt.Errorf("found 0 files"))
			return
		}
		for _, file := range files {
			for t, err := range IterTuplesFile[H](file) {
				if err != nil {
					yield(nil, err)
					return
				}
				if !yield(t, err) {
					return
				}
			}
		}
	}
}
