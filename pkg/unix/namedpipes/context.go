// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package namedpipes

import (
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
)

const (
	// POLL_TIMEOUT defines the default timeout for polling in seconds.
	POLL_TIMEOUT = 0.1
	// POLL_CHUNKSIZE is the default size of chunks to read during polling.
	POLL_CHUNKSIZE = 4096
	// POLL_MAXSIZE is the default maximum size for polling data.
	POLL_MAXSIZE = 0
)

// Context represents the configuration and state for communication handling.
type Context struct {
	// path defines the file system path of the named pipe.
	path string

	// Options defines the optional settings.
	Options dictx.Dict

	// PollTimeout defines the timeout in seconds for read data polling.
	PollTimeout float64
	// PollChunkSize defines the size of chunks to read during polling.
	PollChunkSize int
	// PollMaxSize defines the maximum size for read polling data.
	// use 0 or negative value to disable max limit for read data polling.
	PollMaxSize int
}

// NewContext creates and initializes a new Context instance with optional settings.
// The parsed options are:
//   - poll_timeout: (float64) the timeout in seconds for read data polling.
//   - poll_chunksize: (int) the size of chunks to read during polling.
//   - poll_maxsize: (int) the maximum size for read polling data.
//     use 0 or negative value to disable max limit for read data polling.
func NewContext(path string, opts dictx.Dict) *Context {
	ctx := &Context{
		path:          filepath.Clean(path),
		Options:       opts,
		PollTimeout:   POLL_TIMEOUT,
		PollChunkSize: POLL_CHUNKSIZE,
		PollMaxSize:   POLL_MAXSIZE,
	}

	// Apply custom options.
	if opts != nil {
		if v := dictx.GetFloat(opts, "poll_timeout", 0); v > 0 {
			ctx.PollTimeout = v
		}
		if v := dictx.GetInt(opts, "poll_chunksize", 0); v > 0 {
			ctx.PollChunkSize = v
		}
		if v := dictx.GetInt(opts, "poll_maxsize", 0); v >= 0 {
			ctx.PollMaxSize = v
		}
	}

	return ctx
}

// Path returns the file system path of the named pipe.
func (c *Context) Path() string {
	return c.path
}
