// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	// ErrError indicates the parent error.
	ErrError = errors.New("")

	// ErrUri indicates an invalid URI error.
	ErrUri = fmt.Errorf("%winvalid uri", ErrError)
	// ErrConnection indicates a connection failure.
	ErrConnection = fmt.Errorf("%wconnection failed", ErrError)
	// ErrClosed indicates that the connection is closed.
	ErrClosed = fmt.Errorf("%wconnection closed", ErrError)
	// ErrBreak indicates an operation interruption.
	ErrBreak = fmt.Errorf("%woperation break", ErrError)
	// ErrTimeout indicates that the operation timed out.
	ErrTimeout = fmt.Errorf("%woperation timeout", ErrError)
	// ErrRead indicates a read failure.
	ErrRead = fmt.Errorf("%wread failed", ErrError)
	// ErrWrite indicates a write failure.
	ErrWrite = fmt.Errorf("%wwrite failed", ErrError)
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
		strings.Contains(err.Error(), "has been closed"),
		strings.Contains(err.Error(), "input/output error"):
		return true
	default:
		return false
	}
}
