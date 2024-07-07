package xcomm

import (
	"strings"

	"github.com/exonlabs/go-utils/pkg/types"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

type Logger = xlog.Logger
type Options = types.NDict

const (
	ERROR_DELAY       = float64(0.5)
	POLL_INTERVAL     = float64(0.2)
	POLL_CHUNKSIZE    = int(1024)
	POLL_MAXSIZE      = int(0)
	POLL_WAITNULLREAD = false
)

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
}

// interface representing listener
type Listener interface {
	Uri() string
	Type() string
	IsActive() bool
	Start() error
	Stop()
	SetConnHandler(func(Connection))
}

////////////////////////////////////// creator functions

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
