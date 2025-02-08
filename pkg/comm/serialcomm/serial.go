// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package serialcomm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.bug.st/serial"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/abc/gx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// ParseUri parses a serial URI.
//
//	The expected URI format is `serial@<port>:<baud>:<mode>`
//
//	<port>  serial port name (e.g., /dev/ttyS0 or COM1)
//	<baud>  Baud rate (e.g., 4800,9600,19200,115200...)
//	<mode>  bytesize, parity and stopbits
//	        {8|7}{N|E|O|M|S}{1|2}
//
//	example:
//	   - serial@/dev/ttyS0:9600:8N1          (linux)
//	   - serial@COM1:9600:8N1                (windows)
//
// Returns the serial params and any error encountered.
func ParseUri(uri string) (string, *serial.Mode, error) {
	parts := strings.SplitN(uri, "@", 2)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "serial" {
		return "", nil, comm.ErrUri
	}
	// parts after @
	parts = strings.Split(parts[1], ":")
	if len(parts) < 3 {
		return "", nil, comm.ErrUri
	}

	port, strBaud, strMode := parts[0], parts[1], parts[2]
	if len(strMode) != 3 {
		return "", nil, comm.ErrUri
	}

	var err error
	mode := &serial.Mode{}

	if mode.BaudRate, err = strconv.Atoi(strBaud); err != nil {
		return "", nil, comm.ErrUri
	}
	if mode.DataBits, err = strconv.Atoi(string(strMode[0])); err != nil {
		return "", nil, comm.ErrUri
	}
	switch string(strMode[1]) {
	case "n", "N":
		mode.Parity = serial.NoParity
	case "o", "O":
		mode.Parity = serial.OddParity
	case "e", "E":
		mode.Parity = serial.EvenParity
	case "m", "M":
		mode.Parity = serial.MarkParity
	case "s", "S":
		mode.Parity = serial.SpaceParity
	default:
		return "", nil, comm.ErrUri
	}
	switch string(strMode[2]) {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "2":
		mode.StopBits = serial.TwoStopBits
	default:
		return "", nil, comm.ErrUri
	}

	return port, mode, nil
}

/////////////////////////////////////////////////////

// Connection represents a serial connection with event support and logging.
type Connection struct {
	// Context containing common attributes and functions.
	*comm.Context

	// uri specifies the resource identifier.
	uri string
	// The serial port identifier (e.g., COM1, /dev/ttyS0).
	port string
	// Configuration of serial communication parameters.
	mode *serial.Mode
	// The underlying serial port connection.
	serialPort serial.Port

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
}

// NewConnection creates and initializes a new Connection for the given URI.
func NewConnection(uri string, commlog *logging.Logger, opts dictx.Dict) (*Connection, error) {
	uri = strings.TrimSpace(uri)
	port, mode, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		Context: comm.NewContext(commlog, opts),
		uri:     uri,
		port:    port,
		mode:    mode,
	}

	// set dynamic polling relative to baudrate. actual 1 byte = 10 bits
	// set to max of "user defined", "0.02 mSec" or "20 bytes duration".
	c.PollTimeout = gx.Max(
		c.PollTimeout, 0.02, 20.0/float64(mode.BaudRate))

	return c, nil
}

// String returns a string representation of the Connection.
func (c *Connection) String() string {
	return fmt.Sprintf("<SerialConnection: %s>", c.uri)
}

// Uri returns the URI of the connection
func (c *Connection) Uri() string {
	return c.uri
}

// Type returns the type of the connection as inferred from the Uri.
func (c *Connection) Type() string {
	return "serial"
}

// SerialPort returns the underlying serial port object
func (c *Connection) SerialPort() serial.Port {
	return c.serialPort
}

// Parent returns the parent Listener if any is associated with the Connection.
func (c *Connection) Parent() comm.Listener {
	return c.parent
}

// IsOpened indicates whether the connection is currently open and active.
func (c *Connection) IsOpened() bool {
	return c.isOpened.Load() && !c.closeEvent.Load()
}

