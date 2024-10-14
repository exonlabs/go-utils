// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xutils_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xutils"
)

func ExampleMax() {
	fmt.Println(xutils.Max(2, 3))
	fmt.Println(xutils.Max(1, -2, -3, 0))
	fmt.Println(xutils.Max(1.3, 2.1, -3.5))

	// Output:
	// 3
	// 1
	// 2.1
}

func ExampleMin() {
	fmt.Println(xutils.Min(1, 3))
	fmt.Println(xutils.Min(1, -2, -3, 0))
	fmt.Println(xutils.Min(1.3, 2.1, -3.5))

	// Output:
	// 1
	// -3
	// -3.5
}

func ExampleSliceEqual() {
	s1 := []string{"matching", "strings"}
	s2 := []string{"matching", "strings"}
	s3 := []string{"non", "matching", "strings"}
	fmt.Println("s1 equal s2 -->", xutils.SliceEqual(s1, s2))
	fmt.Println("s1 equal s3 -->", xutils.SliceEqual(s1, s3))

	b1 := []byte{1, 2, 3}
	b2 := []byte{1, 2, 3}
	b3 := []byte{1, 2, 3, 4}
	fmt.Println("b1 equal b2 -->", xutils.SliceEqual(b1, b2))
	fmt.Println("b1 equal b3 -->", xutils.SliceEqual(b1, b3))

	// Output:
	// s1 equal s2 --> true
	// s1 equal s3 --> false
	// b1 equal b2 --> true
	// b1 equal b3 --> false
}

func ExampleSliceIndex() {
	s1 := []string{"1", "2", "3", "3"}
	s2 := []int{1, 2, 3, 3, 4, 4, 5}
	s3 := []any{1, 2, "3", 3, 4, 4.0, false}

	fmt.Println("s1 index of \"3\" -->", xutils.SliceIndex(s1, "3"))
	fmt.Println("s1 index of \"3\" -->", xutils.SliceIndex(s1, "8"))

	fmt.Println("s2 index of 4 -->", xutils.SliceIndex(s2, 4))
	fmt.Println("s2 index of 8 -->", xutils.SliceIndex(s2, 8))

	fmt.Println("s3 index of 3 -->", xutils.SliceIndex(s3, 3))
	fmt.Println("s3 index of false -->", xutils.SliceIndex(s3, false))

	// Output:
	// s1 index of "3" --> 2
	// s1 index of "3" --> -1
	// s2 index of 4 --> 4
	// s2 index of 8 --> -1
	// s3 index of 3 --> 3
	// s3 index of false --> 6
}

func ExampleSliceReverse() {
	s1 := []string{"1", "2", "3", "3"}
	s2 := []any{1, 2, "3", 3, 4, 4.0, false}

	fmt.Println("s1 =", s1)
	xutils.SliceReverse(s1)
	fmt.Println("s1 =", s1)

	fmt.Println("s2 =", s2)
	xutils.SliceReverse(s2)
	fmt.Println("s2 =", s2)

	// Output:
	// s1 = [1 2 3 3]
	// s1 = [3 3 2 1]
	// s2 = [1 2 3 3 4 4 false]
	// s2 = [false 4 4 3 3 2 1]
}

func ExampleSliceReverseCopy() {
	s1 := []string{"1", "2", "3", "3"}
	s2 := []any{1, 2, "3", 3, 4, 4.0, false}

	fmt.Println("s1 =", s1)
	c1 := xutils.SliceReverseCopy(s1)
	fmt.Println("s1 =", s1) // not changed
	fmt.Println("c1 =", c1)

	fmt.Println("s2 =", s2)
	c2 := xutils.SliceReverseCopy(s2)
	fmt.Println("s2 =", s2) // not changed
	fmt.Println("c2 =", c2)

	// Output:
	// s1 = [1 2 3 3]
	// s1 = [1 2 3 3]
	// c1 = [3 3 2 1]
	// s2 = [1 2 3 3 4 4 false]
	// s2 = [1 2 3 3 4 4 false]
	// c2 = [false 4 4 3 3 2 1]
}

func ExampleSliceSplitN() {
	s1 := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	s2 := []any{"1", 2, 3.0, true, "a", -1, nil}

	fmt.Println("split s1 n=3 -->", xutils.SliceSplitN(s1, 3))
	fmt.Println("split s2 n=2 -->", xutils.SliceSplitN(s2, 2))

	// Output:
	// split s1 n=3 --> [[1 2 3] [4 5 6] [7 8]]
	// split s2 n=2 --> [[1 2] [3 true] [a -1] [<nil>]]
}

func ExampleMapFind() {
	m1 := map[int]string{
		1: "number 1",
		2: "number 2",
		3: "number 3",
	}
	fmt.Println(xutils.MapFind(m1, "number 2"))
	fmt.Println(xutils.MapFind(m1, "number 4"))

	m2 := map[string]any{
		"a": "number 1",
		"b": 2,
		"c": 3.0,
	}
	fmt.Println(xutils.MapFind(m2, 2))
	fmt.Println(xutils.MapFind(m2, "number 1"))

	// Output:
	// 2 true
	// 0 false
	// b true
	// a true
}
