// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package slicex

// Equal reports whether two slices are equal, meaning they have the same length
// and all corresponding elements are equal.
// If the slices have different lengths, Equal returns false immediately.
// The comparison proceeds in increasing index order, stopping at the first
// unequal pair.
// Empty and nil slices are considered equal, while floating point NaNs are
// not considered equal.
//
// Example:
//
//	Equal([]int{1, 2, 3}, []int{1, 2, 3}) // returns true
//	Equal([]int{1, 2}, []int{1, 2, 3})    // returns false
//
// (Introduced in Go 1.21)
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

// Index returns the index of the first occurrence of the value v in the slice s,
// or -1 if the value is not present in the slice.
//
// Example:
//
//	Index([]string{"apple", "banana", "cherry"}, "banana") // returns 1
//	Index([]string{"apple", "banana", "cherry"}, "grape")  // returns -1
//
// (Introduced in Go 1.21)
func Index[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// Reverse reverses the elements of the slice s in place.
// This operation modifies the original slice, reversing the order of its
// elements without allocating additional space.
//
// Example:
//
//	s := []int{1, 2, 3, 4}
//	Reverse(s) // s is now []int{4, 3, 2, 1}
//
// (Introduced in Go 1.21)
func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ReverseCopy creates a new slice that is a reversed copy of the original slice s.
// It does not modify the original slice, returning a new slice with elements
// in reverse order.
//
// Example:
//
//	s := []int{1, 2, 3}
//	reversed := ReverseCopy(s) // reversed is now []int{3, 2, 1}
//
// (The original slice s remains unchanged)
func ReverseCopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	Reverse(b)
	return b
}

// SplitN splits the iterable slice s into slices of fixed length n.
// If the length of s is not a multiple of n, the last slice will contain
// the remaining elements. Returns a slice of slices.
//
// Example:
//
//	s := []int{1, 2, 3, 4, 5}
//	result := SplitN(s, 2) // result is now [][]int{{1, 2}, {3, 4}, {5}}
//
// If n is less than or equal to 0, it returns an empty slice.
func SplitN[S ~[]E, E any](s S, n int) []S {
	if n <= 0 {
		return nil
	}
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
