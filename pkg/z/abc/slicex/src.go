// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xslice

// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Empty and nil slices are considered equal.
// Floating point NaNs are not considered equal.
// (go 1.21)
func Equal[S ~[]E, E comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
// (go 1.21)
func Index[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// Reverse reverses the elements of slice s in place.
// (go 1.21)
func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ReverseCopy reverses the elements of s and return a copy.
func ReverseCopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	Reverse(b)
	return b
}

// SplitN splits iterable s into slices of fixed length n.
func SplitN[S ~[]E, E any](s S, n int) []S {
	var r []S
	l := len(s)
	for i := 0; i < l; i += n {
		end := i + n
		if end > l {
			end = l
		}
		r = append(r, s[i:end])
	}
	return r
}
