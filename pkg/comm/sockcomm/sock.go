// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package sockcomm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/netutil"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// ParseUri parses a net socket URI into params.
//
//	The expected URI format is `sock@<path>`
//
//	example:
//	server
//	   - sock@/path/to/sock/file
//	client
//	   - sock@/path/to/sock/file
//
// Returns the sock params and any error encountered.
func ParseUri(uri string) (string, error) {
	parts := strings.SplitN(uri, "@", 2)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "sock" {
		return "", comm.ErrUri
	}

	return filepath.Clean(parts[1]), nil
}

// GetAddr returns (net.Addr) type for net socket path.
func GetAddr(path string) (net.Addr, error) {
	return net.ResolveUnixAddr("unix", path)
}

/////////////////////////////////////////////////////

// Connection represents a net socket connection with event support and logging.
type Connection struct {
	// Context containing common attributes and functions.
	*comm.Context

	// The file system path for net socket.
	path string

	// The underlying network connection.
	netConn net.Conn

	// The parent Listener (if any), managing the connection.
	parent *Listener

	// isOpened represents the connecton status, opened or closed.
	isOpened atomic.Bool
	// closeEvent signals a close operation.
	closeEvent atomic.Bool
	// breakReadEvent signals a read interrupt operation.
	breakReadEvent atomic.Bool

	// sMutex defines mutex for state change operations (open/close).
	sMutex sync.Mutex
	// rMutex defines mutex for read operations.
	rMutex sync.Mutex
	// wMutex defines mutex for write operations.
	wMutex sync.Mutex
	// rwWaitGrp defines wait group for read/write operations.
	rwWaitGrp sync.WaitGroup
}

// NewConnection creates and initializes a new Connection for the given URI.
// The URI specifies the sock path.
func NewConnection(uri string, log *logging.Logger, opts dictx.Dict) (*Connection, error) {
	path, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	return &Connection{
		Context: comm.NewContext(uri, log, opts),
		path:    path,
	}, nil
}

// String returns a string representation of the Connection.
func (c *Connection) String() string {
	return fmt.Sprintf("<SockConnection: %s>", c.Uri())
}

// NetConn returns the net connection instance.
func (c *Connection) NetConn() net.Conn {
	return c.netConn
}

// Parent retrieves the parent Listener, if any, associated with the Connection.
func (c *Connection) Parent() comm.Listener {
	return c.parent
}

// IsOpened indicates whether the connection is currently open and active.
func (c *Connection) IsOpened() bool {
	return c.isOpened.Load() && !c.closeEvent.Load()
}

