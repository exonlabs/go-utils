// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package dictx_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
)

func ExampleClone() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": "value",
		},
	}
	cloned, err := dictx.Clone(d)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// Print original and cloned dictionary
	fmt.Println("Original:", d)
	fmt.Println("Cloned:", cloned)

	// Output:
	// Original: map[a:map[b:value]]
	// Cloned: map[a:map[b:value]]
}

func ExampleKeysN() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": dictx.Dict{
				"c": dictx.Dict{
					"d": "value",
				},
			},
		},
	}

	// Retrieve all keys with unlimited depth
	keys := dictx.KeysN(d, -1)
	fmt.Println(keys)

	// Retrieve top-level keys only
	keysTopLevel := dictx.KeysN(d, 1)
	fmt.Println(keysTopLevel)

	// Output:
	// [a.b.c.d]
	// [a]
}

func ExampleIsExist() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": dictx.Dict{
				"c": "value",
			},
		},
	}

	// Check if keys exist
	fmt.Println(dictx.IsExist(d, "a.b.c")) // true
	fmt.Println(dictx.IsExist(d, "a.b.d")) // false
	fmt.Println(dictx.IsExist(d, "a.x"))   // false

	// Output:
	// true
	// false
	// false
}

func ExampleGet() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": dictx.Dict{
				"c": "value",
			},
		},
	}

	// Get an existing key
	val := dictx.Get(d, "a.b.c", "default")
	fmt.Println(val)

	// Get a non-existing key, return default value
	val = dictx.Get(d, "a.b.x", "default")
	fmt.Println(val)

	// Output:
	// value
	// default
}

func ExampleFetch() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": dictx.Dict{
				"c": 42,
			},
		},
	}

	// Fetch value with correct type (int)
	val := dictx.Fetch(d, "a.b.c", 0)
	fmt.Println(val)

	// Fetch value that doesn't exist (should return default value)
	val = dictx.Fetch(d, "a.b.d", 0)
	fmt.Println(val)

	// Output:
	// 42
	// 0
}

func ExampleSet() {
	d := dictx.Dict{}

	// Set a nested value
	dictx.Set(d, "a.b.c.d", "new_value")

	// Get the value to confirm it was set correctly
	val := dictx.Get(d, "a.b.c.d", "default")
	fmt.Println(val)

	// Output:
	// new_value
}

func ExampleMerge() {
	src := dictx.Dict{
		"a": dictx.Dict{
			"b": "old_value",
		},
	}
	updt := dictx.Dict{
		"a": dictx.Dict{
			"b": "new_value",
			"c": "new_key",
		},
	}

	// Merge the update dict into the source dict
	dictx.Merge(src, updt)

	// Print the updated source dictionary
	fmt.Println(src)

	// Output:
	// map[a:map[b:new_value c:new_key]]
}

func ExampleDelete() {
	d := dictx.Dict{
		"a": dictx.Dict{
			"b": dictx.Dict{
				"c": "value",
			},
		},
	}

	// Delete a key
	dictx.Delete(d, "a.b.c")

	// Check if the key was deleted
	exists := dictx.IsExist(d, "a.b.c")
	fmt.Println(exists)

	// Output:
	// false
}
