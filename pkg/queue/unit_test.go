// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package queue_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/queue"
)

func TestFifo_PushAndPop(t *testing.T) {
	q := queue.NewFifo(2)

	// Test initial size of queue
	assert.Equal(t, 2, q.Size())

	// Test pushing and popping a single element
	q.Push(1)
	assert.Equal(t, 2, q.Size())
	assert.Equal(t, 1, q.Length())
	assert.False(t, q.IsEmpty())

	assert.Equal(t, 1, q.Pop())
	assert.Equal(t, 0, q.Length())
	assert.True(t, q.IsEmpty())

	// Test pushing multiple elements and popping them
	q.Push(2, true, 4.5, map[string]int{"a": 1, "b": 2}, []int{1, 2, 3})
	assert.Equal(t, 5, q.Size())
	assert.Equal(t, 5, q.Length())

	assert.Equal(t, 2, q.Pop())
	assert.Equal(t, 4, q.Length())
	assert.Equal(t, 5, q.Size())

	assert.Equal(t, true, q.Pop())
	assert.Equal(t, 3, q.Length())
	assert.Equal(t, 5, q.Size())

	assert.Equal(t, 4.5, q.Pop())
	assert.Equal(t, 2, q.Length())
	assert.Equal(t, 2, q.Size())

	assert.Equal(t, map[string]int{"a": 1, "b": 2}, q.Pop())
	assert.Equal(t, 1, q.Length())
	assert.Equal(t, 2, q.Size())

	assert.Equal(t, []int{1, 2, 3}, q.Pop())
	assert.Equal(t, 0, q.Length())
	assert.Equal(t, 2, q.Size())
}

func TestFifo_PopN(t *testing.T) {
	q := queue.NewFifo(2)
	q.Push(1, 2, 3, 4, 5)

	// Test popping multiple elements
	result := q.PopN(3)
	assert.Equal(t, []any{1, 2, 3}, result)
	assert.Equal(t, 2, q.Length())
	assert.Equal(t, 2, q.Size())
	assert.False(t, q.IsEmpty())

	// Test popping more elements than present
	result = q.PopN(3)
	assert.Equal(t, []any{4, 5}, result)
	assert.Equal(t, 0, q.Length())
	assert.Equal(t, 2, q.Size())
	assert.True(t, q.IsEmpty())
}

func TestFifo_Resize(t *testing.T) {
	q := queue.NewFifo(2)

	// Test resizing up
	q.Push(1, 2, 3)
	assert.Equal(t, 3, q.Length())
	assert.Equal(t, 3, q.Size())

	// Test resizing down
	q.Pop()
	q.Pop()
	q.Pop()
	assert.Equal(t, 0, q.Length())
}

func TestFifo_Concurrency(t *testing.T) {
	q := queue.NewFifo(1000)
	var wg sync.WaitGroup

	// Test concurrent pushes
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			q.Push(i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 500; i < 1000; i++ {
			q.Push(i)
		}
	}()
	wg.Wait()

	assert.Equal(t, 1000, q.Length())

	// Test concurrent pops
	wg.Add(2)
	results := make([]any, 1000)
	go func() {
		defer wg.Done()
		for i := 0; i < 500; i++ {
			results[i] = q.Pop()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 500; i < 1000; i++ {
			results[i] = q.Pop()
		}
	}()
	wg.Wait()

	assert.Equal(t, 0, q.Length())
	assert.True(t, q.IsEmpty())
}

func TestFifo_EmptyPop(t *testing.T) {
	q := queue.NewFifo(2)
	assert.Nil(t, q.Pop())
	assert.Equal(t, []any{}, q.PopN(3))
}
