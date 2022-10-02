package kmr

import (
	"reflect"
	"testing"

	"github.com/fluhus/kwas/aio"
)

func TestPrabCountAdd(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 321}
	a.Add(b)
	want := uint64(444)
	if a.Count != want {
		t.Fatalf("Add(...)=%v, want %v", a.Count, want)
	}
}

func TestPrabCountAdd_bad(t *testing.T) {
	defer func() { recover() }()
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3}, 321}
	a.Add(b)
	t.Fatalf("Add(...) succeeded, want fail")
}

func TestPrabCountCopy(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	c := b.Copy().(*HasCount)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Copy() changed receiver %v, want %v", b, a)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("Copy()=%v, want %v", c, a)
	}
}

func TestPrabCountEncode(t *testing.T) {
	a := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	b := &HasCount{FullKmer{1, 2, 3, 4}, 123}
	c := &HasCount{}
	buf := aio.NewBuffer(nil)
	if err := b.Encode(buf); err != nil {
		t.Fatalf("%v.Encode() failed: %v", b, err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Encode() changed receiver %v, want %v", b, a)
	}
	if err := c.Decode(buf); err != nil {
		t.Fatalf("%v.Decode() failed: %v", c, err)
	}
	if !reflect.DeepEqual(a, c) {
		t.Fatalf("Decode()=%v, want %v", c, a)
	}
}
