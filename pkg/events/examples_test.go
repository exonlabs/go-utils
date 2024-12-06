// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package events_test

import (
	"fmt"
	"time"

	"github.com/exonlabs/go-utils/pkg/events"
)

func ExampleEvent_Set() {
	e := events.New()

	// Initially, the event is not set
	fmt.Println(e.IsSet())

	// Set the event
	e.Set()
	fmt.Println(e.IsSet())

}

func ExampleEvent_Clear() {
	e := events.New()
	e.Set()
	fmt.Println(e.IsSet())

	// Clear the event
	e.Clear()
	fmt.Println(e.IsSet())

	// Output:
	// true
	// false
}

func ExampleEvent_Wait() {
	e := events.New()

	// Start a goroutine that sets the event after a delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		e.Set()
	}()

	// Wait for the event to be set with a timeout
	if e.Wait(1.0) {
		fmt.Println("Timed out") // Should not reach this if the event is set in time
	} else {
		fmt.Println("Event set before timeout") // Expected output
	}

	// Output:
	// Event set before timeout
}
