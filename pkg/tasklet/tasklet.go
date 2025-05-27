// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package tasklet

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
	ErrStarted   = errors.New("tasklet already started")
	ErrIsDead    = errors.New("tasklet is dead")
	ErrIsRunning = errors.New("tasklet is running")
)

// A tasklet manages a Job through its complete lifecycle, including
// initialization, execution, and termination phases. It provides methods to
// control the execution flow and check the tasklet's current state.
type Tasklet interface {
	// IsEnabled reports whether the tasklet is currently marked as enabled.
	// When enabled, the tasklet can be started and will continue its execution.
	IsEnabled() bool

	// IsActive reports whether the tasklet's is alive and running or not.
	// Returns true if either the execution or termination phases are running.
	IsActive() bool

	// Enable marks the tasklet as enabled.
	// This is used to permit the tasklet to begin or resume execution.
	Enable()

	// Disable marks the tasklet as disabled.
	// This is used to stop the tasklet after completing its current cycle.
	Disable()

	// Start begins the tasklet full lifecycle while it is enabled.
	// It will initialize the job, execute it in a loop, and terminate it when
	// disabled or stopped. Returns an error if the tasklet is already started
	// or if initialization fails.
	Start() error

	// Stop requests graceful termination of the tasklet.
	// It signals the execute loop to exit and invokes the Terminate method.
	// Returns an error if termination cannot be initiated.
	Stop() error

	// Join waits for the tasklet's terminate method to finish for the specified
	// timeout duration in seconds. Returns true if termination completes before
	// the timeout is reached, false otherwise.
	Join(float64) bool
}

// Job defines the three life‑cycle hooks required for a managed job.
// These methods represent the complete lifecycle of a job managed by a Tasklet.
type Job interface {
	// Initialize prepares resources needed for job execution.
	// This is called once before the execution loop begins.
	// Returns an error if initialization fails, preventing the job from starting.
	Initialize() error

	// Execute performs one unit of work in the job's execution cycle.
	// This is called repeatedly in a loop while the tasklet is enabled.
	// Returns an error if execution fails, which will be logged and execution
	// will continue after a delay.
	Execute() error

	// Terminate releases resources acquired during initialization and execution.
	// This is called once after the execution loop ends.
	// Returns an error if termination encounters issues, which will be logged.
	Terminate() error
}

// baseTasklet defines a base tasklet implementation that provides common
// functionality for both synchronous and asynchronous tasklet variants.
// It manages the execution lifecycle and state of a job.
type baseTasklet struct {
	// application logger instance.
	Log *logging.Logger

	// the job instance to manage.
	job Job

	// flag determines if tasklet should be enabled or not.
	isEnabled atomic.Bool

	// run sync mutex
	runMutex sync.Mutex
	// term sync mutex
	termMutex sync.Mutex

	// stopEvent signals a stop operation.
	stopEvent events.Event
	// killEvent signals multiple stop operation to stop immediately.
	killEvent events.Event
	// termEvent signals a terminate operation.
	termEvent events.Event

	// delay in seconds to apply after errors in execution loop.
	ErrorDelay float64
}

// new_baseTasklet returns a new baseTasklet instance configured with the
// provided logger and job. This is an internal constructor used by the
// AsyncTasklet and SyncTasklet implementations.
func new_baseTasklet(log *logging.Logger, job Job) *baseTasklet {
	return &baseTasklet{
		Log:        log,
		job:        job,
		ErrorDelay: 1,
	}
}

// IsEnabled reports whether the tasklet is currently marked as enabled.
func (t *baseTasklet) IsEnabled() bool {
	return t.isEnabled.Load()
}

// IsActive reports whether the tasklet is currently active (running or terminating).
// It checks if either the execution or termination phases are running indicating
// that the tasklet is active.
func (t *baseTasklet) IsActive() bool {
	if !t.runMutex.TryLock() {
		return true
	}
	defer t.runMutex.Unlock()

	if !t.termMutex.TryLock() {
		return true
	}
	defer t.termMutex.Unlock()

	return false
}

