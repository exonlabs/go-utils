// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package events_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/exonlabs/go-utils/pkg/events"
)

func TestNewEvent(t *testing.T) {
	e := events.New()
	assert.NotNil(t, e)
	assert.False(t, e.IsSet())
}

func TestSet(t *testing.T) {
	e := events.New()

	// Set the event
	e.Set()
	assert.True(t, e.IsSet())

	// Set it again; should not change the state
	e.Set()
	assert.True(t, e.IsSet())
}

func TestClear(t *testing.T) {
	e := events.New()

	// Set the event and then clear it
	e.Set()
	assert.True(t, e.IsSet())

	e.Clear()
	assert.False(t, e.IsSet())

	// Clear it again; should not change the state
	e.Clear()
	assert.False(t, e.IsSet())
}

func TestWait(t *testing.T) {
	e := events.New()

	// Test waiting when the event is not set
	go func() {
		time.Sleep(10 * time.Millisecond)
		e.Set()
	}()

	// Wait should return false after the event is set
	assert.False(t, e.Wait(1.0)) // 1 second timeout

	// Clear the event and wait again
	e.Clear()

	// Test waiting with timeout
	assert.True(t, e.Wait(0.01)) // 100 ms timeout, should timeout
}

func TestWaitWithTimeout(t *testing.T) {
	e := events.New()

	// Test immediate wait when the event is set
	e.Set()
	assert.False(t, e.Wait(1.0)) // Should return false immediately

	// Clear the event
	e.Clear()

	// Test timeout wait
	assert.True(t, e.Wait(0.01)) // Should timeout since the event is cleared
}
