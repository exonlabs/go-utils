// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package mapx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/abc/mapx"
)

func TestFind(t *testing.T) {
	intMap := map[int]string{
		1: "apple",
		2: "banana",
		3: "cherry",
	}

	key, found := mapx.Find(intMap, "banana")
	assert.True(t, found, "Value 'banana' should be found")
	assert.Equal(t, 2, key, "The key for 'banana' should be 2")

	key, found = mapx.Find(intMap, "grape")
	assert.False(t, found, "Value 'grape' should not be found")
	assert.Equal(t, 0, key, "The key should be 0 for a non-found value")

	// Test with a map of strings
	strMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	keyStr, found := mapx.Find(strMap, 3)
	assert.True(t, found, "Value '3' should be found")
	assert.Equal(t, "three", keyStr, "The key for '3' should be 'three'")

	keyStr, found = mapx.Find(strMap, 4)
	assert.False(t, found, "Value '4' should not be found")
	assert.Equal(t, "", keyStr, "The key should be an empty string for a non-found value")
}
