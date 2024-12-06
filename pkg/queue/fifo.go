// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package queue

import (
	"sync"
)

// Fifo represents a dynamic, memory-efficient queue with a minimum fixed size.
type Fifo struct {
	buffer        []any
	count         int
	start, end    int
	size, minSize int
	mu            sync.Mutex
}

// NewFifo creates a new fifo data queue with a minimum fixed size.
func NewFifo(minSize int) *Fifo {
	return &Fifo{
		buffer:  make([]any, minSize),
		size:    minSize,
		minSize: minSize,
	}
}

// Size returns the current size of the queue.
func (q *Fifo) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

// Length returns the current number of items in the queue.
func (q *Fifo) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count
}

// IsEmpty returns true if the queue is empty.
func (q *Fifo) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count == 0
}

// Push adds data to the end of the queue, nil data items are ignored.
// Automatically resizes to a larger size if the queue is full.
func (q *Fifo) Push(data ...any) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	n := len(data)
	if n > q.size-q.count {
		q.resize(n + q.count)
	}

	for _, v := range data {
		if v != nil {
			q.buffer[q.end] = v
			q.end = (q.end + 1) % q.size
			q.count++
		}
	}

	return nil
}

// Pop removes and returns data from the start of the queue.
// Returns nil if the queue is empty.
// Shrinks the queue to minimum size if data length is less than minimum size.
func (q *Fifo) Pop() any {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count == 0 {
		return nil
	}

	v := q.buffer[q.start]
	q.start = (q.start + 1) % q.size
	q.count--

	if q.count <= q.minSize && q.size > q.minSize {
		q.resize(q.minSize)
	}

	return v
}

// PopN removes and returns data with length n from the start of the queue.
// May return data less than n when n > count.
// Shrinks the queue to minimum size if remaining data size is less than the minimum size.
func (q *Fifo) PopN(n int) []any {
	q.mu.Lock()
	defer q.mu.Unlock()

	if n > q.count {
		n = q.count
	}

	data := make([]any, n)
	for i := 0; i < n; i++ {
		data[i] = q.buffer[q.start]
		q.start = (q.start + 1) % q.size
	}
	q.count -= n

	if q.count <= q.minSize && q.size > q.minSize {
		q.resize(q.minSize)
	}

	return data
}

// resize adjusts the size of the internal buffer to at least newSize.
func (q *Fifo) resize(newSize int) {
	newBuffer := make([]any, newSize)
	if q.count > 0 {
		if q.start < q.end {
			copy(newBuffer, q.buffer[q.start:q.end])
		} else {
			n := copy(newBuffer, q.buffer[q.start:])
			copy(newBuffer[n:], q.buffer[:q.end])
		}
	}
	q.buffer = newBuffer
	q.start = 0
	q.end = q.count
	q.size = newSize
}
