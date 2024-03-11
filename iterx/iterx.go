package iterx

import (
	"fmt"
	"iter"
)

// An Iter wraps an iterator and adds an unread function.
type Iter[T any] struct {
	next    func() (T, error, bool) // The underlying iterator
	head    T                       // The last read element
	hasHead bool                    // Unread was called
}

// Next returns the next element,
// or the last unread element with a nil error.
func (r *Iter[T]) Next() (T, error, bool) {
	if r.hasHead {
		r.hasHead = false
		return r.head, nil, true
	}
	t, err, ok := r.next()
	if !ok {
		return t, nil, false
	}
	r.head = t
	return t, err, true
}

// Unread makes the next call to [Next] return the last element and a nil error.
// Can be called up to once per call to [Next].
func (r *Iter[T]) Unread() {
	if r.hasHead {
		panic(fmt.Sprintf("called Unread twice: first with %v", r.head))
	}
	r.hasHead = true
}

// Until calls [Next] until an error is returned or stop returns true.
// Returns the encountered error or nil if stop returned true.
func (r *Iter[T]) Until(stop func(T) bool) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			t, err, ok := r.Next()
			if !ok {
				return
			}
			if err != nil {
				if !yield(t, err) {
					return
				}
			}
			if stop(t) {
				r.Unread()
				return
			}
			if !yield(t, nil) {
				return
			}
		}
	}
}

// New returns an Unreader with read as its underlying read function.
func New[T any](seq iter.Seq2[T, error]) *Iter[T] {
	next, _ := iter.Pull2(seq)
	return &Iter[T]{next: next, hasHead: false}
}
