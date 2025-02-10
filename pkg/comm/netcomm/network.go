// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package netcomm

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/netutil"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// ParseUri parses a network URI.
//
//	The expected URI format is `<network>@<host>:<port>`
//
//	<network>  {tcp|tcp4|tcp6|udp|udp4|udp6}
//	<host>     The host FQDN or IP address.
//	<port>     The port number. can be number or protocol name.
//
//	 -- see net.dial for full details.
//		(referance:  https://pkg.go.dev/net#Dial)
//
//	example:
//	server
//	   - tcp@0.0.0.0:1234
//	   - tcp6@[::1]:1234
//	   - udp@0.0.0.0:http
//	client
//	   - tcp@1.2.3.4:1234
//	   - tcp4@1.2.3.4:1234
//	   - tcp6@[2001:db8::1]:http
//	   - udp@1.2.3.4:1234
//
// Returns the parsed network type, address, or error for invalid URI format.
func ParseUri(uri string) (string, string, error) {
	parts := strings.SplitN(uri, "@", 2)
	if len(parts) < 2 {
		return "", "", comm.ErrUri
	}

	network := strings.ToLower(parts[0])
	address := parts[1]

	switch network {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		if strings.Count(address, ":") > 0 {
			return network, address, nil
		}
	}

	return "", "", comm.ErrUri
}

// GetAddr returns (net.Addr) type for network and address.
func GetAddr(network, address string) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return net.ResolveTCPAddr(network, address)
	case "udp", "udp4", "udp6":
		return net.ResolveUDPAddr(network, address)
	}
	return nil, errors.New("invalid network type")
}

/////////////////////////////////////////////////////

// Connection represents a network connection with event support and logging.
type Connection struct {
	// uri specifies the resource identifier.
	uri string
	// The network type (e.g., tcp, udp).
	network string
	// The network address (host:port).
	address string
	// The underlying network socket/packet connection.
	netConn any // (net.Conn|net.PacketConn)

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

	// CommLog is the logger instance for communication data logging.
	CommLog *logging.Logger

	// PollConfig defines the read polling.
	PollConfig *comm.PollingConfig
	// KeepaliveConfig defines the keep-alive probes for TCP connections.
	KeepaliveConfig *comm.KeepaliveConfig
	// TlsConfig defines the TLS attributes for TCP connections.
	TlsConfig *comm.TlsConfig
}

