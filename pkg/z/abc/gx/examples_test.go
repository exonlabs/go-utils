// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package x_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/x"
)

func ExampleMax() {
	fmt.Println(x.Max(2, 3))
	fmt.Println(x.Max(1, -2, -3, 0))
	fmt.Println(x.Max(1.3, 2.1, -3.5))

	// Output:
	// 3
	// 1
	// 2.1
}

func ExampleMin() {
	fmt.Println(x.Min(1, 3))
	fmt.Println(x.Min(1, -2, -3, 0))
	fmt.Println(x.Min(1.3, 2.1, -3.5))

	// Output:
	// 1
	// -3
	// -3.5
}
