// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"bytes"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// CommandHandler defines the function handling commands.
type CommandHandler func(string) string

// Process manages OS signal handling in addition to Tasklet management.
type Process struct {
	*TaskletHandler

	// command handling function and comm listener
	cmdHandler  CommandHandler
	cmdListener comm.Listener

	// Map of signal handlers.
	sigHandlers map[os.Signal]func()
}

// NewProcessHandler creates a new ProcessHandler with signal handlers
// for common signals like SIGINT and SIGTERM.
func NewProcessHandler(log *logging.Logger, tsk Tasklet) *Process {
	h := &Process{
		TaskletHandler: NewTaskletHandler(log, tsk),
	}
	h.sigHandlers = map[os.Signal]func(){
		syscall.SIGINT:  h.Stop, // Handle interruption signals (Ctrl+C).
		syscall.SIGTERM: h.Stop, // Handle termination signals.
		syscall.SIGKILL: h.Stop, // Handle kill signals.
		syscall.SIGQUIT: h.Stop, // Handle quit signals.
		syscall.SIGHUP:  h.Stop, // Handle hangup signals.
	}
	return h
}

// SetCmdHandler sets the command handling function and comm listener to
// enable command handling feature on process.
func (h *Process) SetCmdHandler(l comm.Listener, f CommandHandler) {
	if l != nil {
		l.SetConnHandler(h.handleConnection)
	}
	h.cmdListener = l
	h.cmdHandler = f
}

// SetSignalHandler allows the user to define custom handlers for specific signals.
func (h *Process) SetSignalHandler(sig os.Signal, fn func()) {
	if sig != nil && fn != nil {
		h.sigHandlers[sig] = fn
	}
}

// handleSignal processes incoming signals and triggers the corresponding handler.
func (h *Process) handleSignal(sig os.Signal) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			h.Log.Error("%s", r)
			h.Log.Trace("\n----------\n%s----------", stack[indx:])
		}
	}()

	// Log the received signal and execute the associated handler.
	h.Log.Debug("<received signal: %v>", sig)
	if handler, exists := h.sigHandlers[sig]; exists {
		handler()
	} else {
		h.Log.Warn("no handler registered for signal: %v", sig)
	}
}

// handleConnection handles command connections
func (h *Process) handleConnection(conn comm.Connection) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			h.Log.Error("%s", r)
			h.Log.Trace("\n----------\n%s----------", stack[indx:])
		}
	}()

	for !h.termEvent.IsSet() && conn.IsOpened() {
		b, addr, err := conn.RecvFrom(-1)
		if err != nil {
			if err == comm.ErrClosed {
				return
			}
			h.Log.Error(err.Error())
			continue
		}
		cmd := strings.TrimSpace(string(b))
		if cmd == "" {
			continue
		}
		reply := h.cmdHandler(cmd)
		if reply != "" {
			if err := conn.SendTo([]byte(reply+"\n"), addr, -1); err != nil {
				h.Log.Error(err.Error())
			}
		}
	}
}

// Start begins the process and sets up signal handling.
func (h *Process) Start() {
	// Create a buffered channel to receive multiple signals without blocking.
	sigCh := make(chan os.Signal, 2)
	for sig := range h.sigHandlers {
		// Register for signals defined in sigHandlers.
		signal.Notify(sigCh, sig)
	}

	// Start a goroutine to listen for OS signals and handle them.
	go func() {
		for sig := range sigCh {
			h.handleSignal(sig)
		}
	}()

	var waitGrp sync.WaitGroup

	if h.cmdListener != nil && h.cmdHandler != nil {
		waitGrp.Add(1)
		go func() {
			defer waitGrp.Done()
			if err := h.cmdListener.Start(); err != nil {
				panic(err)
			}
		}()
	}

	// Start the tasklet lifecycle.
	h.TaskletHandler.Enable()
	h.TaskletHandler.Start()

	waitGrp.Wait()
}

// Stop stop the process.
func (h *Process) Stop() {
	h.TaskletHandler.Disable()
	if h.cmdListener != nil {
		h.cmdListener.Stop()
	}
	h.TaskletHandler.Stop()
}
