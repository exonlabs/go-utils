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

// ParseUri parses a network socket URI.
//
//	The expected URI format is `sock@<path>`
//
//	example:
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
	// uri specifies the resource identifier.
	uri string
	// The file system path for net socket.
	path string
	// The underlying network connection.
	netConn net.Conn

	// The parent Listener (if any), managing the connection.
	parent *Listener

	// isOpened represents the connecton status, opened or closed.
	isOpened atomic.Bool
	// closeEvent signals a closing operation.
	closeEvent atomic.Bool
	// breakRecvEvent signals a receive break interrupt operation.
	breakRecvEvent atomic.Bool

	// muState defines mutex for state change operations (open/close).
	muState sync.Mutex
	// muSend defines mutex for write operations.
	muSend sync.Mutex
	// muRecv defines mutex for read operations.
	muRecv sync.Mutex
	// wgClose defines wait group for close operations.
	wgClose sync.WaitGroup

	// Log is the logger instance for communication data logging.
	Log *logging.Logger

	// PollConfig defines the read polling.
	PollConfig *comm.PollingConfig
}

// NewConnection creates and initializes a new Connection for the given URI.
//
// The parsed options are:
//   - Polling Options: detailed in [comm.ParsePollingConfig]
func NewConnection(uri string, log *logging.Logger, opts dictx.Dict) (*Connection, error) {
	uri = strings.TrimSpace(uri)
	path, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		uri:  uri,
		path: path,
		Log:  log,
	}

	// set polling options
	c.PollConfig, err = comm.ParsePollingConfig(opts)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// String returns a string representation of the Connection.
func (c *Connection) String() string {
	return fmt.Sprintf("<SockConnection: %s>", c.uri)
}

// Uri returns the URI of the connection
func (c *Connection) Uri() string {
	return c.uri
}

// Type returns the type of the connection as inferred from the Uri.
func (c *Connection) Type() string {
	return "sock"
}

// NetConn returns the net connection instance.
func (c *Connection) NetConn() net.Conn {
	return c.netConn
}

// Parent returns the parent Listener if any is associated with the Connection.
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

	c.muState.Lock()
	defer c.muState.Unlock()

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
		comm.LogMsg(c.Log, "CONNECT_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	comm.LogMsg(c.Log, "CONNECTED -- %s", c.uri)
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

	c.muState.Lock()
	defer c.muState.Unlock()

	// do nothing if already closed
	if !c.isOpened.Load() {
		return
	}

	// close connection
	c.netConn.Close()

	c.wgClose.Wait()
	comm.LogMsg(c.Log, "DISCONNECTED -- %s", c.uri)
	c.isOpened.Store(false)
}

// Cancel cancels any ongoing operations on the connection.
func (c *Connection) Cancel() {
	c.CancelSend()
	c.CancelRecv()
}

// Cancel interrupts the ongoing sending operation for this Connection.
func (c *Connection) CancelSend() {
	// do nothing
}

// Cancel interrupts the ongoing receiving operation for this Connection.
func (c *Connection) CancelRecv() {
	c.breakRecvEvent.Store(true)
}

// Send transmits data over the connection, with a specified timeout.
func (c *Connection) Send(data []byte, timeout float64) error {
	return c.SendTo(data, nil, timeout)
}

// SendTo transmits data to addr over the connection, with a specified timeout.
//
// Setting timeout 0 or negative value will wait indefinitely.
func (c *Connection) SendTo(data []byte, addr any, timeout float64) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	// Acquire send lock
	c.muSend.Lock()
	defer c.muSend.Unlock()

	// Check connection state after acquiring the lock
	if !c.isOpened.Load() || c.closeEvent.Load() {
		return comm.ErrClosed
	}

	c.wgClose.Add(1)
	defer c.wgClose.Done()

	comm.LogTx(c.Log, data, nil)
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
			comm.LogMsg(c.Log, "CONN_CLOSED -- %v", err)
			go c.Close()
			return comm.ErrClosed
		}
		comm.LogMsg(c.Log, "SEND_ERROR -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrSend, err)
	}

	return nil
}

// Recv waits for incoming data over the connection until a timeout
// or interrupt event occurs. Setting timeout=0 will wait indefinitely.
func (c *Connection) Recv(timeout float64) ([]byte, error) {
	b, _, err := c.RecvFrom(timeout)
	return b, err
}

