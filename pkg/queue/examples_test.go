// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package queue_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/queue"
)

// ExampleFifo_Push demonstrates the usage of the Push method.
func ExampleFifo_Push() {
	q := queue.NewFifo(2)
	q.Push(1, 2, 3)
	fmt.Println(q.Length())

	// Output: 3
}

// ExampleFifo_Pop demonstrates the usage of the Pop method.
func ExampleFifo_Pop() {
	q := queue.NewFifo(2)
	q.Push(1, 2, 3)
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())

	// Output:
	// 1
	// 2
}

// ExampleFifo_PopN demonstrates the usage of the PopN method.
func ExampleFifo_PopN() {
	q := queue.NewFifo(2)
	q.Push(1, 2, 3, 4)
	fmt.Println(q.PopN(3))

	// Output: [1 2 3]
}

// ExampleFifo_IsEmpty demonstrates the usage of the IsEmpty method.
func ExampleFifo_IsEmpty() {
	q := queue.NewFifo(2)
	fmt.Println(q.IsEmpty())
	q.Push(1)
	fmt.Println(q.IsEmpty())

	// Output:
	// true
	// false
}

// ExampleFifo_Length demonstrates the usage of the Length method.
func ExampleFifo_Length() {
	q := queue.NewFifo(2)
	fmt.Println(q.Length())
	q.Push(1, 2)
	fmt.Println(q.Length())

	// Output:
	// 0
	// 2
}
