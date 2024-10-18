// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xevent_test

import (
	"fmt"
	"time"

	"github.com/exonlabs/go-utils/pkg/xevent"
)

func ExampleNew() {
	evt := xevent.New()

	fmt.Printf("initial status: %v\n", evt.IsSet())

	go func() {
		fmt.Printf("routine started, waiting event ...\n")
		for !evt.IsSet() {
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Printf("routine end\n")
	}()

	time.Sleep(10 * time.Millisecond)

	evt.Set()
	fmt.Printf("event set: %v\n", evt.IsSet())
	time.Sleep(20 * time.Millisecond)
	fmt.Printf("exit\n")

	// Output:
	// initial status: false
	// routine started, waiting event ...
	// event set: true
	// routine end
	// exit
}