// Open establishes the connection.
func (c *Connection) Open(timeout float64) error {
	// take no action if managed by parent listener
	if c.parent != nil {
		return nil
	}

	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	// do nothing if already opened
	if c.isOpened.Load() {
		return nil
	}

	dialer := net.Dialer{
		KeepAlive: -1, // disabled
	}
	if timeout > 0 {
		dialer.Timeout = time.Duration(timeout * float64(time.Second))
	}

	conn, err := dialer.Dial("unix", c.path)
	if err != nil {
		c.LogMsg("CONNECT_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	c.LogMsg("CONNECTED -- %s", c.Uri())
	c.netConn = conn

	c.closeEvent.Store(false)
	c.isOpened.Store(true)
	return nil
}

// Close shuts down the connection and cleaning up resources.
func (c *Connection) Close() {
	// take no action if managed by parent listener
	if c.parent != nil {
		return
	}

	c.closeEvent.Store(true)

	c.sMutex.Lock()
	defer c.sMutex.Unlock()

	// do nothing if already closed
	if !c.isOpened.Load() {
		return
	}

	// close connection
	c.netConn.Close()

	c.rwWaitGrp.Wait()
	c.LogMsg("DISCONNECTED -- %s", c.Uri())
	c.isOpened.Store(false)
}

// Cancel cancels any ongoing operations on the connection.
func (c *Connection) Cancel() {
	c.breakReadEvent.Store(true)
}

// Cancel interrupts the ongoing sending operation for this Connection.
func (c *Connection) CancelSend() {
	// do nothing, not available for serial port
}

// Cancel interrupts the ongoing receiving operation for this Connection.
func (c *Connection) CancelRecv() {
	c.breakReadEvent.Store(true)
}

// Send transmits data over the connection, with a specified timeout.
func (c *Connection) Send(data []byte, timeout float64) error {
	return c.SendTo(data, nil, timeout)
}

// SendTo transmits data to addr over the connection, with a specified timeout.
func (c *Connection) SendTo(data []byte, addr any, timeout float64) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	// Acquire write lock
	c.wMutex.Lock()
	defer c.wMutex.Unlock()

	// Check connection state after acquiring the lock
	if c.closeEvent.Load() || !c.isOpened.Load() {
		return comm.ErrClosed
	}

	c.rwWaitGrp.Add(1)
	defer c.rwWaitGrp.Done()

	c.LogTx(data, nil)
	if timeout > 0 {
		c.netConn.SetWriteDeadline(time.Now().Add(
			time.Duration(timeout * float64(time.Second))))
	}
	n, err := c.netConn.Write(data)
	if err == nil && n != len(data) {
		err = errors.New("partial data sent")
	}

	if err != nil {
		if comm.IsClosedError(err) {
			c.closeEvent.Store(true)
			c.LogMsg("CONN_CLOSED -- %v", err)
			go c.Close()
			return comm.ErrClosed
		}
		c.LogMsg("SEND_ERROR -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrWrite, err)
	}

	return nil
}

// Recv waits for incoming data over the connection until a timeout
// or interrupt event occurs. Setting timeout=0 will wait indefinitely.
func (c *Connection) Recv(timeout float64) ([]byte, error) {
	b, _, err := c.RecvFrom(timeout)
	return b, err
}

// Recv waits for incoming data from addr over the connection until a timeout
// or interrupt event occurs. Setting timeout=0 will wait indefinitely.
func (c *Connection) RecvFrom(timeout float64) ([]byte, any, error) {
	// Acquire read lock
	c.rMutex.Lock()
	defer c.rMutex.Unlock()

	// Check connection state after acquiring the lock
	if c.closeEvent.Load() || !c.isOpened.Load() {
		return nil, nil, comm.ErrClosed
	}

	c.rwWaitGrp.Add(1)
	defer c.rwWaitGrp.Done()

	c.breakReadEvent.Store(false)

	// determine read buffer size and polling timeout
	nRead := c.PollChunkSize
	if c.PollMaxSize > 0 {
		nRead = c.PollMaxSize
	}

	tPoll := time.Duration(c.PollTimeout * float64(time.Second))
	if tPoll <= 0 {
		tPoll = time.Duration(comm.POLL_TIMEOUT * float64(time.Second))
	}

	var tBreak time.Time
	if timeout > 0 {
		tBreak = time.Now().Add(
			time.Duration(timeout * float64(time.Second)))
	}

	var data []byte

	b := make([]byte, nRead)
	for {
		c.netConn.SetReadDeadline(time.Now().Add(tPoll))
		n, err := c.netConn.Read(b)
		if err != nil {
			if comm.IsClosedError(err) {
				c.closeEvent.Store(true)
				c.LogMsg("CONN_CLOSED -- %v", err)
				go c.Close()
				return nil, nil, comm.ErrClosed
			}
			if _, ok := err.(net.Error); !ok || !err.(net.Error).Timeout() {
				c.LogMsg("RECV_ERROR -- %v", err)
				return nil, nil, fmt.Errorf("%w, %v", comm.ErrRead, err)
			}
		}

		if n > 0 {
			data = append(data, b[:n]...)
			if c.PollMaxSize > 0 {
				nRead -= n
				if nRead <= 0 {
					break
				} else {
					b = b[:nRead]
				}
			}
		} else if len(data) > 0 {
			break
		}

		if c.parent != nil && c.parent.stopEvent.Load() {
			return nil, nil, comm.ErrClosed
		}
		if c.breakReadEvent.Load() {
			return nil, nil, comm.ErrBreak
		}
		if timeout > 0 && time.Now().After(tBreak) {
			return nil, nil, comm.ErrTimeout
		}
	}

	c.LogRx(data, nil)
	return data, nil, nil
}

/////////////////////////////////////////////////////

// Listener represents a net socket listener that handles incoming connections
// with a custom connection handler.
type Listener struct {
	// Context containing common attributes such as logging and events.
	*comm.Context

	// The file system path of net socket to listen on.
	path string

	// The underlying net listener.
	netListener net.Listener

	// The handler function to be called when a new connection is accepted.
	connectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool
	// stopEvent signals a stop operation.
	stopEvent atomic.Bool

	// sMutex defines mutex for state change operations (start/stop).
	sMutex sync.Mutex
}

// NewListener creates a new net socket Listener.
// The parsed options are:
//   - connections_limit: (int) the limit on number of concurrent connections.
//     use 0 to disable connections limit.
func NewListener(uri string, log *logging.Logger, opts dictx.Dict) (*Listener, error) {
	path, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Context: comm.NewContext(uri, log, opts),
		path:    path,
	}, nil
}

// String returns a string representation of the net socket Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<SockListener: %s>", l.Uri())
}

