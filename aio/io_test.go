package aio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestWriteUvarint(t *testing.T) {
	in := []uint64{23544, 231, 5454, 2762, 665756, 86756}
	buf := &closeBuffer{}
	for _, i := range in {
		if err := WriteUvarint(buf, i); err != nil {
			t.Fatalf("WriteUvarint(%v)=%q, want success", i, err.Error())
		}
	}
	for _, i := range in {
		got, err := ReadUvarint(buf)
		if err != nil {
			t.Fatalf("ReadUvarint()=%q, want success", err.Error())
		}
		if got != i {
			t.Fatalf("ReadUvarint()=%v, want %v", got, i)
		}
	}
	if _, err := ReadUvarint(buf); err != io.EOF {
		t.Fatalf("ReadUvarint()=%q, want EOF", err.Error())
	}
}

func TestWriteBytes(t *testing.T) {
	input := make([]byte, 1000)
	want := make([]byte, 1000)
	for i := range input {
		input[i] = byte(i * i)
	}
	copy(want, input)
	buf := &closeBuffer{}
	if err := WriteBytes(buf, input); err != nil {
		t.Fatalf("WriteBytes(...) failed: %v", err)
	}
	got, err := ReadBytes(buf)
	if err != nil {
		t.Fatalf("ReadBytes(...) failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadBytes(...)=%v, want %v", got, want)
	}
}

func TestWriteString(t *testing.T) {
	want := []string{"amit", "עמית", "hello", "שלום"}
	buf := NewBuffer(nil)
	for _, s := range want {
		if err := WriteString(buf, s); err != nil {
			t.Fatalf("WriteString(%q) failed: %v", s, err)
		}
	}
	for _, s := range want {
		got, err := ReadString(buf)
		if err != nil {
			t.Fatalf("ReadString(...) failed: %v", err)
		}
		if got != s {
			t.Fatalf("ReadString(...)=%s, want %s", got, s)
		}
	}
	if got, err := ReadString(buf); err != io.EOF {
		t.Fatalf("ReadString(...)=%s, want EOF", got)
	}
}

func BenchmarkWriteUvarint(b *testing.B) {
	for i := 0; i < 9; i++ {
		b.Run(fmt.Sprint(i), func(b *testing.B) {
			n := uint64(1) << (7 * i)
			w := &closeBufio{*bufio.NewWriter(io.Discard)}
			for j := 0; j < b.N; j++ {
				WriteUvarint(w, n)
			}
		})
	}
}

type closeBuffer struct {
	bytes.Buffer
}

func (c *closeBuffer) Close() error {
	return nil
}

type closeBufio struct {
	bufio.Writer
}

func (c *closeBufio) Close() error {
	return nil
}
