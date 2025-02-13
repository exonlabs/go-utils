// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package jconfig

import "os"

type FileHandler interface {
	// IsExist checks whether the named file exists.
	IsExist(name string) bool
	// Read reads the named file and returns the contents.
	Read(name string) ([]byte, error)
	// Write writes data to the named file, creating it if necessary.
	Write(name string, data []byte, perm os.FileMode) error
	// Remove delete the named file from system.
	Remove(name string) error
}

// StdFileHandler represents the standard local file access using os pkg.
type StdFileHandler struct{}

// NewStdFileHandler creates new standard local file handler.
func NewStdFileHandler() *StdFileHandler {
	return &StdFileHandler{}
}

// IsExist checks whether the named file exists.
func (h *StdFileHandler) IsExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// Read reads the named file and returns the contents.
func (h *StdFileHandler) Read(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Write writes data to the named file, creating it if necessary.
func (h *StdFileHandler) Write(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Remove delete the named file from system.
func (h *StdFileHandler) Remove(name string) error {
	return os.Remove(name)
}
