package kmr

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fluhus/gostuff/aio"
)

// IterTuplesFile iterates the given file, calling forEach on each tuple.
// Reuses init as input to forEach, so keeping instances should use Copy().
func IterTuplesFile[T Tuple](file string, init T, forEach func(T) error) error {
	f, err := aio.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	for err = init.Decode(f); err == nil; err = init.Decode(f) {
		if err := forEach(init); err != nil {
			return err
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

// IterTuplesFiles iterates the files matching the given glob pattern,
// calling forEach on each tuple.
// Reuses init as input to forEach, so keeping instances should use Copy().
func IterTuplesFiles[T Tuple](glob string, init T, forEach func(T) error) error {
	files, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("found 0 files")
	}
	for _, file := range files {
		if err := IterTuplesFile(file, init, forEach); err != nil {
			return err
		}
	}
	return nil
}
