package xconn

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

// Define common errors
var (
	ErrError    = errors.New("")
	ErrOpen     = fmt.Errorf("%wopen socket failed", ErrError)
	ErrNotOpend = fmt.Errorf("%wnot opened!", ErrError)
	ErrClosed   = fmt.Errorf("%wsocket closed", ErrError)
	ErrBreak    = fmt.Errorf("%woperation break", ErrError)
	ErrTimeout  = fmt.Errorf("%woperation timeout", ErrError)
	ErrRead     = fmt.Errorf("%wread socket failed", ErrError)
	ErrWrite    = fmt.Errorf("%wwrite socket failed", ErrError)
)

const (
	defaultPollInterval  = float64(0.1)
	defaultPollChunkSize = int(4096)
	defaultPollMaxSize   = int(0)
)

// ClientConn is an interface representing a client connection
type ClientConn interface {
	Type() string
	IsOpened() bool
	Open() error
	Close() error
	Send([]byte) error
	Recv() ([]byte, error)
	RecvWait(timeout float64) ([]byte, error)
	Cancel() error
	// Sleep(float64) bool
}

// ServerConn is an interface representing a server connection
type ServerConn interface {
	Open() error
	Close() error
}

// BaseConnection is a base structure for connection handling
type BaseConnection struct {
	Log        *xlog.Logger
	Uri        string
	EvtBreak   *xevent.Event
	EvtKill    *xevent.Event
	ErrorDelay float64

	// read/write polling params
	PollInterval  float64
	PollChunkSize int
	PollMaxSize   int
}

func NewBaseConnection(uri string, log *xlog.Logger) *BaseConnection {
	return &BaseConnection{
		Uri:           uri,
		Log:           log,
		EvtBreak:      xevent.NewEvent(),
		EvtKill:       xevent.NewEvent(),
		PollInterval:  defaultPollInterval,
		PollChunkSize: defaultPollChunkSize,
		PollMaxSize:   defaultPollMaxSize,
	}

}

// Cancel cancels the connection
func (bs *BaseConnection) Cancel() error {
	bs.EvtBreak.Set()
	return nil
}

// Sleep pauses the execution for the specified duration in seconds.
// func (bs *BaseConnection) Sleep(timeout float64) bool {

// 	return true
// }

// internal communication TX logging
func (bs *BaseConnection) TxLog(data []byte) {
	if bs.Log != nil {
		bs.Log.Info("TX >> %s", strings.ToUpper(hex.EncodeToString(data)))
	}
}

// internal communication RX logging
func (bs *BaseConnection) RxLog(data []byte) {
	if bs.Log != nil {
		bs.Log.Info("RX << %s", strings.ToUpper(hex.EncodeToString(data)))
	}
}
