// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"os"
	"sync"
)

// Handler interface for processing log messages.
type Handler interface {
	// HandleMessage process a log message.
	HandleMessage(msg string) error
}

// StdoutHandler writes log messages to standard output.
type StdoutHandler struct {
	mu sync.Mutex
}

// NewStdoutHandler creates a new instance of StdoutHandler.
func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{}
}

// HandleMessage writes a log message to standard output.
func (h *StdoutHandler) HandleMessage(msg string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := os.Stdout.Write([]byte(msg + "\n"))
	return err
}

// FileHandler writes log messages to a specified file.
type FileHandler struct {
	FilePath string // Path to the log file
	mu       sync.Mutex
}

// NewFileHandler creates a new FileHandler for the specified path.
func NewFileHandler(path string) *FileHandler {
	return &FileHandler{
		FilePath: path,
	}
}

// HandleMessage writes the log message to the specified file.
func (h *FileHandler) HandleMessage(msg string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	fh, err := os.OpenFile(
		h.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = fh.Write([]byte(msg + "\n"))
	if err == nil {
		// Ensure the output is flushed
		err = fh.Sync()
	}
	return err
}
