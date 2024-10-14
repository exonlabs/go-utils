// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xdict_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xdict"
)

func ExampleKeys() {
	d := map[string]any{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": xdict.Dict{"1": "xxx", "2": "yyy"},
		"k4": map[string]any{
			"1": "xxx",
			"2": xdict.Dict{"1": "xxx", "2": "yyy"},
			"3": map[string]any{
				"1": "xxx",
				"2": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	}

	fmt.Printf("Keys: %v\n", xdict.Keys(d))

	// Output:
	// Keys: [k1 k2.1 k2.2 k3.1 k3.2 k4.1 k4.2.1 k4.2.2 k4.3.1 k4.3.2.1 k4.3.2.2]
}

func ExampleKeysN() {
	d := map[string]any{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": xdict.Dict{"1": "xxx", "2": "yyy"},
		"k4": map[string]any{
			"1": "xxx",
			"2": xdict.Dict{"1": "xxx", "2": "yyy"},
			"3": map[string]any{
				"1": "xxx",
				"2": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	}

	fmt.Printf("Keys Level 1: %v\n", xdict.KeysN(d, 1))
	fmt.Printf("Keys Level 2: %v\n", xdict.KeysN(d, 2))
	fmt.Printf("Keys Level 3: %v\n", xdict.KeysN(d, 3))
	fmt.Printf("Keys Level 4: %v\n", xdict.KeysN(d, 4))

	// Output:
	// Keys Level 1: [k1 k2 k3 k4]
	// Keys Level 2: [k1 k2.1 k2.2 k3.1 k3.2 k4.1 k4.2 k4.3]
	// Keys Level 3: [k1 k2.1 k2.2 k3.1 k3.2 k4.1 k4.2.1 k4.2.2 k4.3.1 k4.3.2]
	// Keys Level 4: [k1 k2.1 k2.2 k3.1 k3.2 k4.1 k4.2.1 k4.2.2 k4.3.1 k4.3.2.1 k4.3.2.2]
}

func ExampleIsExist() {
	d := map[string]any{
		"k1": "some value",
		"k2": map[string]any{"1": "xxx", "2": "yyy"},
		"k3": xdict.Dict{"1": "xxx", "2": "yyy"},
		"k4": map[string]any{
			"1": "xxx",
			"2": xdict.Dict{"1": "xxx", "2": "yyy"},
			"3": map[string]any{
				"1": "xxx",
				"2": map[string]any{"1": "xxx", "2": "yyy"},
			},
		},
	}

	for _, k := range []string{"k1", "k2.2", "k4.2.1", "k5", "k4.3.3"} {
		fmt.Printf("key \"%s\" --> %v\n", k, xdict.IsExist(d, k))
	}

	// Output:
	// key "k1" --> true
	// key "k2.2" --> true
	// key "k4.2.1" --> true
	// key "k5" --> false
	// key "k4.3.3" --> false
}