// Enable marks the tasklet as enabled, allowing Start to run or continue.
// This permits the tasklet to begin execution if Start is called, or to
// continue its execution loop if it's already running.
func (t *baseTasklet) Enable() {
	t.isEnabled.Store(true)
}

// Disable marks the tasklet as disabled, causing Start to exit its loop.
// This signals the tasklet to gracefully stop execution after completing
// its current cycle. The tasklet will still complete its current execution
// and termination phases.
func (t *baseTasklet) Disable() {
	t.isEnabled.Store(false)
}

// Sleep pauses execution for the given timeout duration in seconds.
// It returns true if the timeout is reached without a stop signal being received,
// and false if execution should stop (because Stop was called).
//
// If a stop signal was already received, Sleep waits for a kill signal
// (multiple Stop calls) instead.
//
// Setting timeout to 0 or a negative value will cause Sleep to wait indefinitely
// until a stop or kill signal is received.
func (t *baseTasklet) Sleep(timeout float64) bool {
	if t.stopEvent.IsSet() {
		return t.killEvent.Wait(timeout)
	}
	return t.stopEvent.Wait(timeout)
}

// stop_event signals a stop event to the execution of the tasklet.
// If this is the first call, it sets the stopEvent.
// If stopEvent is already set (meaning Stop was already called once),
// it sets the killEvent to indicate an urgent kill event request.
func (t *baseTasklet) stop_event() {
	if t.stopEvent.IsSet() {
		t.killEvent.Set()
	} else {
		t.stopEvent.Set()
	}
}

// Join waits for the tasklet's terminate method to finish for the specified
// timeout duration in seconds.
// Returns true if the termination completes before the timeout is reached,
// false otherwise.
//
// This can be used to ensure the tasklet has fully terminated before proceeding
// with other operations.
func (t *baseTasklet) Join(timeout float64) bool {
	return !t.termEvent.Wait(timeout)
}

// log_recover logs or prints panic error with stack
func (t *baseTasklet) log_recover(r any) {
	stack := debug.Stack()
	indx := bytes.Index(stack, []byte("panic({"))
	if t.Log != nil {
		t.Log.Panic("%v\n----------\n%s----------", r, stack[indx:])
	} else {
		fmt.Printf("%v\n----------\n%s----------\n", r, stack[indx:])
	}
}

// run_initialize executes the initialize phase of the tasklet lifecycle,
// and handles any panics that occur during execution.
func (t *baseTasklet) run_initialize() (err error) {
	defer func() {
		if r := recover(); r != nil {
			t.log_recover(r)
			err = fmt.Errorf("%v", r)
		}
	}()

	return t.job.Initialize()
}

// run_execute executes the main execution phase of the tasklet lifecycle.
// It clears all event flags, runs the job's Execute method in a loop until
// a stop event is received, and handles any panics that occur during execution.
// Execution errors are logged and followed by a delay specified by ErrorDelay.
func (t *baseTasklet) run_execute() {
	t.stopEvent.Clear()
	t.killEvent.Clear()
	t.termEvent.Clear()

	defer func() {
		if r := recover(); r != nil {
			t.log_recover(r)
			t.Sleep(t.ErrorDelay)
		}
	}()

	// Run tasklet execution loop until a stop event.
	for !t.stopEvent.IsSet() && !t.killEvent.IsSet() {
		if err := t.job.Execute(); err != nil {
			if t.Log != nil {
				t.Log.Error(err.Error())
			} else {
				fmt.Printf("error: %s\n", err.Error())
			}
			t.Sleep(t.ErrorDelay)
		}
	}
}