// RecvFrom waits for incoming data from addr over the connection until a timeout
// or interrupt event occurs.
//
// Setting timeout 0 or negative value will wait indefinitely.
func (c *Connection) RecvFrom(timeout float64) ([]byte, any, error) {
	// Acquire read lock
	c.muRecv.Lock()
	defer c.muRecv.Unlock()

	// Check connection state after acquiring the lock
	if !c.isOpened.Load() || c.closeEvent.Load() {
		return nil, nil, comm.ErrClosed
	}

	c.wgClose.Add(1)
	defer c.wgClose.Done()

	c.breakRecvEvent.Store(false)

	// determine read buffer size
	nRead := c.PollConfig.ChunkSize
	if c.PollConfig.MaxSize > 0 {
		nRead = c.PollConfig.MaxSize
	}

	// set read polling duration and deadline
	var tPolling time.Duration
	var tDeadline time.Time

	tPolling = time.Duration(
		c.PollConfig.Timeout * float64(time.Second))
	if timeout > 0 {
		tDeadline = time.Now().Add(
			time.Duration(timeout * float64(time.Second)))
	}

	var data []byte

	b := make([]byte, nRead)
	for {
		c.netConn.SetReadDeadline(time.Now().Add(tPolling))
		n, err := c.netConn.Read(b)
		if err != nil {
			if comm.IsClosedError(err) {
				c.closeEvent.Store(true)
				comm.LogMsg(c.Log, "CONN_CLOSED -- %v", err)
				go c.Close()
				return nil, nil, comm.ErrClosed
			}
			if _, ok := err.(net.Error); !ok || !err.(net.Error).Timeout() {
				comm.LogMsg(c.Log, "RECV_ERROR -- %v", err)
				return nil, nil, fmt.Errorf("%w, %v", comm.ErrRecv, err)
			}
		}

		if n > 0 {
			data = append(data, b[:n]...)
			if c.PollConfig.MaxSize > 0 {
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
		if c.breakRecvEvent.Load() {
			return nil, nil, comm.ErrBreak
		}
		if timeout > 0 && time.Now().After(tDeadline) {
			return nil, nil, comm.ErrTimeout
		}
	}

	comm.LogRx(c.Log, data, nil)
	return data, nil, nil
}

/////////////////////////////////////////////////////

// Listener represents a net socket listener that handles incoming connections
// with a custom connection handler.
type Listener struct {
	// uri specifies the resource identifier.
	uri string
	// The file system path of net socket to listen on.
	path string
	// The underlying net listener.
	netListener net.Listener

	// ConnectionHandler defines the function to handle incoming connections.
	ConnectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool
	// stopEvent signals a stop operation.
	stopEvent atomic.Bool

	// muState defines mutex for state change operations (start/stop).
	muState sync.Mutex

	// Log is the logger instance for communication data logging.
	Log *logging.Logger

	// PollConfig defines the read polling.
	PollConfig *comm.PollingConfig
	// LimiterConfig defines the limits for TCP connections.
	LimiterConfig *comm.LimiterConfig
}

// NewListener creates a new Listener.
//
// The parsed options are:
//   - Polling Options: detailed in [comm.ParsePollingConfig]
//   - Limiter Options: detailed in [comm.ParseLimiterConfig]
func NewListener(uri string, log *logging.Logger, opts dictx.Dict) (*Listener, error) {
	uri = strings.TrimSpace(uri)
	path, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		uri:  uri,
		path: path,
		Log:  log,
	}

	// set polling options
	l.PollConfig, err = comm.ParsePollingConfig(opts)
	if err != nil {
		return nil, err
	}

	// set limiter options
	l.LimiterConfig, err = comm.ParseLimiterConfig(opts)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// String returns a string representation of the Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<SockListener: %s>", l.uri)
}

// Uri returns the URI of the listener
func (l *Listener) Uri() string {
	return l.uri
}

// Type returns the type of the listener as inferred from the Uri.
func (l *Listener) Type() string {
	return "sock"
}

// NetListener returns the net listener instance.
func (l *Listener) NetListener() net.Listener {
	return l.netListener
}

// SetConnHandler sets a callback function to handle connections.
func (l *Listener) SetConnHandler(handler func(comm.Connection)) {
	l.ConnectionHandler = handler
}

// IsActive checks if the listener is currently active.
func (l *Listener) IsActive() bool {
	return l.isActive.Load() && !l.stopEvent.Load()
}

func (l *Listener) startListener() error {
	lConfig := net.ListenConfig{
		KeepAlive: -1, // disabled
	}

	// net listener
	netListener, err := lConfig.Listen(
		context.Background(), "unix", l.path)
	if err != nil {
		return err
	}
	// set connection limit
	if l.LimiterConfig.SimultaneousConn > 0 {
		netListener = netutil.LimitListener(
			netListener, l.LimiterConfig.SimultaneousConn)
	}
	comm.LogMsg(l.Log, "LISTENING -- %s", l.uri)
	l.netListener = netListener

	var wg sync.WaitGroup

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		netListener.Close()
		// wait all connections handlers termination
		wg.Wait()
		os.Remove(l.path)
		comm.LogMsg(l.Log, "CLOSED -- %s", l.uri)
		l.isActive.Store(false)
	}()

	for !l.stopEvent.Load() {
		// wait for new connection
		new_conn, err := netListener.Accept()
		if err != nil {
			if comm.IsClosedError(err) {
				break
			} else {
				comm.LogMsg(l.Log, "CONN_ERROR -- %v", err)
				continue
			}
		}

		// handle new connection
		wg.Add(1)
		go func(conn net.Conn) {
			c := &Connection{
				uri:        l.uri,
				path:       l.path,
				netConn:    conn,
				Log:        l.Log,
				PollConfig: l.PollConfig,
			}
			c.parent = l
			c.isOpened.Store(true)
			comm.LogMsg(c.Log, "CONNECTED")

			defer func() {
				conn.Close()
				comm.LogMsg(c.Log, "DISCONNECTED")
				wg.Done()
			}()

			l.ConnectionHandler(c)
		}(new_conn)
	}

	return nil
}

// Start begins listening for connections, calling the connectionHandler
// for each established connection.
func (l *Listener) Start() error {
	if l.ConnectionHandler == nil {
		return errors.New("empty connection handler")
	}

	// error if already started
	if !l.muState.TryLock() {
		return errors.New("Listener already started")
	}
	defer l.muState.Unlock()

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
