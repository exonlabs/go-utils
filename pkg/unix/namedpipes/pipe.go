// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package namedpipes

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/events"
)

// NamedPipe represents a named pipe and provides methods for reading,
// writing, and managing the pipe.
type NamedPipe struct {
	// path defines the file system path of the named pipe.
	path string
	// fd is the OS file descriptor instance for the named pipe.
	fd *os.File

	// breakEvent signals an interrupt in operations.
	breakEvent *events.Event

	// mu defines mutex for operations.
	mu sync.Mutex

	// PollTimeout defines the timeout in seconds for read data polling.
	PollTimeout float64
	// PollChunkSize defines the size of chunks to read during polling.
	PollChunkSize int
	// PollMaxSize defines the maximum size for read polling data.
	// use 0 or negative value to disable max limit for read data polling.
	PollMaxSize int
}

// New creates a new NamedPipe instance with options.
//
// The parsed options are:
//   - poll_timeout: (float64) the timeout in seconds for read data polling.
//     polling timeout value must be > 0.
//   - poll_chunksize: (int) the size of chunks to read during polling.
//     polling chunk size value must be > 0.
//   - poll_maxsize: (int) the maximum size for read data.
//     use 0 or negative value to disable max limit for read polling.
func New(path string, opts dictx.Dict) *NamedPipe {
	p := &NamedPipe{
		path:          filepath.Clean(path),
		breakEvent:    events.New(),
		PollTimeout:   0.1,
		PollChunkSize: 102400,
		PollMaxSize:   -1,
	}

	// Apply custom options.
	if v := dictx.GetFloat(opts, "poll_timeout", 0); v > 0 {
		p.PollTimeout = v
	}
	if v := dictx.GetInt(opts, "poll_chunksize", 0); v > 0 {
		p.PollChunkSize = v
	}
	if v := dictx.GetInt(opts, "poll_maxsize", 0); v > 0 {
		p.PollMaxSize = v
	}

	return p
}

// Path returns the file system path of the named pipe.
func (c *NamedPipe) Path() string {
	return c.path
}

// Cancel triggers the pipe's BreakEvent, cancelling any waiting operations.
func (p *NamedPipe) Cancel() {
	p.breakEvent.Set()
}
