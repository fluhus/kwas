// Package lazy provides a lazy writer.
package lazy

import (
	"os"

	"github.com/fluhus/gostuff/aio"
)

// Writer writes to a buffer which gets flushed into a file when full.
type Writer struct {
	f string
	n int
	b []byte
}

// NewWriter returns a new writer to the given file with a buffer of size n.
func NewWriter(file string, n int) *Writer {
	os.Remove(file)
	return &Writer{file, n, make([]byte, 0, n)}
}

// Write adds b to the buffer and flushes if needed.
func (w *Writer) Write(b []byte) (int, error) {
	old := w.b[:0] // Keep a reference to the original buffer so we can reuse it.
	w.b = append(w.b, b...)
	if len(w.b) <= w.n {
		return len(b), nil
	}
	toWrite := w.b
	w.b = old
	f, err := aio.Append(w.f)
	if err != nil {
		return 0, err
	}
	if n, err := f.Write(toWrite); err != nil {
		f.Close()
		return n, err
	}
	return len(b), f.Close()
}

// Flush empties the buffer into the file.
func (w *Writer) Flush() error {
	if len(w.b) == 0 {
		return nil
	}
	toWrite := w.b
	w.b = w.b[:0]
	f, err := aio.Append(w.f)
	if err != nil {
		return err
	}
	if _, err := f.Write(toWrite); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Close is equivalent to Flush. Implements the [io.WriteCloser] interface.
func (w *Writer) Close() error {
	return w.Flush()
}
