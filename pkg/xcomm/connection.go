package xcomm

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	defaultErrorDelay    = float64(1)
	defaultPollInterval  = float64(0.005)
	defaultPollChunkSize = int(4096)
	defaultPollMaxSize   = int(0)
)

// interface representing connection
type Connection interface {
	Uri() string
	Type() string
	Parent() Listener
	IsOpened() bool
	Open() error
	Close()
	Send([]byte) error
	Recv() ([]byte, error)
	RecvWait(float64) ([]byte, error)
	Cancel()
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
	SetHandler(func(Connection))
}

// BaseConnection is a base structure for connection handling
type BaseConnection struct {
	Log    *xlog.Logger
	uri    string
	logUri bool

	// operation events
	evtBreak *xevent.Event
	evtKill  *xevent.Event

	// error delay for execution loop
	ErrorDelay float64

	// read/write polling params
	PollInterval  float64
	PollChunkSize int
	PollMaxSize   int
}

func newBaseConnection(uri string, log *xlog.Logger) *BaseConnection {
	return &BaseConnection{
		uri:           strings.TrimSpace(strings.ToLower(uri)),
		Log:           log,
		evtBreak:      xevent.NewEvent(),
		evtKill:       xevent.NewEvent(),
		ErrorDelay:    defaultErrorDelay,
		PollInterval:  defaultPollInterval,
		PollChunkSize: defaultPollChunkSize,
		PollMaxSize:   defaultPollMaxSize,
	}
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
		if bs.logUri {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.Log.Info(msg)
	}
}

// internal communication RX logging
func (bs *BaseConnection) rxLog(data []byte) {
	if bs.Log != nil && data != nil {
		msg := "RX << " + strings.ToUpper(hex.EncodeToString(data))
		if bs.logUri {
			msg = "[" + bs.uri + "]  " + msg
		}
		bs.Log.Info(msg)
	}
}