// NetListener returns the net listener instance.
func (l *Listener) NetListener() net.Listener {
	return l.netListener
}

// SetConnHandler sets a callback function to handle connections.
func (l *Listener) SetConnHandler(h func(comm.Connection)) {
	l.connectionHandler = h
}

// IsActive checks if the net socket listener is currently active.
func (l *Listener) IsActive() bool {
	return l.isActive.Load() && !l.stopEvent.Load()
}

func (l *Listener) startListener() error {
	cfg := net.ListenConfig{
		KeepAlive: -1, // disabled
	}
	if v := dictx.GetFloat(l.Options, "keepalive_interval", -1); v >= 0 {
		cfg.KeepAlive = time.Duration(v * float64(time.Second))
	}

	// listener instance
	netListener, err := cfg.Listen(context.Background(), "unix", l.path)
	if err != nil {
		return err
	}
	// set connection limit (if configured)
	if v := dictx.GetInt(l.Options, "connections_limit", 0); v > 0 {
		netListener = netutil.LimitListener(netListener, v)
	}
	l.LogMsg("LISTENING -- %s", l.Uri())
	l.netListener = netListener

	var waitGrp sync.WaitGroup

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		netListener.Close()
		// wait all connections handlers termination
		waitGrp.Wait()
		os.Remove(l.path)
		l.LogMsg("CLOSED -- %s", l.Uri())
		l.isActive.Store(false)
	}()

	for !l.stopEvent.Load() {
		// wait for new connection
		c, err := netListener.Accept()
		if err != nil {
			if comm.IsClosedError(err) {
				break
			} else {
				l.LogMsg("CONN_ERROR -- %v", err)
				continue
			}
		}

		// handle new connection
		waitGrp.Add(1)
		go func(netConn net.Conn) {
			uri := fmt.Sprintf("%s@%s", l.Type(), netConn.RemoteAddr())
			nc, err := NewConnection(uri, l.CommLog, l.Options)
			if err != nil {
				l.LogMsg("CONN_ERROR -- %v", err)
				netConn.Close()
				return
			}
			nc.netConn = netConn
			nc.parent = l
			nc.isOpened.Store(true)
			nc.LogMsg("CONNECTED")

			defer func() {
				netConn.Close()
				nc.LogMsg("DISCONNECTED")
				waitGrp.Done()
			}()

			l.connectionHandler(nc)
		}(c)
	}

	return nil
}

// Start begins listening for connections, calling the connectionHandler
// for each established connection.
func (l *Listener) Start() error {
	if l.connectionHandler == nil {
		return errors.New("empty connection handler")
	}

	// error if already started
	if !l.sMutex.TryLock() {
		return errors.New("Listener already started")
	}
	defer l.sMutex.Unlock()

	return l.startListener()
}

// Stop gracefully shuts down the listener.
func (l *Listener) Stop() {
	l.stopEvent.Store(true)

	// do nothing if already stopped
	if !l.isActive.Load() {
		return
	}

	l.netListener.Close()
}
