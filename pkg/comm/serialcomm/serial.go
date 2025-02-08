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

// ParseUri parses a serial URI into serial params.
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
func ParseUri(uri string) (string, serial.Mode, error) {
	parts := strings.SplitN(uri, "@", 2)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "serial" {
		return "", serial.Mode{}, comm.ErrUri
	}
	// parts after @
	parts = strings.Split(parts[1], ":")
	if len(parts) < 3 {
		return "", serial.Mode{}, comm.ErrUri
	}

	port, strBaud, strMode := parts[0], parts[1], parts[2]
	if len(strMode) != 3 {
		return "", serial.Mode{}, comm.ErrUri
	}

	var err error
	mode := serial.Mode{}

	if mode.BaudRate, err = strconv.Atoi(strBaud); err != nil {
		return "", serial.Mode{}, comm.ErrUri
	}
	if mode.DataBits, err = strconv.Atoi(string(strMode[0])); err != nil {
		return "", serial.Mode{}, comm.ErrUri
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
		return "", serial.Mode{}, comm.ErrUri
	}
	switch string(strMode[2]) {
	case "1":
		mode.StopBits = serial.OneStopBit
	case "2":
		mode.StopBits = serial.TwoStopBits
	default:
		return "", serial.Mode{}, comm.ErrUri
	}

	return port, mode, nil
}

/////////////////////////////////////////////////////

