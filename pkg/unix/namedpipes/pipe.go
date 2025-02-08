// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package namedpipes

import (
	"os"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/events"
)

// NamedPipe represents a named pipe and provides methods for reading,
// writing, and managing the pipe.
type NamedPipe struct {
	// Context containing common attributes and functions.
	*Context

	// fd is the OS file descriptor instance for the named pipe.
	fd *os.File

	// breakEvent signals an interrupt in operations.
	breakEvent *events.Event
}

// New creates a new NamedPipe instance with options.
func New(path string, opts dictx.Dict) *NamedPipe {
	return &NamedPipe{
		Context:    NewContext(path, opts),
		breakEvent: events.New(),
	}
}

// Cancel triggers the pipe's BreakEvent, cancelling any waiting operations.
func (p *NamedPipe) Cancel() {
	p.breakEvent.Set()
}
