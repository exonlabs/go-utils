// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xutils

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// (go 1.21)
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Max function returns the largest value of a fixed number of arguments
// of [Ordered] types. There must be at least one argument.
// (go 1.21)
func Max[T Ordered](a T, b ...T) T {
	r := a
	for _, v := range b {
		if v > r {
			r = v
		}
	}
	return r
}

// Min function returns the smallest value of a fixed number of arguments
// of [Ordered] types. There must be at least one argument.
// (go 1.21)
func Min[T Ordered](a T, b ...T) T {
	r := a
	for _, v := range b {
		if v < r {
			r = v
		}
	}
	return r
}

// SliceEqual reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, SliceEqual returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Empty and nil slices are considered equal.
// Floating point NaNs are not considered equal.
// (go 1.21)
func SliceEqual[S ~[]E, E comparable](s1, s2 S) bool {
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

// SliceIndex returns the index of the first occurrence of v in s,
// or -1 if not present.
// (go 1.21)
func SliceIndex[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// SliceReverse reverses the elements of slice s in place.
// (go 1.21)
func SliceReverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// SliceReverseCopy reverses the elements of s and return a copy.
func SliceReverseCopy[S ~[]T, T any](s S) S {
	b := make([]T, len(s))
	copy(b, s)
	SliceReverse(b)
	return b
}

// SliceSplitN splits iterable s into slices of fixed length n.
func SliceSplitN[S ~[]E, E any](s S, n int) []S {
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

// MapFind finds the key of the first match value v in map m.
// returns the key if value is found and bool status indication.
func MapFind[M ~map[T]E, T, E comparable](m M, v E) (T, bool) {
	var r T
	for key, val := range m {
		if val == v {
			return key, true
		}
	}
	return r, false
}
