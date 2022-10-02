// Binary encoding and decoding of values.

package aio

import (
	"fmt"
	"io"
	"strings"
)

// WriteUvarint writes a varint-encoded uint64 to the given writer.
func WriteUvarint(w Writer, x uint64) error {
	if x == 0 {
		return w.WriteByte(0)
	}
	for {
		b := byte(x & 127)
		x >>= 7
		if x > 0 {
			b |= 128
		}
		if err := w.WriteByte(b); err != nil {
			return err
		}
		if x == 0 {
			break
		}
	}
	return nil
}

// ReadUvarint reads a varint-encoded uint64 from the given reader.
func ReadUvarint(r Reader) (uint64, error) {
	var result uint64
	first := true
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			if !first {
				// EOF is unexpected if we read an indication that more bytes
				// are coming.
				return 0, NotExpectingEOF(err)
			}
			return 0, err
		}
		first = false
		result += uint64(b&127) << (7 * i)
		if b&128 == 0 {
			break
		}
	}
	return result, nil
}

// ReadBytes reads a slice of bytes from the given reader.
func ReadBytes(r Reader) ([]byte, error) {
	n, err := ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, NotExpectingEOF(err)
	}
	return b, nil
}

// WriteBytes writes a slice of bytes to the given writer.
func WriteBytes(w Writer, b []byte) error {
	if err := WriteUvarint(w, uint64(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

// ReadString reads a string from the given reader.
func ReadString(r Reader) (string, error) {
	n, err := ReadUvarint(r)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	if n >= 1000000 {
		return "", fmt.Errorf("string too long: %v", n)
	}
	var b strings.Builder
	b.Grow(int(n))
	_, err = io.CopyN(&b, r, int64(n))
	if err != nil {
		return "", fmt.Errorf("read string len=%v: %v",
			n, NotExpectingEOF(err))
	}
	return b.String(), nil
}

// WriteString writes a string to the given writer.
func WriteString(w Writer, s string) error {
	if err := WriteUvarint(w, uint64(len(s))); err != nil {
		return err
	}
	_, err := w.WriteString(s)
	return err
}

func NotExpectingEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
