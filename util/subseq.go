package util

type SubseqIter[T any] struct {
	s    []T
	i, n int
}

func (s *SubseqIter[T]) Next() []T {
	if s.i > len(s.s)-s.n {
		return nil
	}
	r := s.s[s.i : s.i+s.n]
	s.i++
	return r
}

func Subseqs[T any](s []T, n int) ([]T, *SubseqIter[T]) {
	ss := &SubseqIter[T]{s, 0, n}
	return ss.Next(), ss
}
