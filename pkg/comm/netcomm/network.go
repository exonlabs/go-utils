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

/////////////////////////////////////////////////////

// Connection represents a network connection with event support and logging.
type Connection struct {
	// Context containing common attributes and functions.
	*comm.Context

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

	// TlsConfig defines the TLS attributes for TCP connections.
	tlsConfig *tls.Config
}

// NewConnection creates and initializes a new Connection for the given URI.
func NewConnection(uri string, commlog *logging.Logger, opts dictx.Dict) (*Connection, error) {
	uri = strings.TrimSpace(uri)
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		Context: comm.NewContext(commlog, opts),
		uri:     uri,
		network: network,
		address: address,
	}

	// set TLS options
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
		comm.LogMsg(c.CommLog, "CONNECT_FAIL -- %v", err)
		return fmt.Errorf("%w, %v", comm.ErrConnection, err)
	}
	// set tls config for connection
	if c.tlsConfig != nil {
		conn = tls.Client(conn, c.tlsConfig)
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

	// determine read buffer size and polling timeout
	nRead := c.PollChunkSize
	if c.PollMaxSize > 0 {
		nRead = c.PollMaxSize
	}

	// set read polling duration and deadline
	var tPolling time.Duration
	var tDeadline time.Time

	// no polling for packet session-less connections
	if _, ok := c.netConn.(net.PacketConn); ok {
		if timeout > 0 {
			tPolling = time.Duration(timeout * float64(time.Second))
			tDeadline = time.Now().Add(tPolling)
		}
	} else {
		if c.PollTimeout > 0 {
			tPolling = time.Duration(c.PollTimeout * float64(time.Second))
		} else {
			tPolling = time.Duration(comm.POLL_TIMEOUT * float64(time.Second))
		}
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
	// Context containing common attributes such as logging and events.
	*comm.Context

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

	// TlsConfig defines the TLS attributes for TCP connections.
	tlsConfig *tls.Config
}

// NewListener creates a new network Listener.
func NewListener(uri string, commlog *logging.Logger, opts dictx.Dict) (*Listener, error) {
	uri = strings.TrimSpace(uri)
	network, address, err := ParseUri(uri)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		Context: comm.NewContext(commlog, opts),
		uri:     uri,
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
	if v := dictx.GetFloat(l.Options, "keepalive_interval", -1); v >= 0 {
		lConfig.KeepAlive = time.Duration(v * float64(time.Second))
	}

	// net listener
	netListener, err := lConfig.Listen(
		context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	// set connection limit
	if v := dictx.GetInt(l.Options, "connections_limit", 0); v > 0 {
		netListener = netutil.LimitListener(netListener, v)
	}
	// set tls config for listener
	if l.tlsConfig != nil {
		netListener = tls.NewListener(netListener, l.tlsConfig)
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
			uri := fmt.Sprintf("%s@%s", l.network, conn.RemoteAddr())
			c, err := NewConnection(uri, nil, l.Options)
			if err != nil {
				comm.LogMsg(l.CommLog, "CONN_ERROR -- %v", err)
				conn.Close()
				return
			}
			if l.CommLog != nil {
				c.CommLog = l.CommLog.SubLogger(uri)
			}
			c.netConn = conn
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
	var lConfig net.ListenConfig

	// packet connection
	packetConn, err := lConfig.ListenPacket(
		context.Background(), l.network, l.address)
	if err != nil {
		return err
	}
	l.netListener = packetConn

	comm.LogMsg(l.CommLog, "LISTENING -- %s", l.uri)

	c := &Connection{
		Context: comm.NewContext(l.CommLog, l.Options),
		uri:     l.uri,
		network: l.network,
		address: l.address,
		netConn: packetConn,
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
