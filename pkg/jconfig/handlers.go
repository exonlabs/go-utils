// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package jconfig

import "os"

type FileHandler interface {
	// Read reads the named file and returns the contents.
	Read(name string) ([]byte, error)
	// Write writes data to the named file, creating it if necessary.
	Write(name string, data []byte, perm os.FileMode) error
}

// StdFileHandler represents the standard local file access using os pkg.
type StdFileHandler struct{}

// NewStdFileHandler creates new standard local file handler.
func NewStdFileHandler() *StdFileHandler {
	return &StdFileHandler{}
}

// Read reads the named file and returns the contents.
func (h *StdFileHandler) Read(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Write writes data to the named file, creating it if necessary.
func (h *StdFileHandler) Write(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
