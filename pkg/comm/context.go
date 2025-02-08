// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package comm

import (
	"strings"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/logging"
)

const (
	// POLL_TIMEOUT defines the default timeout for read polling in seconds.
	POLL_TIMEOUT = 0.01
	// POLL_CHUNKSIZE defines the default size of chunks to read during polling.
	POLL_CHUNKSIZE = 4096
	// POLL_MAXSIZE is the default maximum size for read polling data.
	POLL_MAXSIZE = 0
)

// Context represents the configuration and state for communication handling.
type Context struct {
	// uri specifies the resource identifier.
	uri string

	// CommLog is the logger instance for communication data logging.
	CommLog *logging.Logger

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
func NewContext(uri string, log *logging.Logger, opts dictx.Dict) *Context {
	ctx := &Context{
		uri:           strings.TrimSpace(uri),
		CommLog:       log,
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

// Uri returns the connection Uri.
func (c *Context) Uri() string {
	return c.uri
}

// Type returns the type of the connection as inferred from the Uri.
// (ex.: tcp, tcp4, udp, sock, serial)
func (c *Context) Type() string {
	return strings.ToLower(strings.SplitN(c.uri, "@", 2)[0])
}

// LogMsg logs messages using the communication logger.
func (c *Context) LogMsg(msg string, args ...any) {
	if c.CommLog != nil && msg != "" {
		c.CommLog.Info(msg, args...)
	}
}

// LogTx logs transmitted data in a formatted hexadecimal string.
//
//	2006-01-02 15:04:05.000000 TX >> 0102030405060708090A0B0C0D0E0F
func (c *Context) LogTx(data []byte, addr any) {
	if c.CommLog != nil && len(data) > 0 {
		if addr != nil {
			c.CommLog.Info("(%s) TX >> %X", addr, data)
		} else {
			c.CommLog.Info("TX >> %X", data)
		}
	}
}

// LogRx logs received data in a formatted hexadecimal string.
//
//	2006-01-02 15:04:05.000000 RX << 0102030405060708090A0B0C0D0E0F
func (c *Context) LogRx(data []byte, addr any) {
	if c.CommLog != nil && len(data) > 0 {
		if addr != nil {
			c.CommLog.Info("(%s) RX << %X", addr, data)
		} else {
			c.CommLog.Info("RX << %X", data)
		}
	}
}
