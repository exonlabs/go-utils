// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package namedpipes

import (
	"errors"
)

var (
	// ErrOpen indicates a connection open failure.
	ErrOpen = errors.New("open failed")

	// ErrRead indicates a read failure.
	ErrRead = errors.New("read failed")

	// ErrWrite indicates a write failure.
	ErrWrite = errors.New("write failed")

	// ErrBreak indicates an operation interruption.
	ErrBreak = errors.New("operation break")

	// ErrTimeout indicates that the operation timed out.
	ErrTimeout = errors.New("operation timeout")
)