// Open opens the serial port.
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

	com, err := serial.Open(c.port, c.mode)
	if err != nil {
		comm.LogMsg(c.CommLog, "OPEN_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	comm.LogMsg(c.CommLog, "OPENED -- %s", c.uri)
	c.serialPort = com
	c.serialPort.ResetInputBuffer()
	c.serialPort.ResetOutputBuffer()

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

	// close port clean buffers
	c.serialPort.Close()
	c.serialPort.ResetInputBuffer()
	c.serialPort.ResetOutputBuffer()

	c.wgClose.Wait()
	comm.LogMsg(c.CommLog, "CLOSED -- %s", c.uri)
	c.isOpened.Store(false)
}

// Mode sets new mode for the serial port.
// mode has the format `<baud>:<mode>` as defined for the URI.
func (c *Connection) SetMode(mode string) error {
	newUri := fmt.Sprintf("serial@%s:%s", c.port, mode)
	_, newMode, err := ParseUri(newUri)
	if err != nil {
		return err
	}

	c.muState.Lock()
	defer c.muState.Unlock()

	// set new params
	c.uri = newUri
	c.mode = newMode

	// set dynamic polling relative to baudrate. actual 1 byte = 10 bits
	// set to max of "user defined", "0.02 mSec" or "20 bytes duration".
	c.PollTimeout = gx.Max(0.02, 20.0/float64(newMode.BaudRate))

	// apply new mode if port is already opened
	if c.isOpened.Load() {
		comm.LogMsg(c.CommLog, "SETMODE -- %s", newUri)
		if err := c.serialPort.SetMode(newMode); err != nil {
			return fmt.Errorf("%w, %v", comm.ErrConnection, err)
		}
		c.serialPort.ResetInputBuffer()
		c.serialPort.ResetOutputBuffer()
	}

	return nil
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
func (c *Connection) SendTo(data []byte, _ any, timeout float64) error {
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

	comm.LogTx(c.CommLog, data, nil)
	n, err := c.serialPort.Write(data)
	if err == nil {
		// Ensure all data is flushed from the OS buffer
		err = c.serialPort.Drain()
		if err == nil && n != len(data) {
			err = errors.New("partial data written to port")
		}
	}

	if err != nil {
		if comm.IsClosedError(err) {
			c.closeEvent.Store(true)
			comm.LogMsg(c.CommLog, "PORT_CLOSED -- %v", err)
			go c.Close()
			return comm.ErrClosed
		}
		comm.LogMsg(c.CommLog, "SEND_ERROR -- %v", err)
		c.serialPort.ResetOutputBuffer()
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

	// determine read buffer size and polling timeout
	nRead := c.PollChunkSize
	if c.PollMaxSize > 0 {
		nRead = c.PollMaxSize
	}

	// set read polling duration and deadline
	var tPolling time.Duration
	var tDeadline time.Time

	tPolling = time.Duration(c.PollTimeout * float64(time.Second))
	if tPolling <= 0 {
		tPolling = time.Duration(comm.POLL_TIMEOUT * float64(time.Second))
	}
	if timeout > 0 {
		tDeadline = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	}

	c.serialPort.SetReadTimeout(tPolling)

	var data []byte

	b := make([]byte, nRead)
	for {
		n, err := c.serialPort.Read(b)
		if err != nil {
			if comm.IsClosedError(err) {
				c.closeEvent.Store(true)
				comm.LogMsg(c.CommLog, "PORT_CLOSED -- %v", err)
				go c.Close()
				return nil, nil, comm.ErrClosed
			}
			comm.LogMsg(c.CommLog, "RECV_ERROR -- %v", err)
			return nil, nil, fmt.Errorf("%w, %v", comm.ErrRecv, err)
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

		if c.breakRecvEvent.Load() {
			return nil, nil, comm.ErrBreak
		}
		if timeout > 0 && time.Now().After(tDeadline) {
			return nil, nil, comm.ErrTimeout
		}
	}

	comm.LogRx(c.CommLog, data, nil)
	return data, nil, nil
}

/////////////////////////////////////////////////////

// Listener represents a serial listener that handles incoming connections
type Listener struct {
	// Context containing common attributes such as logging and events.
	*comm.Context

	// serial port connection.
	serialConn *Connection

	// ConnectionHandler defines the function to handle incoming connections.
	ConnectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool

	// muState defines mutex for state change operations (start/stop).
	muState sync.Mutex
}

// NewListener creates a new Listener.
func NewListener(uri string, commlog *logging.Logger, opts dictx.Dict) (*Listener, error) {
	conn, err := NewConnection(uri, commlog, opts)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Context:    comm.NewContext(commlog, opts),
		serialConn: conn,
	}, nil
}

// String returns a string representation of the Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<SerialListener: %s>", l.serialConn.uri)
}

// Uri returns the URI of the listener
func (l *Listener) Uri() string {
	return l.serialConn.uri
}

// Type returns the type of the listener as inferred from the Uri.
func (l *Listener) Type() string {
	return "serial"
}

// SerialPort returns the underlying serial port object
func (l *Listener) SerialPort() serial.Port {
	return l.serialConn.SerialPort()
}

// SetConnHandler sets a callback function to handle connections.
func (l *Listener) SetConnHandler(handler func(comm.Connection)) {
	l.ConnectionHandler = handler
}

// IsActive checks if the listener is currently active.
func (l *Listener) IsActive() bool {
	return l.isActive.Load()
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

	if err := l.serialConn.Open(-1); err != nil {
		return err
	}
	l.serialConn.parent = l

	l.isActive.Store(true)
	defer func() {
		l.serialConn.parent = nil
		l.serialConn.Close()
		l.isActive.Store(false)
	}()

	// run connection handler
	l.ConnectionHandler(l.serialConn)

	return nil
}

// Stop gracefully shuts down the listener.
func (l *Listener) Stop() {
	// do nothing if already stopped
	if !l.isActive.Load() {
		return
	}

	l.serialConn.parent = nil
	l.serialConn.Close()
}
