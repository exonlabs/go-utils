// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package gx

// Ordered is a constraint that allows any ordered type, i.e.,
// types that support comparison operators like <, <=, >=, and >.
// This includes various numeric types (integers and floats) and strings.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Max returns the largest value among a fixed number of arguments of type [Ordered].
// At least one argument must be passed, and it returns that if it's the only one.
// It compares all the provided values and returns the maximum value.
//
//	max := Max(3, 5, 2)   // max is 5
func Max[T Ordered](first T, rest ...T) T {
	r := first
	for _, v := range rest {
		if v > r {
			r = v
		}
	}
	return r
}

// Min returns the smallest value among a fixed number of arguments of type [Ordered].
// At least one argument must be passed, and it returns that if it's the only one.
// It compares all the provided values and returns the minimum value.
//
//	min := Min(3, 5, 2)   // min is 2
func Min[T Ordered](first T, rest ...T) T {
	r := first
	for _, v := range rest {
		if v < r {
			r = v
		}
	}
	return r
}
