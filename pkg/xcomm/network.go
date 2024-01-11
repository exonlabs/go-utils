package xcomm

import (
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	defaultConnectTimeout    = float64(10)
	defaultKeepAliveInterval = float64(30)
	defaultListenerPool      = int(1)
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
func parseNetURI(uri string) (string, string, error) {
	p := strings.SplitN(uri, "@", 2)
	if len(p) < 2 {
		return "", "", ErrInvalidUri
	}

	switch p[0] {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		v := strings.Split(p[1], ":")
		if len(v) < 2 {
			return "", "", ErrInvalidUri
		}
	case "unix", "unixgram", "unixpacket":
		if p[1][len(p[1])-1] == filepath.Separator ||
			filepath.IsAbs(p[1]) != true {
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
func NewNetConnection(uri string, log *xlog.Logger) (*NetConnection, error) {
	nc := &NetConnection{
		BaseConnection:    NewBaseConnection(uri, log),
		ConnectTimeout:    defaultConnectTimeout,
		KeepAliveInterval: defaultKeepAliveInterval,
	}
	var err error
	nc.network, nc.address, err = parseNetURI(uri)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func (nc *NetConnection) Parent() Listener {
	return nc.parent
}

func (nc *NetConnection) IsOpened() bool {
	return bool(nc.sock != nil)
}

// Opens the socket connection
func (nc *NetConnection) Open() error {
	nc.Close()

	nc.evtBreak.Clear()
	nc.evtKill.Clear()

	nc.log("OPEN -- %s", nc.uri)

	var err error

	nc.sock, err = net.DialTimeout(
		nc.network, nc.address, time.Duration(nc.ConnectTimeout*1000000000))
	if err != nil {
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

// Sends data over the socket connection
func (nc *NetConnection) Send(data []byte) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("%w, empty data", ErrError)
	}
	if nc.sock == nil {
		return ErrNotOpend
	}

	nc.txLog(data)
	_, err := nc.sock.Write(data)
	if err != nil {
		if errIsClosed(err) {
			nc.Close()
		}
		return fmt.Errorf("%w, %s", ErrWrite, err.Error())
	}
	return nil
}

// Recv data from the socket connection
func (nc *NetConnection) Recv() ([]byte, error) {
	if nc.sock == nil {
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
		} else if data != nil && len(data) > 0 {
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

// cancel blocking operations
func (nc *NetConnection) Cancel() {
	nc.evtBreak.Set()
}

// //////////////////////////////////////////////////

// Network Listener
type NetListener struct {
	*BaseConnection
	network string
	address string
	sock    net.Listener

	// callback connection handler function
	connHandler func(Connection)

	// max number for connected clients
	ListenerPool int
}

func NewNetListener(uri string, log *xlog.Logger) (*NetListener, error) {
	nl := &NetListener{
		BaseConnection: NewBaseConnection(uri, log),
		ListenerPool:   defaultListenerPool,
	}
	var err error
	nl.network, nl.address, err = parseNetURI(uri)
	if err != nil {
		return nil, err
	}
	return nl, nil
}

func (nl *NetListener) IsActive() bool {
	return bool(nl.sock != nil)
}

func (nl *NetListener) Start() error {
	if nl.connHandler == nil {
		return fmt.Errorf("%w, invalid connection handler", ErrOpen)
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
	for !nl.evtKill.IsSet() {
		nc_sock, err := nl.sock.Accept()
		if err != nil {
			nl.log("TRACE -- %s", err.Error())
			continue
		}

		nc_uri := fmt.Sprintf("%s@%s", nl.network, nc_sock.RemoteAddr())
		nc, err := NewNetConnection(nc_uri, nl.Log)
		nc.sock = nc_sock
		nc.parent = nl
		nc.logUri = true
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

func (nl *NetListener) SetHandler(f func(Connection)) {
	nl.connHandler = f
}
