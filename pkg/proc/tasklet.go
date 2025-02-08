// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"bytes"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/exonlabs/go-utils/pkg/events"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// Tasklet defines the interface for tasklets.
type Tasklet interface {
	Initialize() error
	Execute() error
	Terminate() error
}

// TaskletHandler manages a Tasklet's lifecycle.
type TaskletHandler struct {
	// Log is the logger instance for application logging.
	Log *logging.Logger

	// The tasklet instance to manage
	tasklet Tasklet

	// flag determines if routine should be enabled or not
	isEnabled atomic.Bool
	// flag to track current tasklet execution state
	isAlive atomic.Bool
	// flag to track current tasklet initialization state
	isInitialized atomic.Bool

	// TermEvent signals a termination operation.
	TermEvent *events.Event
	// KillEvent signals a forceful termination operation.
	KillEvent *events.Event
}

// NewTaskletHandler creates a new tasklet handler.
func NewTaskletHandler(log *logging.Logger, tsk Tasklet) *TaskletHandler {
	return &TaskletHandler{
		Log:       log,
		tasklet:   tsk,
		TermEvent: events.New(),
		KillEvent: events.New(),
	}
}

// IsEnabled returns whether the tasklet is currently enabled.
func (h *TaskletHandler) IsEnabled() bool {
	return h.isEnabled.Load()
}

// IsAlive returns whether the tasklet is currently active and running.
func (h *TaskletHandler) IsAlive() bool {
	return h.isAlive.Load()
}

// IsInitialized returns whether the tasklet is currently initialized.
func (h *TaskletHandler) IsInitialized() bool {
	return h.isInitialized.Load()
}

// Enable sets the tasklet as enabled
func (h *TaskletHandler) Enable() {
	h.isEnabled.Store(true)
}

// Disable sets the tasklet as disabled
func (h *TaskletHandler) Disable() {
	h.isEnabled.Store(false)
}

// Run initiates the tasklet lifecycle, handling initialization,
// execution, and termination.
func (h *TaskletHandler) Run() {
	// set default logger
	if h.Log == nil {
		h.Log = logging.NewStdoutLogger("tasklet")
		h.Log.SetFormatter(logging.BasicFormatter)
	}

	defer func() {
		// Panic recovery to handle unexpected errors during execution.
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			h.Log.Error("%s", r)
			h.Log.Trace("\n----------\n%s----------", stack[indx:])
		}
		// Ensure termination execute if initialized and not killed.
		if h.isInitialized.Load() && !h.KillEvent.IsSet() {
			if err := h.tasklet.Terminate(); err != nil {
				h.Log.Error("termination failed: %s", err.Error())
			}
		}
	}()

	h.TermEvent.Clear()
	h.KillEvent.Clear()

	// Attempt to initialize the tasklet.
	if err := h.tasklet.Initialize(); err != nil {
		h.Log.Error("initialization failed: %s", err.Error())
		return
	}
	h.isInitialized.Store(true)

	// Run tasklet execution loop until a termination event is set.
	for !h.TermEvent.IsSet() {
		if err := h.tasklet.Execute(); err != nil {
			h.Log.Error("execution error: %s", err.Error())
		}
	}
}

// Start initiates the tasklet lifecycle, handling initialization,
// execution, and termination.
func (h *TaskletHandler) Start() {
	h.isAlive.Store(true)
	defer h.isAlive.Store(false)

	for h.isEnabled.Load() {
		h.Run()
	}
}

// Stop gracefully stops the tasklet by setting the termination event.
func (h *TaskletHandler) Stop() {
	// If already stopping, forcefully kill.
	if h.TermEvent.IsSet() {
		h.KillEvent.Set()
	} else {
		h.TermEvent.Set()
	}
}

// Kill terminates the tasklet by setting both kill and termination events.
func (h *TaskletHandler) Kill() {
	h.KillEvent.Set()
	h.TermEvent.Set()
}

// Sleep pauses execution for the given timeout duration (in seconds),
// and waits for either a termination or kill event.
func (h *TaskletHandler) Sleep(timeout float64) bool {
	// Wait for kill event if termination is already set.
	if h.TermEvent.IsSet() {
		return h.KillEvent.Wait(timeout)
	}
	return h.TermEvent.Wait(timeout)
}

// WaitStop waits for tasklet to stop for the given timeout duration (in seconds),
// returns true if tasklet is stopped before timeout is reached.
func (h *TaskletHandler) WaitStop(timeout float64) bool {
	var tBreak time.Time
	if timeout > 0 {
		tBreak = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	}
	for h.Sleep(0.05) {
		if !h.isAlive.Load() {
			return true
		}
		if timeout > 0 && time.Now().After(tBreak) {
			return false
		}
	}
	return false
}
