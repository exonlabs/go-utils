// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package mapx_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/mapx"
)

func ExampleFind() {
	m := map[string]int{
		"first":  1,
		"second": 2,
		"third":  3,
	}

	// Find the key for the value 2
	key, found := mapx.Find(m, 2)
	if found {
		fmt.Println("Found key:", key)
	} else {
		fmt.Println("Value not found")
	}

	// Find a non-existing value
	key, found = mapx.Find(m, 4)
	if !found {
		fmt.Println("Value not found")
	}

	// Output:
	// Found key: second
	// Value not found
}