// run_terminate handles the termination phase of the tasklet lifecycle.
// It calls the job's Terminate method to release resources, handles any
// panics that occur during termination, and sets the termEvent when complete
// to signal that termination has finished.
func (t *baseTasklet) run_terminate() {
	t.termEvent.Clear()

	defer func() {
		t.termEvent.Set()
		if r := recover(); r != nil {
			t.log_recover(r)
		}
	}()

	if err := t.job.Terminate(); err != nil {
		if t.Log != nil {
			t.Log.Error(err.Error())
		} else {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
}

// AsyncTasklet defines an asynchronous tasklet's lifecycle, providing enable/disable
// flags, liveness tracking, and cooperative sleep/stop primitives.
//
// The key characteristic of AsyncTasklet is that it runs the termination sequence
// in the Stop function, which allows the termination sequence to be executed in
// a parallel routine to the main execution. This means Stop() will block until
// termination is complete, but Start() can return while termination is still in progress.
type AsyncTasklet struct {
	*baseTasklet
}

// NewAsyncTasklet returns a new asynchronous tasklet that manages the supplied job.
func NewAsyncTasklet(log *logging.Logger, job Job) *AsyncTasklet {
	return &AsyncTasklet{
		baseTasklet: new_baseTasklet(log, job),
	}
}

// Start runs the full lifecycle of the AsyncTasklet while it is enabled.
// It initializes the job and executes it in a loop until the tasklet is disabled or stopped.
//
// The AsyncTasklet does not run the termination sequence as part of Start;
// termination occurs separately in the Stop method.
//
// Returns ErrStarted if the tasklet is already running, or any error returned
// by the job's Initialize method.
func (t *AsyncTasklet) Start() error {
	if !t.runMutex.TryLock() {
		return ErrStarted
	}
	defer t.runMutex.Unlock()

	for t.isEnabled.Load() {
		if err := t.run_initialize(); err != nil {
			return err
		}
		t.run_execute()
	}
	return nil
}

// Stop requests graceful termination of the AsyncTasklet.
// It signals the execute loop to exit and then invokes the job's Terminate method.
//
// This method blocks until termination is complete.
// If termination is already in progress this method returns immediately without error.
//
// Returns nil on success or if termination is already in progress.
func (t *AsyncTasklet) Stop() error {
	t.stop_event()

	if !t.termMutex.TryLock() {
		return nil
	}
	defer t.termMutex.Unlock()

	t.run_terminate()
	return nil
}

// SyncTasklet defines a synchronous tasklet's lifecycle, providing enable/disable
// flags, liveness tracking, and cooperative sleep/stop primitives.
//
// The key characteristic of SyncTasklet is that it runs the termination sequence
// as part of the Start method, immediately after the execution phase completes.
// This ensures that initialization, execution, and termination all occur in sequence
// within the same goroutine, providing a more predictable lifecycle.
type SyncTasklet struct {
	*baseTasklet
}

// NewSyncTasklet returns a new synchronous tasklet that manages the supplied job.
func NewSyncTasklet(log *logging.Logger, job Job) *SyncTasklet {
	return &SyncTasklet{
		baseTasklet: new_baseTasklet(log, job),
	}
}

// Start runs the full lifecycle of the SyncTasklet while it is enabled.
// It initializes the job, executes it in a loop until the tasklet is disabled
// or stopped, and then runs the termination sequence.
//
// SyncTasklet handles the complete lifecycle (including termination)
// within the Start method, ensuring a predictable execution sequence.
//
// Returns ErrStarted if the tasklet is already running, or any error returned
// by the job's Initialize method.
func (t *SyncTasklet) Start() error {
	if !t.runMutex.TryLock() {
		return ErrStarted
	}
	defer t.runMutex.Unlock()

	for t.isEnabled.Load() {
		if err := t.run_initialize(); err != nil {
			return err
		}
		t.run_execute()
		t.run_terminate()
	}
	return nil
}

// Stop requests graceful termination of the SyncTasklet.
// It simply signals the execute loop to exit by setting the stop event.
//
// The SyncTasklet's Stop method does not directly invoke
// the termination sequence - termination occurs as part of the Start method
// after the execution phase completes. This means Stop returns immediately
// without blocking for termination to complete.
//
// Returns nil on success.
func (t *SyncTasklet) Stop() error {
	t.stop_event()
	return nil
}
