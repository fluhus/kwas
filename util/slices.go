package util

// MapSlice returns a slice of the same length containing the results
// of applying f to the elements of s.
func MapSlice[S any, T any](s []S, f func(S) T) []T {
	t := make([]T, len(s))
	for i := range s {
		t[i] = f(s[i])
	}
	return t
}

// FilterSlice returns a slice containing only the elements
// for which f returns true.
func FilterSlice[S any](s []S, f func(S) bool) []S {
	var result []S
	for _, e := range s {
		if f(e) {
			result = append(result, e)
		}
	}
	return result
}
