// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package slicex_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/abc/slicex"
)

func TestEqual(t *testing.T) {
	// Test equal slices
	assert.True(t, slicex.Equal([]int{1, 2, 3}, []int{1, 2, 3}))
	assert.True(t, slicex.Equal([]string{"a", "b"}, []string{"a", "b"}))

	// Test unequal slices
	assert.False(t, slicex.Equal([]int{1, 2, 3}, []int{1, 2, 4}))
	assert.False(t, slicex.Equal([]int{1}, []int{1, 2}))

	// Test empty and nil slices
	assert.True(t, slicex.Equal([]int{}, []int{}))
	assert.True(t, slicex.Equal([]int{}, nil))
	assert.True(t, slicex.Equal(nil, []int{}))
}

func TestIndex(t *testing.T) {
	// Test existing value
	assert.Equal(t, 1,
		slicex.Index([]string{"apple", "banana", "cherry"}, "banana"))

	// Test non-existing value
	assert.Equal(t, -1,
		slicex.Index([]string{"apple", "banana", "cherry"}, "grape"))

	// Test empty slice
	assert.Equal(t, -1,
		slicex.Index([]string{}, "banana"))
}

func TestReverse(t *testing.T) {
	// Test reversing a normal slice
	s := []int{1, 2, 3, 4}
	slicex.Reverse(s)
	assert.Equal(t, []int{4, 3, 2, 1}, s)

	// Test reversing an empty slice
	s = []int{}
	slicex.Reverse(s)
	assert.Equal(t, []int{}, s)

	// Test reversing a single-element slice
	s = []int{42}
	slicex.Reverse(s)
	assert.Equal(t, []int{42}, s)
}

func TestReverseCopy(t *testing.T) {
	// Test reversing and copying a slice
	s := []int{1, 2, 3}
	reversed := slicex.ReverseCopy(s)
	assert.Equal(t, []int{3, 2, 1}, reversed)
	assert.Equal(t, []int{1, 2, 3}, s) // Original slice should remain unchanged

	// Test reversing and copying an empty slice
	s = []int{}
	reversed = slicex.ReverseCopy(s)
	assert.Equal(t, []int{}, reversed)

	// Test reversing and copying a single-element slice
	s = []int{42}
	reversed = slicex.ReverseCopy(s)
	assert.Equal(t, []int{42}, reversed)
}

func TestSplitN(t *testing.T) {
	// Test splitting into fixed-length slices
	s := []int{1, 2, 3, 4, 5}
	result := slicex.SplitN(s, 2)
	assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, result)

	// Test splitting with a slice of length less than n
	s = []int{1}
	result = slicex.SplitN(s, 2)
	assert.Equal(t, [][]int{{1}}, result)

	// Test splitting with n = 0 (should return nil)
	result = slicex.SplitN(s, 0)
	assert.Nil(t, result)

	// Test splitting with n < 0 (should return nil)
	result = slicex.SplitN(s, -1)
	assert.Nil(t, result)

	// Test splitting an empty slice
	s = []int{}
	result = slicex.SplitN(s, 2)
	assert.Empty(t, result)
}
