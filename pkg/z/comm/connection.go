package xcomm

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
)

// baseConnection is a base structure for connection handling
type baseConnection struct {
	uri        string
	uriLogging bool
	commLogger *Logger

	// operation events
	evtBreak  *xevent.Event
	evtKill   *xevent.Event
	ctxCancel context.CancelFunc
	// operation sync mutex
	op_mux sync.Mutex

	// error delay for execution loop
	ErrorDelay float64

	// read/write polling params
	PollInterval     float64
	PollChunkSize    int
	PollMaxSize      int
	PollWaitNullRead bool
}

func new_base_connection(
	uri string, commlog *Logger, opts Options) *baseConnection {
	conn := &baseConnection{
		uri:              strings.TrimSpace(uri),
		commLogger:       commlog,
		evtBreak:         xevent.NewEvent(),
		evtKill:          xevent.NewEvent(),
		ErrorDelay:       ERROR_DELAY,
		PollInterval:     POLL_INTERVAL,
		PollChunkSize:    POLL_CHUNKSIZE,
		PollMaxSize:      POLL_MAXSIZE,
		PollWaitNullRead: POLL_WAITNULLREAD,
	}
	if opts != nil {
		if v := opts.GetFloat64("poll_interval", 0); v > 0 {
			conn.PollInterval = v
		}
		if v := opts.GetInt("poll_chunksize", 0); v > 0 {
			conn.PollChunkSize = v
		}
		if v := opts.GetInt("poll_maxsize", 0); v >= 0 {
			conn.PollMaxSize = v
		}
		if v := opts.GetBool("poll_waitnullread", false); v {
			conn.PollWaitNullRead = v
		}
	}
	return conn
}

// implement Stringer interface
func (bs *baseConnection) String() string {
	return fmt.Sprintf("<Connection: %v>", bs.uri)
}

func (bs *baseConnection) Uri() string {
	return bs.uri
}

func (bs *baseConnection) Type() string {
	return strings.ToLower(strings.SplitN(bs.uri, "@", 2)[0])
}

func (bs *baseConnection) comm_log(msg string, args ...any) {
	if bs.commLogger != nil && msg != "" {
		if bs.uriLogging {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.commLogger.Info(msg, args...)
	}
}

// internal communication TX logging
func (bs *baseConnection) tx_Log(data []byte) {
	if bs.commLogger != nil && data != nil {
		msg := "TX >> " + strings.ToUpper(hex.EncodeToString(data))
		if bs.uriLogging {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.commLogger.Info(msg)
	}
}

// internal communication RX logging
func (bs *baseConnection) rx_Log(data []byte) {
	if bs.commLogger != nil && data != nil {
		msg := "RX << " + strings.ToUpper(hex.EncodeToString(data))
		if bs.uriLogging {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.commLogger.Info(msg)
	}
}
