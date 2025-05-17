// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"bytes"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/exonlabs/go-utils/pkg/events"
	"github.com/exonlabs/go-utils/pkg/logging"
)

var (
	ErrTskStarted = errors.New("tasklet already started")
)

// Tasklet defines the three life‑cycle hooks required for a managed job.
type Tasklet interface {
	Initialize() error // prepares resources
	Execute() error    // performs one unit of work
	Terminate() error  // releases resources
}

// TaskletHandler coordinates a Tasklet's life cycle, providing enable/disable
// flags, liveness tracking, and cooperative sleep/stop primitives.
type TaskletHandler struct {
	// application logger instance
	Log *logging.Logger

	// the tasklet instance to manage
	tasklet Tasklet

	// flag determines if routine should be enabled or not
	isEnabled atomic.Bool
	// flag to track current tasklet execution state
	isAlive atomic.Bool

	// starting sync mutex
	startMutex sync.Mutex
	// stopping sync mutex
	stopMutex sync.Mutex
	// stopEvent signals a stop operation.
	stopEvent events.Event
	// killEvent signals multiple stop operation to stop immediately.
	killEvent events.Event
	// termEvent signals a terminate operation.
	termEvent events.Event

	// delay in seconds to apply after errors in execution loop.
	ErrorDelay float64
}

// NewTaskletHandler returns a TaskletHandler that manages the supplied Tasklet
// and logs via the provided logger.
func NewTaskletHandler(log *logging.Logger, tsk Tasklet) *TaskletHandler {
	return &TaskletHandler{
		Log:        log,
		tasklet:    tsk,
		ErrorDelay: 1,
	}
}

// IsEnabled reports whether the tasklet is currently marked as enabled.
func (h *TaskletHandler) IsEnabled() bool {
	return h.isEnabled.Load()
}

// IsAlive reports whether the tasklet's execution loop is running.
func (h *TaskletHandler) IsAlive() bool {
	return h.isAlive.Load()
}

// Enable marks the tasklet as enabled, allowing Start to run or continue.
func (h *TaskletHandler) Enable() {
	h.isEnabled.Store(true)
}

// Disable marks the tasklet as disabled, causing Start to exit its loop.
func (h *TaskletHandler) Disable() {
	h.isEnabled.Store(false)
}

// runs the tasklet lifecycle, return error in case of error
// in the tasklet initialization.
func (h *TaskletHandler) run() error {
	// initialize the tasklet
	if err := h.tasklet.Initialize(); err != nil {
		return err
	}

	h.isAlive.Store(true)
	defer func() {
		h.isAlive.Store(false)
		// Panic recovery to handle unexpected errors during execution.
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			if h.Log != nil {
				h.Log.Panic("%v\n----------\n%s----------", r, stack[indx:])
			} else {
				fmt.Printf("%v\n----------\n%s----------\n", r, stack[indx:])
			}
			h.Sleep(h.ErrorDelay)
		}
	}()

	h.stopEvent.Clear()
	h.killEvent.Clear()
	h.termEvent.Clear()

	// Run tasklet execution loop until a stop event.
	for !h.stopEvent.IsSet() {
		if err := h.tasklet.Execute(); err != nil {
			if h.Log != nil {
				h.Log.Error(err.Error())
			} else {
				fmt.Printf("error: %s\n", err.Error())
			}
			h.Sleep(h.ErrorDelay)
		}
	}
	return nil
}

// Start runs the full life cycle while the tasklet is enabled and sets the
// liveness flag for the duration. start return error if already started.
func (h *TaskletHandler) Start() error {
	if !h.startMutex.TryLock() {
		return ErrTskStarted
	}
	defer h.startMutex.Unlock()

	for h.isEnabled.Load() {
		if err := h.run(); err != nil {
			return err
		}
	}
	return nil
}

// Stop requests graceful termination: it signals the execute loop to exit and
// invokes the Tasklet's Terminate method once.
func (h *TaskletHandler) Stop() error {
	if h.stopEvent.IsSet() {
		h.killEvent.Set()
	} else {
		h.stopEvent.Set()
	}

	if !h.stopMutex.TryLock() {
		return nil
	}
	defer h.stopMutex.Unlock()

	h.termEvent.Clear()
	defer h.termEvent.Set()

	if h.isAlive.Load() {
		return h.tasklet.Terminate()
	}
	return nil
}

// Sleep pauses execution for the given timeout duration in seconds. it returns
// true if timeout reached without stop function is called, false other wise.
// Setting timeout 0 or negative value will wait indefinitely untill stop
// function is called. if multiple stop is called then returns immediately.
func (h *TaskletHandler) Sleep(timeout float64) bool {
	if h.stopEvent.IsSet() {
		return h.killEvent.Wait(timeout)
	}
	return h.stopEvent.Wait(timeout)
}

// WaitTerm waits for tasklet terminate finish for timeout duration in seconds,
// returns true if tasklet terminate finishes before timeout is reached.
func (h *TaskletHandler) WaitTerm(timeout float64) bool {
	return !h.termEvent.Wait(timeout)
}
