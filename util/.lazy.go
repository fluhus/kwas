// Implementation of lazy reader.

package util

import (
	"io"
	"os"
	"path/filepath"
)

// LazyReader is a reader that opens its input stream only on the first call to
// Read. It automatically closes the stream upon EOF.
type LazyReader struct {
	rc   io.ReadCloser
	open func() (io.ReadCloser, error)
}

func (r *LazyReader) Read(b []byte) (int, error) {
	if r.rc == nil {
		rc, err := r.open()
		if err != nil {
			return 0, err
		}
		r.rc = rc
	}

	n, err := r.rc.Read(b)
	if err == io.EOF {
		if err := r.rc.Close(); err != nil {
			return n, err
		}
	}
	return n, err
}

// NewLazyReader creates a new lazy reader with the given function.
// The function will be called once upon the first call to Read.
func NewLazyReader(open func() (io.ReadCloser, error)) *LazyReader {
	return &LazyReader{nil, open}
}

// Openp returns a reader that concatenates all the files that match the given
// pattern. Uses LazyReader so that each file is opened only when needed and
// closed upon EOF.
func Openp(pattern string) (io.Reader, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return Opens(files)
}

// Opens returns a reader that concatenates all the files in the given slice.
// Uses LazyReader so that each file is opened only when needed and closed upon
// EOF.
func Opens(files []string) (io.Reader, error) {
	var rs []io.Reader
	for i := range files {
		f := files[i] // Closure for the anonymous function.
		rs = append(rs, NewLazyReader(func() (io.ReadCloser, error) {
			return os.Open(f)
		}))
	}
	return io.MultiReader(rs...), nil
}