// Connection represents a serial connection with event support and logging.
type Connection struct {
	// Context containing common attributes and functions.
	*comm.Context

	// The serial port identifier (e.g., COM1, /dev/ttyS0).
	port string
	// Configuration of serial communication parameters.
	mode serial.Mode
	// The underlying serial port connection.
	serialPort serial.Port

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
// The URI specifies the network type and address.
func NewConnection(uri string, log *logging.Logger, opts dictx.Dict) (*Connection, error) {
	port, mode, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	ctx := comm.NewContext(uri, log, opts)

	// set dynamic polling time relative to baudrate.
	// set to 10 bytes duration, where actual 1 byte = 10 bits
	ctx.PollTimeout = gx.Max(
		ctx.PollTimeout, 0.02, 20.0/float64(mode.BaudRate))

	return &Connection{
		Context: ctx,
		port:    port,
		mode:    mode,
	}, nil
}

// String returns a string representation of the Connection.
func (sc *Connection) String() string {
	return fmt.Sprintf("<SerialConnection: %s>", sc.Uri())
}

// SerialPort returns the underlying serial port object
func (sc *Connection) SerialPort() serial.Port {
	return sc.serialPort
}

// Parent retrieves the parent Listener, if any, associated with the Connection.
func (sc *Connection) Parent() comm.Listener {
	return sc.parent
}

// IsOpened checks if the Connection is currently open.
func (sc *Connection) IsOpened() bool {
	return sc.isOpened.Load() && !sc.closeEvent.Load()
}

// Open opens the Connection, establishing communication over the serial port.
func (sc *Connection) Open(timeout float64) error {
	// take no action if managed by parent listener
	if sc.parent != nil {
		return nil
	}

	sc.sMutex.Lock()
	defer sc.sMutex.Unlock()

	// do nothing if already opened
	if sc.isOpened.Load() {
		return nil
	}

	com, err := serial.Open(sc.port, &sc.mode)
	if err != nil {
		sc.LogMsg("OPEN_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	sc.serialPort = com
	sc.serialPort.ResetInputBuffer()
	sc.serialPort.ResetOutputBuffer()

	sc.LogMsg("OPENED -- %s", sc.Uri())
	sc.closeEvent.Store(false)
	sc.isOpened.Store(true)
	return nil
}

// Close closes the Connection and cleans up resources.
func (sc *Connection) Close() {
	// take no action if managed by parent listener
	if sc.parent != nil {
		return
	}

	sc.closeEvent.Store(true)

	sc.sMutex.Lock()
	defer sc.sMutex.Unlock()

	// do nothing if already closed
	if !sc.isOpened.Load() {
		return
	}

	// close port clean buffers
	sc.serialPort.Close()
	sc.serialPort.ResetInputBuffer()
	sc.serialPort.ResetOutputBuffer()

	sc.rwWaitGrp.Wait()
	sc.LogMsg("CLOSED -- %s", sc.Uri())
	sc.isOpened.Store(false)
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
	c.mode = newMode

	// set dynamic polling time relative to baudrate.
	// set to 10 bytes duration, where actual 1 byte = 10 bits
	c.PollTimeout = gx.Max(0.02, 20.0/float64(newMode.BaudRate))

	// apply new mode if port is already opened
	if c.isOpened.Load() {
		c.LogMsg("SETMODE -- %s", newUri)
		if err := c.serialPort.SetMode(newMode); err != nil {
			return fmt.Errorf("%w, %v", comm.ErrConnection, err)
		}
		c.serialPort.ResetInputBuffer()
		c.serialPort.ResetOutputBuffer()
	}

	return nil
}

// Cancel interrupts the ongoing communication for this Connection.
func (sc *Connection) Cancel() {
	sc.breakReadEvent.Store(true)
}

// Cancel interrupts the ongoing sending operation for this Connection.
func (sc *Connection) CancelSend() {
	// do nothing, not available for serial port
}

// Cancel interrupts the ongoing receiving operation for this Connection.
func (sc *Connection) CancelRecv() {
	sc.breakReadEvent.Store(true)
}

// Send transmits data over the connection, with a specified timeout.
func (sc *Connection) Send(data []byte, timeout float64) error {
	return sc.SendTo(data, nil, timeout)
}

// SendTo transmits data to addr over the connection, with a specified timeout.
func (sc *Connection) SendTo(data []byte, _ any, timeout float64) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	// Acquire write lock
	sc.wMutex.Lock()
	defer sc.wMutex.Unlock()

	// Check connection state after acquiring the lock
	if sc.closeEvent.Load() || !sc.isOpened.Load() {
		return comm.ErrClosed
	}

	sc.rwWaitGrp.Add(1)
	defer sc.rwWaitGrp.Done()

	sc.LogTx(data, nil)
	n, err := sc.serialPort.Write(data)
	if err == nil {
		// Ensure all data is flushed from the OS buffer
		err = sc.serialPort.Drain()
		if err == nil && n != len(data) {
			err = errors.New("partial data written to port")
		}
	}

	if err != nil {
		if comm.IsClosedError(err) {
			sc.closeEvent.Store(true)
			sc.LogMsg("PORT_CLOSED -- %v", err)
			go sc.Close()
			return comm.ErrClosed
		}
		sc.LogMsg("SEND_ERROR -- %v", err)
		sc.serialPort.ResetOutputBuffer()
		return fmt.Errorf("%w, %v", comm.ErrWrite, err)
	}

	return nil
}

// Recv waits for incoming data over the connection until a timeout
// or interrupt event occurs. Setting timeout=0 will wait indefinitely.
func (sc *Connection) Recv(timeout float64) ([]byte, error) {
	b, _, err := sc.RecvFrom(timeout)
	return b, err
}

// Recv waits for incoming data from addr over the connection until a timeout
// or interrupt event occurs. Setting timeout=0 will wait indefinitely.
func (sc *Connection) RecvFrom(timeout float64) ([]byte, any, error) {
	// Acquire read lock
	sc.rMutex.Lock()
	defer sc.rMutex.Unlock()

	// Check connection state after acquiring the lock
	if sc.closeEvent.Load() || !sc.isOpened.Load() {
		return nil, nil, comm.ErrClosed
	}

	sc.rwWaitGrp.Add(1)
	defer sc.rwWaitGrp.Done()

	sc.breakReadEvent.Store(false)

	// determine read buffer size and polling timeout
	nRead := sc.PollChunkSize
	if sc.PollMaxSize > 0 {
		nRead = sc.PollMaxSize
	}

	tPoll := time.Duration(sc.PollTimeout * float64(time.Second))
	if tPoll <= 0 {
		tPoll = time.Duration(comm.POLL_TIMEOUT * float64(time.Second))
	}

	var tBreak time.Time
	if timeout > 0 {
		tBreak = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	}
	sc.serialPort.SetReadTimeout(tPoll)

	var data []byte

	b := make([]byte, nRead)
	for {
		n, err := sc.serialPort.Read(b)
		if err != nil {
			if comm.IsClosedError(err) {
				sc.closeEvent.Store(true)
				sc.LogMsg("PORT_CLOSED -- %v", err)
				go sc.Close()
				return nil, nil, comm.ErrClosed
			}
			sc.LogMsg("RECV_ERROR -- %v", err)
			return nil, nil, fmt.Errorf("%w, %v", comm.ErrRead, err)
		}

		if n > 0 {
			data = append(data, b[:n]...)
			if sc.PollMaxSize > 0 {
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

		if sc.breakReadEvent.Load() {
			return nil, nil, comm.ErrBreak
		}
		if timeout > 0 && time.Now().After(tBreak) {
			return nil, nil, comm.ErrTimeout
		}
	}

	sc.LogRx(data, nil)
	return data, nil, nil
}

/////////////////////////////////////////////////////

// Listener represents a serial listener that handles incoming connections
type Listener struct {
	// Context containing common attributes such as logging and events.
	*comm.Context

	// serial port connection.
	serialConn *Connection

	// The handler function to be called for new connection.
	connectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool

	// sMutex defines mutex for state change operations (start/stop).
	sMutex sync.Mutex
}

// NewListener creates a new Listener for the specified URI, with
// optional logging and connection limit.
func NewListener(uri string, log *logging.Logger, opts dictx.Dict) (*Listener, error) {
	conn, err := NewConnection(uri, log, opts)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Context:    comm.NewContext(uri, log, opts),
		serialConn: conn,
	}, nil
}

// String returns a string representation of the Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<SerialListener: %s>", l.Uri())
}

// Port returns the underlying serial port object
func (l *Listener) SerialPort() serial.Port {
	return l.serialConn.SerialPort()
}

// SetConnHandler sets a callback function to handle connections.
func (l *Listener) SetConnHandler(h func(comm.Connection)) {
	l.connectionHandler = h
}

// IsActive checks if the listener is currently active.
func (l *Listener) IsActive() bool {
	return l.isActive.Load()
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

	if err := l.serialConn.Open(0); err != nil {
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
	l.connectionHandler(l.serialConn)

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
