package xcomm

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/types"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	DEFAULT_ERRORDELAY    = float64(1)
	DEFAULT_POLLINTERVAL  = float64(0.005)
	DEFAULT_POLLCHUNKSIZE = int(4096)
	DEFAULT_POLLMAXSIZE   = int(0)
)

type Logger = xlog.Logger
type Options = types.NDict

// interface representing connection
type Connection interface {
	Uri() string
	Type() string
	Parent() Listener
	IsOpened() bool
	Open() error
	Close()
	Cancel()
	Send([]byte) error
	Recv() ([]byte, error)
	RecvWait(float64) ([]byte, error)
	Sleep(float64) bool
}

// interface representing listener
type Listener interface {
	Uri() string
	Type() string
	IsActive() bool
	Start() error
	Stop()
	Sleep(float64) bool
	SetConnHandler(func(Connection))
}

// BaseConnection is a base structure for connection handling
type BaseConnection struct {
	Log        *Logger
	uri        string
	uriLogging bool

	// operation events
	evtBreak  *xevent.Event
	evtKill   *xevent.Event
	ctxCancel context.CancelFunc

	// error delay for execution loop
	ErrorDelay float64

	// read/write polling params
	PollInterval  float64
	PollChunkSize int
	PollMaxSize   int
}

func new_base_connection(
	uri string, log *Logger, opts Options) *BaseConnection {
	conn := &BaseConnection{
		uri:           strings.TrimSpace(uri),
		Log:           log,
		evtBreak:      xevent.NewEvent(),
		evtKill:       xevent.NewEvent(),
		ErrorDelay:    DEFAULT_ERRORDELAY,
		PollInterval:  DEFAULT_POLLINTERVAL,
		PollChunkSize: DEFAULT_POLLCHUNKSIZE,
		PollMaxSize:   DEFAULT_POLLMAXSIZE,
	}
	if opts != nil {
		conn.ErrorDelay = opts.GetFloat64(
			"error_delay", DEFAULT_ERRORDELAY)
		conn.PollInterval = opts.GetFloat64(
			"poll_interval", DEFAULT_POLLINTERVAL)
		conn.PollChunkSize = opts.GetInt(
			"poll_chunksize", DEFAULT_POLLCHUNKSIZE)
		conn.PollMaxSize = opts.GetInt(
			"poll_maxsize", DEFAULT_POLLMAXSIZE)
	}
	return conn
}

// implement Stringer interface
func (bs *BaseConnection) String() string {
	return fmt.Sprintf("<Connection: %v>", bs.uri)
}

func (bs *BaseConnection) Uri() string {
	return bs.uri
}

func (bs *BaseConnection) Type() string {
	return strings.ToLower(strings.SplitN(bs.uri, "@", 2)[0])
}

// Sleep pauses the execution for the specified duration in seconds.
func (bs *BaseConnection) Sleep(timeout float64) bool {
	return bs.evtKill.Wait(timeout)
}

func (bs *BaseConnection) log(msg string, args ...any) {
	if bs.Log != nil && msg != "" {
		bs.Log.Info(msg, args...)
	}
}

// internal communication TX logging
func (bs *BaseConnection) txLog(data []byte) {
	if bs.Log != nil && data != nil {
		msg := "TX >> " + strings.ToUpper(hex.EncodeToString(data))
		if bs.uriLogging {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.Log.Info(msg)
	}
}

// internal communication RX logging
func (bs *BaseConnection) rxLog(data []byte) {
	if bs.Log != nil && data != nil {
		msg := "RX << " + strings.ToUpper(hex.EncodeToString(data))
		if bs.uriLogging {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.Log.Info(msg)
	}
}

////////////////////////////////////// utils

// create new connection handler
func NewConnection(
	uri string, log *Logger, opts Options) (Connection, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialConnection(uri, log, opts)
	default:
		return NewNetConnection(uri, log, opts)
	}
}

// create new listener handler
func NewListener(
	uri string, log *Logger, opts Options) (Listener, error) {
	t := strings.ToLower(strings.SplitN(uri, "@", 2)[0])
	switch t {
	case "serial":
		return NewSerialListener(uri, log, opts)
	default:
		return NewNetListener(uri, log, opts)
	}
}
