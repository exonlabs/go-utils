package xcomm

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DEFAULT_CONNECTTIMEOUT    = float64(10)
	DEFAULT_KEEPALIVEINTERVAL = float64(30)
	DEFAULT_LISTENERPOOL      = int(1)
)

// Network Connection URI
//
// format:  <network>@<address>
//   <network>  {tcp|tcp4|tcp6|udp|udp4|udp6|ip|ip4|ip6|unix|unixgram|unixpacket}
//              For IP networks, the network must be "ip", "ip4" or "ip6"
//              followed by a colon and a literal protocol number or a
//              protocol name. ex: "ip4:1", "ip6:ipv6-icmp"
//   <address>  For TCP and UDP networks, the address has the form "host:port".
//              For Unix networks, the address must be a file system path.
//              For IP networks, the address has the form "host".
//
// -- see net.dial for full details.
//    (referance:  https://pkg.go.dev/net#Dial)
//
// example:
// server
//   - tcp@0.0.0.0:1234
//   - tcp6@[::1]:1234
//   - udp@0.0.0.0:http
//   - unix@/path/to/socket/file
// client
//   - tcp@1.2.3.4:1234
//   - tcp4@1.2.3.4:1234
//   - tcp6@[2001:db8::1]:http
//   - udp@1.2.3.4:1234
//   - unix@/path/to/socket/file

// parse and validate uri
func parse_net_uri(uri string) (string, string, error) {
	p := strings.SplitN(uri, "@", 2)
	if len(p) < 2 {
		return "", "", ErrInvalidUri
	}
	p[0] = strings.ToLower(p[0])

	switch p[0] {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		v := strings.Split(p[1], ":")
		if len(v) < 2 {
			return "", "", ErrInvalidUri
		}
	case "unix", "unixgram", "unixpacket":
		if p[1][len(p[1])-1] == filepath.Separator || !filepath.IsAbs(p[1]) {
			return "", "", ErrInvalidUri
		}
	}

	return p[0], p[1], nil
}

// //////////////////////////////////////////////////

// Network Connection
type NetConnection struct {
	*BaseConnection

	network string
	address string

	// low level network connection handler
	sock net.Conn

	// parent server handler
	parent *NetListener

	// connect timeout in sec
	ConnectTimeout float64
	// keepalive interval in sec, use 0 to disable
	KeepAliveInterval float64
}

// NewNetConnection creates a new SockClient instance
func NewNetConnection(
	uri string, log *Logger, opts Options) (*NetConnection, error) {
	var err error
	nc := &NetConnection{
		BaseConnection:    new_base_connection(uri, log, opts),
		ConnectTimeout:    DEFAULT_CONNECTTIMEOUT,
		KeepAliveInterval: DEFAULT_KEEPALIVEINTERVAL,
	}
	nc.network, nc.address, err = parse_net_uri(uri)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		nc.ConnectTimeout = opts.GetFloat64(
			"connect_timeout", DEFAULT_CONNECTTIMEOUT)
		nc.KeepAliveInterval = opts.GetFloat64(
			"keepalive_interval", DEFAULT_KEEPALIVEINTERVAL)
	}
	return nc, nil
}

func (nc *NetConnection) Parent() Listener {
	return nc.parent
}

func (nc *NetConnection) NetHandler() net.Conn {
	return nc.sock
}

func (nc *NetConnection) IsOpened() bool {
	return !(nc.evtKill.IsSet() || nc.sock == nil)
}

// Opens the socket connection
func (nc *NetConnection) Open() error {
	nc.Close()

	nc.evtBreak.Clear()
	nc.evtKill.Clear()

	nc.log("OPEN -- %s", nc.uri)

	var err error
	var ctx context.Context

	ctx, nc.ctxCancel = context.WithCancel(context.Background())
	d := net.Dialer{
		Timeout: time.Duration(nc.ConnectTimeout * 1000000000)}
	nc.sock, err = d.DialContext(ctx, nc.network, nc.address)
	nc.ctxCancel = nil
	if ctx.Err() != nil {
		return ErrBreak
	} else if err != nil {
		return fmt.Errorf("%w, %s", ErrOpen, err.Error())
	}

	if tcpConn, ok := nc.sock.(*net.TCPConn); ok &&
		nc.KeepAliveInterval > 0 && runtime.GOOS != "windows" {
		if err := tcpConn.SetKeepAlive(true); err != nil {
			return fmt.Errorf("%w, %s", ErrOpen, err.Error())
		}
		if err := tcpConn.SetKeepAlivePeriod(
			time.Duration(nc.KeepAliveInterval * 1000000000)); err != nil {
			return fmt.Errorf("%w, %s", ErrOpen, err.Error())
		}
		nc.sock = tcpConn
	}

	return nil
}

// Closes the socket connection
func (nc *NetConnection) Close() {
	nc.evtKill.Set()
	if nc.sock != nil {
		nc.sock.Close()
		nc.log("CLOSE -- %s", nc.uri)
	}
	nc.sock = nil
}

