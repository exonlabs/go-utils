// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package x

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
