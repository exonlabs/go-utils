// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package slicex_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/slicex"
)

func ExampleEqual() {
	// Using Equal to compare slices
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	slice3 := []int{1, 2, 4}
	fmt.Println(slicex.Equal(slice1, slice2))
	fmt.Println(slicex.Equal(slice1, slice3))

	// Output:
	// true
	// false
}

func ExampleIndex() {
	// Finding the index of an element in a slice
	slice := []string{"apple", "banana", "cherry"}
	index := slicex.Index(slice, "banana")
	fmt.Println(index)

	index = slicex.Index(slice, "grape")
	fmt.Println(index)

	// Output:
	// 1
	// -1
}

func ExampleReverse() {
	// Reversing a slice in place
	slice := []int{1, 2, 3, 4}
	slicex.Reverse(slice)
	fmt.Println(slice)

	// Output: [4 3 2 1]
}

func ExampleReverseCopy() {
	// Creating a reversed copy of a slice
	slice := []int{1, 2, 3}
	reversed := slicex.ReverseCopy(slice)
	fmt.Println(reversed)
	fmt.Println(slice)

	// Output:
	// [3 2 1]
	// [1 2 3]
}

func ExampleSplitN() {
	// Splitting a slice into fixed-length sub-slices
	slice := []int{1, 2, 3, 4, 5}
	result := slicex.SplitN(slice, 2)
	fmt.Println(result)

	// Output:
	// [[1 2] [3 4] [5]]
}