// cancel blocking operations
func (nc *NetConnection) Cancel() {
	nc.evtBreak.Set()
	if nc.ctxCancel != nil {
		nc.ctxCancel()
		nc.ctxCancel = nil
	}
}

// Sends data over the socket connection
func (nc *NetConnection) Send(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%w, empty data", ErrError)
	}
	if !nc.IsOpened() {
		return ErrNotOpend
	}
	nc.txLog(data)
	_, err := nc.sock.Write(data)
	if err != nil {
		if errIsClosed(err) {
			nc.Close()
			return ErrClosed
		}
		return fmt.Errorf("%w, %s", ErrWrite, err.Error())
	}
	return nil
}

// Recv data from the socket connection
func (nc *NetConnection) Recv() ([]byte, error) {
	if !nc.IsOpened() {
		return nil, ErrNotOpend
	}

	data := []byte(nil)
	for {
		b := make([]byte, nc.PollChunkSize)

		nc.sock.SetReadDeadline(
			time.Now().Add(time.Duration(nc.PollInterval * 1000000000)))

		n, err := nc.sock.Read(b)
		if err != nil {
			if errIsClosed(err) {
				nc.rxLog(data)
				nc.Close()
				return nil, ErrClosed
			}
			if err, ok := err.(net.Error); ok && err.Timeout() {
				break
			}
			return nil, fmt.Errorf("%w, %s", ErrRead, err.Error())
		}
		if n > 0 {
			data = append(data, b[:n]...)
		} else {
			break
		}
		if data != nil && n < nc.PollChunkSize {
			break
		}

		if nc.PollMaxSize > 0 && len(data) > nc.PollMaxSize {
			break
		}

		if nc.evtKill.IsSet() {
			nc.rxLog(data)
			return nil, ErrClosed
		}
		if nc.evtBreak.IsSet() {
			nc.rxLog(data)
			return nil, ErrBreak
		}
	}
	nc.rxLog(data)
	return data, nil
}

// Receives data with a specified timeout from the socket connection
func (nc *NetConnection) RecvWait(timeout float64) ([]byte, error) {
	nc.evtBreak.Clear()
	tbreak := float64(time.Now().Unix()) + timeout
	for {
		data, err := nc.Recv()
		if err != nil {
			return nil, err
		} else if len(data) > 0 {
			return data, nil
		}
		if nc.evtKill.IsSet() {
			return nil, ErrClosed
		}
		if nc.evtBreak.IsSet() {
			return nil, ErrBreak
		}
		if timeout > 0 && float64(time.Now().Unix()) >= tbreak {
			return nil, ErrTimeout
		}
	}
}

// //////////////////////////////////////////////////

// Network Listener
type NetListener struct {
	*BaseConnection

	network string
	address string

	// low level network listener handler
	sock net.Listener

	// callback connection handler function
	connHandler func(Connection)

	// max number for connected clients
	ListenerPool int
}

func NewNetListener(
	uri string, log *Logger, opts Options) (*NetListener, error) {
	var err error
	nl := &NetListener{
		BaseConnection: new_base_connection(uri, log, opts),
		ListenerPool:   DEFAULT_LISTENERPOOL,
	}
	nl.network, nl.address, err = parse_net_uri(uri)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		nl.ListenerPool = opts.GetInt("listener_pool", DEFAULT_LISTENERPOOL)
	}
	return nl, nil
}

func (nc *NetListener) NetHandler() net.Listener {
	return nc.sock
}

func (nl *NetListener) SetConnHandler(f func(Connection)) {
	nl.connHandler = f
}

func (nl *NetListener) IsActive() bool {
	return !(nl.evtKill.IsSet() || nl.sock == nil)
}

func (nl *NetListener) Start() error {
	if nl.connHandler == nil {
		return fmt.Errorf("%w, connection handler not set", ErrOpen)
	}

	nl.Stop()

	nl.evtBreak.Clear()
	nl.evtKill.Clear()

	nl.log("LISTEN -- %s", nl.uri)

	var err error
	nl.sock, err = net.Listen(nl.network, nl.address)
	if err != nil {
		return fmt.Errorf("%w, %s", ErrOpen, err.Error())
	}

	nl.run()
	return nil
}

func (nl *NetListener) run() {
	for nl.IsActive() {
		nc_sock, err := nl.sock.Accept()
		if err != nil {
			nl.log("TRACE -- %s", err.Error())
			continue
		}

		nc_uri := fmt.Sprintf("%s@%s", nl.network, nc_sock.RemoteAddr())
		nc, err := NewNetConnection(nc_uri, nl.Log, nil)
		nc.sock = nc_sock
		nc.parent = nl
		nc.uriLogging = true
		if err != nil {
			nl.log("TRACE -- %s", err.Error())
			continue
		}
		nl.log("NEWCONN -- %s", nc_uri)
		nl.connHandler(nc)
	}
}

func (nl *NetListener) Stop() {
	nl.evtKill.Set()
	if nl.sock != nil {
		nl.sock.Close()
		nl.log("CLOSE -- %s", nl.uri)
	}
	nl.sock = nil
}
