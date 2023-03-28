package util

import "fmt"

// An Unreader wraps a read function and adds an unread function.
type Unreader[T any] struct {
	read    func() (T, error) // The underlying reader
	head    T                 // The last read element
	hasHead bool              // Unread was called
}

// Read returns the result of calling read(), or the last
// unread element with a nil error.
func (r *Unreader[T]) Read() (T, error) {
	if r.hasHead {
		r.hasHead = false
		return r.head, nil
	}
	t, err := r.read()
	r.head = t
	return t, err
}

// Unread makes the next call to Read return the last element and a nil error.
// Can be called up to once per call to Read.
func (r *Unreader[T]) Unread() {
	if r.hasHead {
		panic(fmt.Sprintf("called Unread twice: first with %v", r.head))
	}
	r.hasHead = true
}

// ReadUntil calls Read until an error is returned or stop returns true.
// Returns the encountered error or nil if stop returned true.
func (r *Unreader[T]) ReadUntil(stop func(T) bool, forEach func(T) error) error {
	for {
		t, err := r.Read()
		if err != nil {
			return err
		}
		if stop(t) {
			r.Unread()
			return nil
		}
		if err := forEach(t); err != nil {
			return err
		}
	}
}

// NewUnreader returns an Unreader with read as its underlying read function.
func NewUnreader[T any](read func() (T, error)) *Unreader[T] {
	return &Unreader[T]{read: read, hasHead: false}
}
