// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xmap_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/types/xmap"
)

func ExampleMapFind() {
	m1 := map[int]string{
		1: "number 1",
		2: "number 2",
		3: "number 3",
	}
	fmt.Println(xmap.Find(m1, "number 2"))
	fmt.Println(xmap.Find(m1, "number 4"))

	m2 := map[string]any{
		"a": "number 1",
		"b": 2,
		"c": 3.0,
	}
	fmt.Println(xmap.Find(m2, 2))
	fmt.Println(xmap.Find(m2, "number 1"))

	// Output:
	// 2 true
	// 0 false
	// b true
	// a true
}
