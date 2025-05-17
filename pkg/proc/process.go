// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/logging"
)

type SignalHandler func() error

// Process manages OS‑signal handling and delegates lifecycle control to an
// embedded TaskletHandler.
type Process struct {
	*TaskletHandler

	// Map of signal handlers.
	sigHandlers map[os.Signal]SignalHandler
}

// NewProcessHandler returns a Process that wraps the provided TaskletHandler
// and installs default handlers for SIGINT, SIGTERM, SIGKILL, SIGQUIT and
// SIGHUP.
func NewProcessHandler(log *logging.Logger, tsk Tasklet) *Process {
	p := &Process{
		TaskletHandler: NewTaskletHandler(log, tsk),
	}
	p.sigHandlers = map[os.Signal]SignalHandler{
		syscall.SIGINT:  p.Stop, // Handle interruption signals (Ctrl+C).
		syscall.SIGTERM: p.Stop, // Handle termination signals.
		syscall.SIGKILL: p.Stop, // Handle kill signals.
		syscall.SIGQUIT: p.Stop, // Handle quit signals.
		syscall.SIGHUP:  p.Stop, // Handle hangup signals.
	}
	return p
}

// SetSignalHandler registers a custom callback for the given signal,
// overwriting any previously‑registered handler.
func (p *Process) SetSignalHandler(sig os.Signal, fn SignalHandler) {
	if sig != nil && fn != nil {
		p.sigHandlers[sig] = fn
	}
}

// handleSignal processes incoming signals and triggers the corresponding handler.
func (p *Process) handleSignal(sig os.Signal) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			if p.Log != nil {
				p.Log.Panic("%v\n----------\n%s----------", r, stack[indx:])
			} else {
				fmt.Printf("%v\n----------\n%s----------\n", r, stack[indx:])
			}
		}
	}()

	if p.Log != nil {
		p.Log.Info("<received signal: %s>", sig)
	} else {
		fmt.Printf("<received signal: %s>\n", sig)
	}

	if handler, exists := p.sigHandlers[sig]; exists {
		if err := handler(); err != nil {
			if p.Log != nil {
				p.Log.Error(err.Error())
			} else {
				fmt.Println(err.Error())
			}
		}
	} else if p.Log != nil {
		p.Log.Trace("no handler registered for signal: %s", sig)
	}
}

// Start enables the underlying tasklet, launches its execution loop, and
// begins listening for OS signals defined in sigHandlers.
func (p *Process) Start() error {
	// buffered channel to receive multiple signals without blocking.
	sigCh := make(chan os.Signal, 2)
	// Register signals defined in sigHandlers.
	for sig := range p.sigHandlers {
		signal.Notify(sigCh, sig)
	}

	// Start a goroutine to listen for OS signals and handle them.
	go func() {
		for sig := range sigCh {
			go p.handleSignal(sig)
		}
	}()

	// Start the tasklet lifecycle.
	p.TaskletHandler.Enable()
	return p.TaskletHandler.Start()
}

// Stop disables the underlying tasklet and blocks until its execution loop
// has terminated.
func (p *Process) Stop() error {
	p.TaskletHandler.Disable()
	return p.TaskletHandler.Stop()
}
