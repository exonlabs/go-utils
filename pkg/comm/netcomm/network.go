// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package netcomm

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/netutil"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/comm"
	"github.com/exonlabs/go-utils/pkg/logging"
)

// ParseUri parses a network URI into network type and address.
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
// Returns the parsed network type, address, and error for invalid URI format.
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

// GetTlsConfig returns tls configuration from parsed options.
// The parsed options are:
//   - tls_enable: (bool) enable/disable TLS, default disabled.
//   - tls_mutual_auth: (bool) enable/disable mutual TLS auth, default disabled.
//   - tls_server_name: (string) server name to use by client TLS session.
//   - tls_min_version: (float64) min TLS version to use. default TLS v1.2
//   - tls_max_version: (float64) max TLS version to use. default TLS v1.3
//   - tls_ca_certs: (string) comma separated list of CA certs to use.
//     cert values could be file paths to load or cert content in PEM format.
//   - tls_local_cert: (string) cert to use for TLS session.
//     cert could be file path to load or cert content in PEM format.
//   - tls_local_key: (string) private key to use for TLS session.
//     key could be file path to load or key content in PEM format.
func GetTlsConfig(opts dictx.Dict) (*tls.Config, error) {
	if !dictx.Fetch(opts, "tls_enable", false) {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		ClientAuth:         tls.RequireAnyClientCert,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	if dictx.Fetch(opts, "tls_mutual_auth", false) {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.InsecureSkipVerify = false
	}
	if v := strings.TrimSpace(dictx.Fetch(opts, "tls_server_name", "")); v != "" {
		tlsConfig.ServerName = v
	}

	if v := dictx.GetFloat(opts, "tls_min_version", 0); v > 0 {
		switch v {
		case 1.0:
			tlsConfig.MinVersion = tls.VersionTLS10
		case 1.1:
			tlsConfig.MinVersion = tls.VersionTLS11
		case 1.2:
			tlsConfig.MinVersion = tls.VersionTLS12
		case 1.3:
			tlsConfig.MinVersion = tls.VersionTLS13
		default:
			return nil, errors.New("invalid tls_min_version value")
		}
	}
	if v := dictx.GetFloat(opts, "tls_max_version", 0); v > 0 {
		switch v {
		case 1.0:
			tlsConfig.MaxVersion = tls.VersionTLS10
		case 1.1:
			tlsConfig.MaxVersion = tls.VersionTLS11
		case 1.2:
			tlsConfig.MaxVersion = tls.VersionTLS12
		case 1.3:
			tlsConfig.MaxVersion = tls.VersionTLS13
		default:
			return nil, errors.New("invalid tls_max_version value")
		}
	}

	if v := strings.TrimSpace(dictx.Fetch(opts, "tls_ca_certs", "")); v != "" {
		certPool := x509.NewCertPool()

		for _, crtStr := range strings.Split(v, ",") {
			crtStr = strings.TrimSpace(crtStr)

			var crtByte []byte
			var err error
			if strings.HasPrefix(crtStr, "-----BEGIN") {
				crtByte = []byte(crtStr)
			} else {
				crtByte, err = os.ReadFile(crtStr)
				if err != nil {
					return nil, fmt.Errorf(
						"error loading tls_ca_certs - %v", err)
				}
			}
			if !certPool.AppendCertsFromPEM(crtByte) {
				return nil, errors.New("invalid tls_ca_certs value")
			}
		}

		tlsConfig.RootCAs = certPool
		tlsConfig.ClientCAs = certPool
	}

	if crtStr := strings.TrimSpace(
		dictx.Fetch(opts, "tls_local_cert", "")); crtStr != "" {
		keyStr := strings.TrimSpace(dictx.Fetch(opts, "tls_local_key", ""))
		if keyStr == "" {
			return nil, errors.New("empty tls_local_key value")
		}

		var cert tls.Certificate
		var err error
		if strings.HasPrefix(crtStr, "-----BEGIN") &&
			strings.HasPrefix(keyStr, "-----BEGIN") {
			cert, err = tls.X509KeyPair([]byte(crtStr), []byte(keyStr))
		} else if strings.HasPrefix(crtStr, "-----BEGIN") ||
			strings.HasPrefix(keyStr, "-----BEGIN") {
			return nil, errors.New(
				"both options tls_local_cert and tls_local_key should be " +
					"file paths or PEM formatted contents for cert and key")
		} else {
			cert, err = tls.LoadX509KeyPair(crtStr, keyStr)
		}
		if err != nil {
			return nil, fmt.Errorf(
				"error loading tls_local_cert, tls_local_key - %v", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// IsTLSError checks if the error is related to TLS error
func IsTLSError(err error) bool {
	switch {
	case strings.Contains(err.Error(), "tls: "):
		return true
	default:
		return false
	}
}

/////////////////////////////////////////////////////

// Connection represents a network connection with event support and logging.
type Connection struct {
	// Context containing common attributes and functions.
	*comm.Context

	// The network type (e.g., tcp, udp).
	network string
	// The network address (host:port).
	address string
	// TlsConfig defines the TLS attributes for TCP connections.
	tlsConfig *tls.Config

	// The underlying network socket/packet connection.
	netConn any // (net.Conn|net.PacketConn)

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
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		Context: comm.NewContext(uri, log, opts),
		network: network,
		address: address,
	}

	// set TLS config for connection
	// only TCP supported now. DTLS requires 3rd party lib
	if strings.HasPrefix(network, "tcp") {
		c.tlsConfig, err = GetTlsConfig(opts)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// String returns a string representation of the Connection.
func (c *Connection) String() string {
	return fmt.Sprintf("<NetConnection: %s>", c.Uri())
}

// NetConn returns the net connection instance (net.Conn|net.PacketConn).
func (c *Connection) NetConn() any {
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
// The parsed options are:
//   - keepalive_interval: (float64) the keep-alive interval in seconds.
//     use 0 to enable keep-alive probes with OS defined values. (default is 0)
//     use -1 to disable keep-alive probes.
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
		KeepAlive: 0, // enabled by default with defaults
	}
	if v := dictx.GetFloat(c.Options, "keepalive_interval", 0); v >= 0 {
		dialer.KeepAlive = time.Duration(v * float64(time.Second))
	}
	if timeout > 0 {
		dialer.Timeout = time.Duration(timeout * float64(time.Second))
	}

	conn, err := dialer.Dial(c.network, c.address)
	if err != nil {
		c.LogMsg("CONNECT_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	// set tls config for connection
	if c.tlsConfig != nil {
		conn = tls.Client(conn, c.tlsConfig)
		c.LogMsg("CONNECTED TLS -- %s", c.Uri())
	} else {
		c.LogMsg("CONNECTED -- %s", c.Uri())
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

	c.sMutex.Lock()
	defer c.sMutex.Unlock()

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

	var err error
	var n int

	if conn, ok := c.netConn.(net.PacketConn); ok && c.parent != nil {
		if addr == nil {
			return errors.New("empty address")
		}
		if a, ok := addr.(net.Addr); ok {
			c.LogTx(data, a)
			if timeout > 0 {
				conn.SetWriteDeadline(time.Now().Add(
					time.Duration(timeout * float64(time.Second))))
			}
			n, err = conn.WriteTo(data, a)
		} else {
			return errors.New("invalid address type")
		}
	} else if conn, ok := c.netConn.(net.Conn); ok {
		c.LogTx(data, nil)
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
		if comm.IsClosedError(err) || IsTLSError(err) {
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

	// set read polling and tbreak
	var tPoll time.Duration
	var tBreak time.Time
	if _, ok := c.netConn.(net.PacketConn); ok {
		if timeout > 0 {
			tPoll = time.Duration(timeout * float64(time.Second))
			tBreak = time.Now().Add(tPoll)
		}
	} else {
		if c.PollTimeout > 0 {
			tPoll = time.Duration(c.PollTimeout * float64(time.Second))
		} else {
			tPoll = time.Duration(comm.POLL_TIMEOUT * float64(time.Second))
		}
		if timeout > 0 {
			tBreak = time.Now().Add(
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
				conn.SetReadDeadline(tBreak)
			}
			n, addr, err = conn.ReadFrom(b)
		} else if conn, ok := c.netConn.(net.Conn); ok {
			conn.SetReadDeadline(time.Now().Add(tPoll))
			n, err = conn.Read(b)
		} else {
			return nil, nil, errors.New("invalid connection type")
		}
		if err != nil {
			if comm.IsClosedError(err) || IsTLSError(err) {
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
			if _, ok := c.netConn.(net.PacketConn); ok {
				break
			}
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

	c.LogRx(data, addr)
	return data, addr, nil
}

/////////////////////////////////////////////////////

// Listener represents a network listener that handles incoming connections
// with a custom connection handler.
type Listener struct {
	// Context containing common attributes such as logging and events.
	*comm.Context

	// The network type (e.g., tcp, udp).
	network string
	// The network address to listen on (host:port).
	address string
	// TlsConfig defines the TLS attributes for TCP connections.
	tlsConfig *tls.Config

	// The underlying network listener/Packet connection.
	netListener any // (net.Listener|net.PacketConn)

	// The handler function to be called when a new connection is accepted.
	connectionHandler func(comm.Connection)

	// isActive represents the listener status, started or stopped.
	isActive atomic.Bool
	// stopEvent signals a stop operation.
	stopEvent atomic.Bool

	// sMutex defines mutex for state change operations (start/stop).
	sMutex sync.Mutex
}

// NewListener creates a new network Listener.
// The parsed options are:
//   - connections_limit: (int) the limit on number of concurrent connections.
//     use 0 to disable connections limit.
//   - keepalive_interval: (float64) the keep-alive interval in seconds.
//     use 0 to enable keep-alive probes with OS defined values.
//     use -1 to disable keep-alive probes. (default is -1)
func NewListener(uri string, log *logging.Logger, opts dictx.Dict) (*Listener, error) {
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		Context: comm.NewContext(uri, log, opts),
		network: network,
		address: address,
	}

	// set TLS config for connection
	// only TCP supported now. DTLS requires 3rd party lib
	if strings.HasPrefix(network, "tcp") {
		l.tlsConfig, err = GetTlsConfig(opts)
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

// String returns a string representation of the Listener.
func (l *Listener) String() string {
	return fmt.Sprintf("<NetListener: %s>", l.Uri())
}

// NetListener returns the net listener instance (net.Listener|net.PacketConn).
func (l *Listener) NetListener() any {
	return l.netListener
}

// ConnectionHandler sets a callback function to handle connections.
func (l *Listener) ConnectionHandler(h func(comm.Connection)) {
	l.connectionHandler = h
}

// IsActive checks if the listener is currently active.
func (l *Listener) IsActive() bool {
	return l.isActive.Load() && !l.stopEvent.Load()
}

func (l *Listener) startListener() error {
	cfg := net.ListenConfig{
		KeepAlive: -1, // disabled by default
	}
	if v := dictx.GetFloat(l.Options, "keepalive_interval", -1); v >= 0 {
		cfg.KeepAlive = time.Duration(v * float64(time.Second))
	}

	// listener instance
	netListener, err := cfg.Listen(context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	// set connection limit (if configured)
	if v := dictx.GetInt(l.Options, "connections_limit", 0); v > 0 {
		netListener = netutil.LimitListener(netListener, v)
	}
	// set tls config for listener
	if l.tlsConfig != nil {
		netListener = tls.NewListener(netListener, l.tlsConfig)
		l.LogMsg("LISTENING TLS -- %s", l.Uri())
	} else {
		l.LogMsg("LISTENING -- %s", l.Uri())
	}
	l.netListener = netListener

	var waitGrp sync.WaitGroup

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		netListener.Close()
		// wait all connections handlers termination
		waitGrp.Wait()
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
			nc, err := NewConnection(uri, nil, l.Options)
			if err != nil {
				l.LogMsg("CONN_ERROR -- %v", err)
				netConn.Close()
				return
			}
			if l.CommLog != nil {
				nc.CommLog = l.CommLog.SubLogger(fmt.Sprintf("(%s) ", uri))
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

func (l *Listener) startPacketConn() error {
	var cfg net.ListenConfig

	// packet connection instance
	packetConn, err := cfg.ListenPacket(
		context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	l.netListener = packetConn

	l.LogMsg("LISTENING -- %s", l.Uri())

	nc := &Connection{
		Context: comm.NewContext(l.Uri(), l.CommLog, l.Options),
		network: l.network,
		address: l.address,
		netConn: packetConn,
		parent:  l,
	}
	nc.isOpened.Store(true)

	l.stopEvent.Store(false)
	l.isActive.Store(true)
	defer func() {
		l.stopEvent.Store(true)
		nc.parent = nil
		nc.Close()
		l.isActive.Store(false)
	}()

	// run connection handler
	l.connectionHandler(nc)

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
