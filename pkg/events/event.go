// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package events

import (
	"sync"
	"sync/atomic"
	"time"
)

// An event manages an internal flag that can be set to true with the set()
// method and reset to false with the clear() method.
// The wait() method blocks until the flag is true.
// The flag is initially false.
type Event struct {
	state   atomic.Bool
	opMutex sync.Mutex
	waitCh  chan struct{}
}

// Creates new [Event] instance.
func New() *Event {
	return &Event{
		waitCh: make(chan struct{}),
	}
}

// Set the internal flag to true. All threads waiting for it to become
// true are awakened. Threads that call wait() once the flag is true
// will not block at all.
func (e *Event) Set() {
	e.opMutex.Lock()
	defer e.opMutex.Unlock()

	// If already set, do nothing.
	if e.state.Load() {
		return
	}

	e.state.Store(true)
	// Close channel to wake up all waiters
	if e.waitCh != nil {
		close(e.waitCh)
	}
}

// Reset the internal flag to false. Subsequently, threads calling wait()
// will block until set() is called to set the internal flag to true again.
func (e *Event) Clear() {
	e.opMutex.Lock()
	defer e.opMutex.Unlock()

	// If already cleared, do nothing.
	if !e.state.Load() {
		return
	}

	e.state.Store(false)
	// Create a new channel for future waiters
	e.waitCh = make(chan struct{})
}

// Check event state whether it's Set or Clear.
func (e *Event) IsSet() bool {
	return e.state.Load()
}

// Wait blocks until timeout and returns true if the internal flag is not set
// before the timeout. If the internal flag is set before the timeout ends,
// wait returns immediately with false.
//
// Setting timeout 0 or negative value will wait indefinitely untill
// the internal flag is set.
func (e *Event) Wait(timeout float64) bool {
	if e.state.Load() {
		return false
	}

	// lazy chan create at first use if not created
	if e.waitCh == nil {
		e.waitCh = make(chan struct{})
	}

	if timeout > 0 {
		timer := time.NewTimer(time.Duration(timeout * float64(time.Second)))
		select {
		case <-timer.C:
			return true // wait duration was reached.
		case <-e.waitCh:
			timer.Stop()
			return false // woken up because event was set.
		}
	} else {
		<-e.waitCh
		return false // woken up because event was set.
	}
}
