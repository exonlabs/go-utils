// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

import (
	"errors"
	"io"
	"strings"
)

var (
	// ErrUri indicates an invalid URI error.
	ErrUri = errors.New("invalid uri")

	// ErrConnection indicates a connection failure.
	ErrConnection = errors.New("connection failed")

	// ErrClosed indicates that the connection is closed.
	ErrClosed = errors.New("connection closed")

	// ErrBreak indicates an operation interruption.
	ErrBreak = errors.New("operation break")

	// ErrTimeout indicates that the operation timed out.
	ErrTimeout = errors.New("operation timeout")

	// ErrRead indicates a read failure.
	ErrRead = errors.New("read failed")

	// ErrWrite indicates a write failure.
	ErrWrite = errors.New("write failed")
)

// IsClosedError checks if the error is related to a closed connection.
// It recognizes common network errors like EOF, broken pipe, and reset by peer errors.
func IsClosedError(err error) bool {
	switch {
	case errors.Is(err, io.EOF):
		return true
	case strings.Contains(err.Error(), "closed network"),
		strings.Contains(err.Error(), "broken pipe"),
		strings.Contains(err.Error(), "reset by peer"),
		strings.Contains(err.Error(), "bad file descriptor"),
		strings.Contains(err.Error(), "input/output error"):
		return true
	default:
		return false
	}
}
