// Reader + writer types and file opening functionality.

// Package aio provides convenience functions for buffered reading and writing
// to files.
package aio

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// Reader provides the required functions for reading and decoding data.
type Reader interface {
	io.ReadCloser
	io.ByteReader
}

type breader struct {
	*bufio.Reader
	f *os.File
}

func (r *breader) Close() error {
	return r.f.Close()
}

func newBReader(file string) (Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	b := bufio.NewReader(f)
	return &breader{b, f}, nil
}

func newBReaderGzip(file string) (Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	b := bufio.NewReader(z)
	return &breader{b, f}, nil
}

func newBReaderZStd(file string) (Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	z, err := zstd.NewReader(f, zstd.WithDecoderConcurrency(1))
	if err != nil {
		f.Close()
		return nil, err
	}
	b := bufio.NewReader(z)
	return &breader{b, f}, nil
}

// Open opens a file for reading. Gzip files are automatically decompressed.
func Open(file string) (Reader, error) {
	if strings.HasSuffix(file, ".gz") {
		return newBReaderGzip(file)
	}
	if strings.HasSuffix(file, ".zst") {
		return newBReaderZStd(file)
	}
	return newBReader(file)
}

// Writer provides the required functions for writing and encoding data.
type Writer interface {
	io.WriteCloser
	io.ByteWriter
	io.StringWriter
}

type bwriter struct {
	*bufio.Writer
	f *os.File
}

func (w *bwriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	return nil
}

func newBWriter(file string, append bool) (Writer, error) {
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	} else {
		f, err = os.Create(file)
	}
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriter(f)
	return &bwriter{b, f}, nil
}

type gzwriter struct {
	*bufio.Writer
	f *os.File
	z *gzip.Writer
}

func (w *gzwriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.z.Close(); err != nil {
		return err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	return nil
}

func newGZWriter(file string, append bool) (Writer, error) {
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	} else {
		f, err = os.Create(file)
	}
	if err != nil {
		return nil, err
	}
	z, err := gzip.NewWriterLevel(f, 1)
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriter(z)
	return &gzwriter{b, f, z}, nil
}

type zstdwriter struct {
	*bufio.Writer
	f *os.File
	z *zstd.Encoder
}

func (w *zstdwriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.z.Close(); err != nil {
		return err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	return nil
}

func newZStdWriter(file string, append bool) (Writer, error) {
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	} else {
		f, err = os.Create(file)
	}
	if err != nil {
		return nil, err
	}
	z, err := zstd.NewWriter(f, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	b := bufio.NewWriter(z)
	return &zstdwriter{b, f, z}, nil
}

// Create opens a file for writing. .gz and .zst files are automatically
// compressed.
func Create(file string) (Writer, error) {
	if strings.HasSuffix(file, ".gz") {
		return newGZWriter(file, false)
	}
	if strings.HasSuffix(file, ".zst") {
		return newZStdWriter(file, false)
	}
	return newBWriter(file, false)
}

// Append opens a file for appending. .gz and .zst files are automatically
// compressed.
func Append(file string) (Writer, error) {
	if strings.HasSuffix(file, ".gz") {
		return newGZWriter(file, true)
	}
	if strings.HasSuffix(file, ".zst") {
		return newZStdWriter(file, true)
	}
	return newBWriter(file, true)
}

type rwrapper struct {
	*bufio.Reader
	r io.ReadCloser
}

func (r *rwrapper) Close() error {
	return r.r.Close()
}

func WrapReader(r io.ReadCloser) Reader {
	b := bufio.NewReader(r)
	return &rwrapper{b, r}
}

type wwrapper struct {
	*bufio.Writer
	w io.WriteCloser
}

func (w *wwrapper) Close() error {
	if err := w.Flush(); err != nil {
		w.w.Close()
		return err
	}
	return w.w.Close()
}

func WrapWriter(w io.WriteCloser) Writer {
	b := bufio.NewWriter(w)
	return &wwrapper{b, w}
}

// Buffer is a regular buffer that implements the Reader and Writer interfaces.
// Closing a buffer is a no-op.
type Buffer struct {
	bytes.Buffer
}

// NewBuffer creates a new buffer with b as its initial content.
func NewBuffer(b []byte) *Buffer {
	return &Buffer{*bytes.NewBuffer(b)}
}

// NewBufferString creates a new buffer with s as its initial content.
func NewBufferString(s string) *Buffer {
	return &Buffer{*bytes.NewBufferString(s)}
}

func (b *Buffer) Close() error {
	return nil
}

// Discard is a Writer that does nothing and always succeeds.
type Discard struct{}

func (d Discard) Write(b []byte) (int, error) {
	return len(b), nil
}

func (d Discard) WriteByte(byte) error {
	return nil
}

func (d Discard) WriteString(s string) (int, error) {
	return len(s), nil
}

func (d Discard) Close() error {
	return nil
}
