// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package dictx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	original := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": "e",
				},
			},
		},
	}
	cloned, err := Clone(original)
	assert.Nil(t, err)
	assert.Equal(t, original, cloned)

	// Modifying cloned dict shouldn't affect original
	cloned["a"].(Dict)["b"].(Dict)["c"].(Dict)["d"] = "modified"
	assert.Equal(t, "e", original["a"].(Dict)["b"].(Dict)["c"].(Dict)["d"])
}

func TestString(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b1": Dict{
				"c": Dict{
					"d1": "value",
					"d2": "value",
				},
			},
			"b2": Dict{
				"c": Dict{
					"d": "value",
				},
			},
		},
	}
	s := String(d)
	expected := "{a: {b1: {c: {d1: value, d2: value}}, b2: {c: {d: value}}}}"
	assert.Equal(t, expected, s)
}

func TestKeysN(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": "value",
				},
			},
		},
	}
	keys := KeysN(d, -1)
	expected := []string{"a.b.c.d"}
	assert.ElementsMatch(t, expected, keys)

	// Test for limited depth
	keysLevel1 := KeysN(d, 1)
	expectedLevel1 := []string{"a"}
	assert.ElementsMatch(t, expectedLevel1, keysLevel1)

	keysLevel2 := KeysN(d, 2)
	expectedLevel2 := []string{"a.b"}
	assert.ElementsMatch(t, expectedLevel2, keysLevel2)

	keysLevel3 := KeysN(d, 3)
	expectedLevel3 := []string{"a.b.c"}
	assert.ElementsMatch(t, expectedLevel3, keysLevel3)
}

func TestIsExist(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": "value",
				},
			},
		},
	}
	assert.True(t, IsExist(d, "a.b.c.d"))
	assert.False(t, IsExist(d, "a.b.c.e"))
	assert.False(t, IsExist(d, "a.x"))
	assert.False(t, IsExist(d, "x.y.z"))
}

func TestGet(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": "value",
				},
			},
		},
	}
	assert.Equal(t, "value", Get(d, "a.b.c.d", "default"))
	assert.Equal(t, "default", Get(d, "a.b.c.e", "default"))
	assert.Equal(t, "default", Get(d, "a.b.x", "default"))
}

func TestFetch(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": 42,
				},
			},
		},
	}
	assert.Equal(t, 42, Fetch(d, "a.b.c.d", 0))
	assert.Equal(t, 0, Fetch(d, "a.b.c.e", 0))
	assert.Equal(t, 0, Fetch(d, "a.b.x", 0))
}

func TestSet(t *testing.T) {
	d := Dict{}
	Set(d, "a.b.c.d", "value")
	assert.True(t, IsExist(d, "a.b.c.d"))
	assert.Equal(t, "value", Get(d, "a.b.c.d", "default"))

	// Test overwriting existing value
	Set(d, "a.b.c.d", "new_value")
	assert.Equal(t, "new_value", Get(d, "a.b.c.d", "default"))
}

func TestMerge(t *testing.T) {
	src := Dict{
		"a": Dict{
			"b": Dict{
				"c": "old_value",
			},
		},
	}
	updt := Dict{
		"a": Dict{
			"b": Dict{
				"c": "new_value",
				"d": "new_key",
			},
		},
	}
	Merge(src, updt)
	assert.Equal(t, "new_value", Get(src, "a.b.c", "default"))
	assert.Equal(t, "new_key", Get(src, "a.b.d", "default"))
}

func TestDelete(t *testing.T) {
	d := Dict{
		"a": Dict{
			"b": Dict{
				"c": Dict{
					"d": "value",
				},
			},
		},
	}
	Delete(d, "a.b.c.d")
	assert.False(t, IsExist(d, "a.b.c.d"))
	assert.True(t, IsExist(d, "a.b.c"))
}
