// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"os"
)

// Handler interface for processing log records.
type Handler interface {
	HandleRecord(string) error
}

// StdoutHandler writes log messages to standard output.
type StdoutHandler struct{}

// NewStdoutHandler creates a new instance of StdoutHandler.
func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{}
}

// HandleRecord writes the log record to standard output.
func (h *StdoutHandler) HandleRecord(record string) error {
	_, err := os.Stdout.Write([]byte(record + "\n"))
	return err
}

// FileHandler writes log messages to a specified file.
type FileHandler struct {
	FilePath string // Path to the log file
}

// NewFileHandler creates a new instance of FileHandler for the specified path.
func NewFileHandler(path string) *FileHandler {
	return &FileHandler{
		FilePath: path,
	}
}

// HandleRecord writes the log record to the specified file.
func (h *FileHandler) HandleRecord(record string) error {
	f, err := os.OpenFile(h.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(record + "\n"))
	if err == nil {
		// Ensure the output is flushed
		err = f.Sync()
	}
	return err
}
