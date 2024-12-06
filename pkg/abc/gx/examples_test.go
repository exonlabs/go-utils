// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package gx_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/gx"
)

func ExampleMax() {
	// Finding the max value
	maxInt := gx.Max(3, 1, 4, 2)        // maxInt will be 4
	maxFloat := gx.Max(5.1, 2.3, 8.7)   // maxFloat will be 8.7
	maxStr := gx.Max("apple", "banana") // maxStr will be "banana"

	fmt.Println(maxInt, maxFloat, maxStr)
	// Output:
	// 4 8.7 banana
}

func ExampleMin() {
	// Finding the min value
	minInt := gx.Min(3, 1, 4, 2)        // minInt will be 1
	minFloat := gx.Min(5.1, 2.3, 8.7)   // minFloat will be 2.3
	minStr := gx.Min("apple", "banana") // minStr will be "apple"

	fmt.Println(minInt, minFloat, minStr)
	// Output:
	// 1 2.3 apple
}
