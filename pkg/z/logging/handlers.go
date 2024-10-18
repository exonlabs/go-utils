// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xlog

import "os"

type Handler interface {
	HandleRecord(string) error
}

// Handler writing log messages to Stdout
type StdoutHandler struct{}

// Creates new [StdoutHandler] for logger.
func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{}
}

// Handles the log record message and writes the log to stdout.
func (h *StdoutHandler) HandleRecord(r string) error {
	_, err := os.Stdout.WriteString(r + "\n")
	if err == nil {
		err = os.Stdout.Sync()
	}
	return err
}

// Handler writing log messages to file on system.
type FileHandler struct {
	FilePath string // path to file on system
}

// Creates new [FileHandler] for logger.
func NewFileHandler(path string) *FileHandler {
	return &FileHandler{
		FilePath: path,
	}
}

// Handles the log record message and writes the log to file.
func (h *FileHandler) HandleRecord(r string) error {
	f, err := os.OpenFile(
		h.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o664)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(r + "\n")
	if err == nil {
		err = f.Sync()
	}
	return err
}