// NewConnection creates and initializes a new Connection for the given URI.
//
// The parsed options are:
//   - Polling Options: detailed in [comm.ParsePollingConfig]
//   - Keepalive Options: detailed in [comm.ParseKeepaliveConfig]
//   - TLS Options: detailed in [comm.ParseTlsConfig]
func NewConnection(uri string, commlog *logging.Logger, opts dictx.Dict) (*Connection, error) {
	uri = strings.TrimSpace(uri)
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		uri:     uri,
		network: network,
		address: address,
		CommLog: commlog,
	}

	// set polling options
	c.PollConfig, err = comm.ParsePollingConfig(opts)
	if err != nil {
		return nil, err
	}

	// set keep-alive options
	c.KeepaliveConfig, err = comm.ParseKeepaliveConfig(opts)
	if err != nil {
		return nil, err
	}

	// set TLS options
	// only TCP supported now. DTLS requires 3rd party lib
	if strings.HasPrefix(network, "tcp") {
		c.TlsConfig, err = comm.ParseTlsConfig(opts)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// String returns a string representation of the Connection.
func (c *Connection) String() string {
	return fmt.Sprintf("<NetConnection: %s>", c.uri)
}

// Uri returns the URI of the connection
func (c *Connection) Uri() string {
	return c.uri
}

// Type returns the type of the connection as inferred from the Uri.
func (c *Connection) Type() string {
	return c.network
}

// NetConn returns the net connection (net.Conn|net.PacketConn).
func (c *Connection) NetConn() any {
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
	if c.KeepaliveConfig.Interval >= 0 {
		dialer.KeepAlive = time.Duration(
			c.KeepaliveConfig.Interval * int(time.Second))
	}
	if timeout > 0 {
		dialer.Timeout = time.Duration(timeout * float64(time.Second))
	}

	conn, err := dialer.Dial(c.network, c.address)
	if err != nil {
		comm.LogMsg(c.CommLog, "CONNECT_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	// set tls config for connection
	if c.TlsConfig != nil {
		conn = tls.Client(conn, c.TlsConfig)
		comm.LogMsg(c.CommLog, "CONNECTED TLS -- %s", c.uri)
	} else {
		comm.LogMsg(c.CommLog, "CONNECTED -- %s", c.uri)
	}
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
	if conn, ok := c.netConn.(net.Conn); ok {
		conn.Close()
	} else if conn, ok := c.netConn.(net.PacketConn); ok {
		conn.Close()
	}

	c.wgClose.Wait()
	comm.LogMsg(c.CommLog, "DISCONNECTED -- %s", c.uri)
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

	var err error
	var n int

	if conn, ok := c.netConn.(net.PacketConn); ok && c.parent != nil {
		if a, ok := addr.(net.Addr); ok {
			comm.LogTx(c.CommLog, data, a)
			if timeout > 0 {
				conn.SetWriteDeadline(time.Now().Add(
					time.Duration(timeout * float64(time.Second))))
			}
			n, err = conn.WriteTo(data, a)
		} else {
			if addr == nil {
				return errors.New("empty address")
			} else {
				return errors.New("invalid address type")
			}
		}
	} else if conn, ok := c.netConn.(net.Conn); ok {
		comm.LogTx(c.CommLog, data, nil)
		if timeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(
				time.Duration(timeout * float64(time.Second))))
		}
		n, err = conn.Write(data)
	} else {
		return errors.New("invalid connection type")
	}
	if err == nil && n != len(data) {
		err = errors.New("partial data sent")
	}

	if err != nil {
		if comm.IsClosedError(err) || comm.IsTLSError(err) {
			c.closeEvent.Store(true)
			comm.LogMsg(c.CommLog, "CONN_CLOSED -- %v", err)
			go c.Close()
			return comm.ErrClosed
		}
		comm.LogMsg(c.CommLog, "SEND_ERROR -- %v", err)
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

	// no polling for packet session-less connections
	if _, ok := c.netConn.(net.PacketConn); ok {
		if timeout > 0 {
			tDeadline = time.Now().Add(
				time.Duration(timeout * float64(time.Second)))
		}
	} else {
		tPolling = time.Duration(
			c.PollConfig.Timeout * float64(time.Second))
		if timeout > 0 {
			tDeadline = time.Now().Add(
				time.Duration(timeout * float64(time.Second)))
		}
	}

	var err error
	var data []byte
	var n int
	var addr net.Addr

	b := make([]byte, nRead)
	for {
		if conn, ok := c.netConn.(net.PacketConn); ok && c.parent != nil {
			if timeout > 0 {
				conn.SetReadDeadline(tDeadline)
			}
			n, addr, err = conn.ReadFrom(b)
		} else if conn, ok := c.netConn.(net.Conn); ok {
			conn.SetReadDeadline(time.Now().Add(tPolling))
			n, err = conn.Read(b)
		} else {
			return nil, nil, errors.New("invalid connection type")
		}
		if err != nil {
			if comm.IsClosedError(err) || comm.IsTLSError(err) {
				c.closeEvent.Store(true)
				comm.LogMsg(c.CommLog, "CONN_CLOSED -- %v", err)
				go c.Close()
				return nil, nil, comm.ErrClosed
			}
			if _, ok := err.(net.Error); !ok || !err.(net.Error).Timeout() {
				comm.LogMsg(c.CommLog, "RECV_ERROR -- %v", err)
				return nil, nil, fmt.Errorf("%w, %v", comm.ErrRecv, err)
			}
		}

		if n > 0 {
			data = append(data, b[:n]...)
			if _, ok := c.netConn.(net.PacketConn); ok {
				break
			}
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

	comm.LogRx(c.CommLog, data, addr)
	return data, addr, nil
}

/////////////////////////////////////////////////////

// Listener represents a network listener that handles incoming connections
// with a custom connection handler.
type Listener struct {
	// uri specifies the resource identifier.
	uri string
	// The network type (e.g., tcp, udp).
	network string
	// The network address to listen on (host:port).
	address string
	// The underlying network listener/Packet-connection.
	netListener any // (net.Listener|net.PacketConn)

	// ConnectionHandler defines the function to handle incoming connections.
	ConnectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool
	// stopEvent signals a stop operation.
	stopEvent atomic.Bool

	// muState defines mutex for state change operations (start/stop).
	muState sync.Mutex

	// CommLog is the logger instance for communication data logging.
	CommLog *logging.Logger

	// PollConfig defines the read polling.
	PollConfig *comm.PollingConfig
	// LimiterConfig defines the limits for TCP connections.
	LimiterConfig *comm.LimiterConfig
	// KeepaliveConfig defines the keep-alive probes for TCP connections.
	KeepaliveConfig *comm.KeepaliveConfig
	// TlsConfig defines the TLS attributes for TCP connections.
	TlsConfig *comm.TlsConfig
}

// NewListener creates a new network Listener.
//
// The parsed options are:
//   - Polling Options: detailed in [comm.ParsePollingConfig]
//   - Limiter Options: detailed in [comm.ParseLimiterConfig]
//   - Keepalive Options: detailed in [comm.ParseKeepaliveConfig]
//   - TLS Options: detailed in [comm.ParseTlsConfig]
func NewListener(uri string, commlog *logging.Logger, opts dictx.Dict) (*Listener, error) {
	uri = strings.TrimSpace(uri)
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		uri:     uri,
		network: network,
		address: address,
		CommLog: commlog,
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

	// set keep-alive options
	l.KeepaliveConfig, err = comm.ParseKeepaliveConfig(opts)
	if err != nil {
		return nil, err
	}

	// set TLS config for connection
	// only TCP supported now. DTLS requires 3rd party lib
	if strings.HasPrefix(network, "tcp") {
		l.TlsConfig, err = comm.ParseTlsConfig(opts)
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

// String returns a string representation of the Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<NetListener: %s>", l.uri)
}

// Uri returns the URI of the listener
func (l *Listener) Uri() string {
	return l.uri
}

// Type returns the type of the listener as inferred from the Uri.
func (l *Listener) Type() string {
	return l.network
}

// NetListener returns the net listener instance (net.Listener|net.PacketConn).
func (l *Listener) NetListener() any {
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
	if l.KeepaliveConfig.Interval >= 0 {
		lConfig.KeepAlive = time.Duration(
			l.KeepaliveConfig.Interval * int(time.Second))
	}

	// net listener
	netListener, err := lConfig.Listen(
		context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	// set connection limit
	if l.LimiterConfig.SimultaneousConn > 0 {
		netListener = netutil.LimitListener(
			netListener, l.LimiterConfig.SimultaneousConn)
	}
	// set tls config for listener
	if l.TlsConfig != nil {
		netListener = tls.NewListener(netListener, l.TlsConfig)
		comm.LogMsg(l.CommLog, "LISTENING TLS -- %s", l.uri)
	} else {
		comm.LogMsg(l.CommLog, "LISTENING -- %s", l.uri)
	}
	l.netListener = netListener

	var wg sync.WaitGroup

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		netListener.Close()
		// wait all connections handlers termination
		wg.Wait()
		comm.LogMsg(l.CommLog, "CLOSED -- %s", l.uri)
		l.isActive.Store(false)
	}()

	for !l.stopEvent.Load() {
		// wait for new connection
		new_conn, err := netListener.Accept()
		if err != nil {
			if comm.IsClosedError(err) {
				break
			} else {
				comm.LogMsg(l.CommLog, "CONN_ERROR -- %v", err)
				continue
			}
		}

		// handle new connection
		wg.Add(1)
		go func(conn net.Conn) {
			address := fmt.Sprint(conn.RemoteAddr())
			uri := fmt.Sprintf("%s@%s", l.network, address)
			c := &Connection{
				uri:             uri,
				network:         l.network,
				address:         address,
				netConn:         conn,
				PollConfig:      l.PollConfig,
				KeepaliveConfig: l.KeepaliveConfig,
			}
			if l.CommLog != nil {
				c.CommLog = l.CommLog.SubLogger(uri)
			}
			c.parent = l
			c.isOpened.Store(true)
			comm.LogMsg(c.CommLog, "CONNECTED")

			defer func() {
				conn.Close()
				comm.LogMsg(c.CommLog, "DISCONNECTED")
				wg.Done()
			}()

			l.ConnectionHandler(c)
		}(new_conn)
	}

	return nil
}

func (l *Listener) startPacketConn() error {
	lConfig := net.ListenConfig{
		KeepAlive: -1, // disabled
	}

	// packet connection
	packetConn, err := lConfig.ListenPacket(
		context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	l.netListener = packetConn

	comm.LogMsg(l.CommLog, "LISTENING -- %s", l.uri)

	c := &Connection{
		uri:        l.uri,
		network:    l.network,
		address:    l.address,
		netConn:    packetConn,
		PollConfig: l.PollConfig,
	}
	c.parent = l
	c.isOpened.Store(true)

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		c.parent = nil
		c.Close()
		l.isActive.Store(false)
	}()

	// run connection handler
	l.ConnectionHandler(c)

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

	if strings.HasPrefix(l.network, "udp") {
		return l.startPacketConn()
	}
	return l.startListener()
}

// Stop gracefully shuts down the listener.
func (l *Listener) Stop() {
	l.stopEvent.Store(true)

	// do nothing if already stopped
	if !l.isActive.Load() {
		return
	}

	if v, ok := l.netListener.(net.PacketConn); ok {
		v.Close()
	} else if v, ok := l.netListener.(net.Listener); ok {
		v.Close()
	}
}
