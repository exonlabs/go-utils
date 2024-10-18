// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xevent

import (
	"sync/atomic"
	"time"
)

// An event manages an internal flag that can be set to true with the set()
// method and reset to false with the clear() method.
// The wait() method blocks until the flag is true.
// The flag is initially false.
type Event struct {
	state atomic.Bool
	setch chan bool
}

// Creates new [Event] instance.
func New() *Event {
	e := &Event{}
	e.state.Store(false)
	e.setch = make(chan bool, 1)
	return e
}

// Set the internal flag to true. All threads waiting for it to become
// true are awakened. Threads that call wait() once the flag is true
// will not block at all.
func (e *Event) Set() {
	e.state.Store(true)
	if len(e.setch) == 0 {
		e.setch <- true
	}
}

// Reset the internal flag to false. Subsequently, threads calling wait()
// will block until set() is called to set the internal flag to true again.
func (e *Event) Clear() {
	e.state.Store(false)
	if len(e.setch) > 0 {
		<-e.setch
	}
}

// Check event state weather its Set or Clear
func (e *Event) IsSet() bool {
	return e.state.Load()
}

// Wait blocks until timeout and returns true. if internal flag is set
// before timeout ends, wait returns immediately with false.
func (e *Event) Wait(timeout float64) bool {
	select {
	case <-time.After(time.Duration(timeout * 1000000000)):
		return true
	case <-e.setch:
	}
	return false
}
