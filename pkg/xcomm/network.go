package xcomm

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/netutil"
)

const (
	NET_CONNECT_TIMEOUT    = float64(10)
	NET_KEEPALIVE_INTERVAL = float64(30)
	NET_CONNECTIONS_LIMIT  = int(0)
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
		return "", "", ErrUri
	}
	p[0] = strings.ToLower(p[0])

	switch p[0] {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		v := strings.Split(p[1], ":")
		if len(v) < 2 {
			return "", "", ErrUri
		}
	case "unix", "unixgram", "unixpacket":
		if p[1][len(p[1])-1] == filepath.Separator || !filepath.IsAbs(p[1]) {
			return "", "", ErrUri
		}
	}

	return p[0], p[1], nil
}

// //////////////////////////////////////////////////

// Network Connection
type NetConnection struct {
	*baseConnection

	// connection attrs
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
		baseConnection:    new_base_connection(uri, log, opts),
		ConnectTimeout:    NET_CONNECT_TIMEOUT,
		KeepAliveInterval: NET_KEEPALIVE_INTERVAL,
	}
	nc.network, nc.address, err = parse_net_uri(uri)
	if err != nil {
		return nil, fmt.Errorf("%w%s", ErrError, err)
	}
	if opts != nil {
		if v := opts.GetFloat64("connect_timeout", 0); v >= 0 {
			nc.ConnectTimeout = v
		}
		if v := opts.GetFloat64("keepalive_interval", 0); v >= 0 {
			nc.KeepAliveInterval = v
		}
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
	if !nc.op_mux.TryLock() {
		return nil
	}
	defer nc.op_mux.Unlock()

	nc.evtBreak.Clear()
	nc.evtKill.Clear()

	if nc.uriLogging {
		nc.comm_log("CONNECT")
	} else {
		nc.comm_log("OPEN -- %s", nc.uri)
	}

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
		return fmt.Errorf("%w, %s", ErrConnection, err)
	}

	if tcpConn, ok := nc.sock.(*net.TCPConn); ok &&
		nc.KeepAliveInterval > 0 && runtime.GOOS != "windows" {
		if err := tcpConn.SetKeepAlive(true); err != nil {
			return fmt.Errorf("%w, %s", ErrConnection, err)
		}
		if err := tcpConn.SetKeepAlivePeriod(
			time.Duration(nc.KeepAliveInterval * 1000000000)); err != nil {
			return fmt.Errorf("%w, %s", ErrConnection, err)
		}
		nc.sock = tcpConn
	}

	return nil
}

// Closes the socket connection
func (nc *NetConnection) Close() {
	nc.op_mux.Lock()
	defer nc.op_mux.Unlock()

	nc.evtKill.Set()
	if nc.sock != nil {
		nc.sock.Close()
		nc.sock = nil
		if nc.uriLogging {
			nc.comm_log("DISCONNECT")
		} else {
			nc.comm_log("CLOSE -- %s", nc.uri)
		}
	}
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
		return ErrClosed
	}
	nc.tx_Log(data)
	_, err := nc.sock.Write(data)
	if err != nil {
		if errIsClosed(err) {
			nc.comm_log("SOCK_CLOSED - %s", err.Error())
			nc.Close()
			return ErrClosed
		}
		return fmt.Errorf("%w, %s", ErrWrite, err)
	}
	return nil
}

// Recv data from the socket connection
func (nc *NetConnection) Recv() ([]byte, error) {
	if !nc.IsOpened() {
		return nil, ErrClosed
	}

	data := []byte(nil)
	td := time.Duration(nc.PollInterval * 1000000000)
	for {
		b := make([]byte, nc.PollChunkSize)
		if td > 0 {
			nc.sock.SetReadDeadline(time.Now().Add(td))
		}
		n, err := nc.sock.Read(b)
		if err != nil {
			if errIsClosed(err) {
				nc.rx_Log(data)
				nc.comm_log("SOCK_CLOSED - %s", err.Error())
				nc.Close()
				return nil, ErrClosed
			}
			if err, ok := err.(net.Error); ok && err.Timeout() {
				break
			}
			return nil, fmt.Errorf("%w, %s", ErrRead, err)
		}
		if n > 0 {
			data = append(data, b[:n]...)
			if n < nc.PollChunkSize {
				break
			}
		} else {
			break
		}

		if nc.PollMaxSize > 0 && len(data) > nc.PollMaxSize {
			break
		}

		if nc.evtKill.IsSet() || (nc.parent != nil && !nc.parent.IsActive()) {
			nc.rx_Log(data)
			return nil, ErrClosed
		}
		if nc.evtBreak.IsSet() {
			nc.rx_Log(data)
			return nil, ErrBreak
		}
	}
	nc.rx_Log(data)
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
		if nc.evtKill.IsSet() || (nc.parent != nil && !nc.parent.IsActive()) {
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
	*baseConnection

	// connection attrs
	network string
	address string

	// low level network listener handler
	sock net.Listener
	// operation sync wait group
	op_wg sync.WaitGroup
	// callback connection handler function
	connHandler func(Connection)

	// limit simultaneous connections, 0 for unlimited
	connLimit int
}

func NewNetListener(
	uri string, log *Logger, opts Options) (*NetListener, error) {
	var err error
	nl := &NetListener{
		baseConnection: new_base_connection(uri, log, opts),
		connLimit:      NET_CONNECTIONS_LIMIT,
	}
	nl.network, nl.address, err = parse_net_uri(uri)
	if err != nil {
		return nil, err
	}
	if opts != nil {
		nl.connLimit = opts.GetInt("connections_limit", NET_CONNECTIONS_LIMIT)
	}
	return nl, nil
}

func (nl *NetListener) NetHandler() net.Listener {
	return nl.sock
}

func (nl *NetListener) SetConnHandler(h func(Connection)) {
	nl.connHandler = h
}

func (nl *NetListener) IsActive() bool {
	return !(nl.evtKill.IsSet() || nl.sock == nil)
}

func (nl *NetListener) Start() error {
	if nl.connHandler == nil {
		return fmt.Errorf("%wconnection handler not set", ErrError)
	}

	nl.evtBreak.Clear()
	nl.evtKill.Clear()

	nl.comm_log("LISTEN -- %s", nl.uri)

	var err error
	nl.sock, err = net.Listen(nl.network, nl.address)
	if err != nil {
		return fmt.Errorf("%w%s", ErrError, err)
	}
	// set connections limit
	if nl.connLimit > 0 {
		nl.sock = netutil.LimitListener(nl.sock, nl.connLimit)
	}

	// run forever
	for nl.IsActive() {
		nc_sock, err := nl.sock.Accept()
		if err != nil {
			if nl.IsActive() {
				nl.comm_log("CONN_ERROR -- %s", err.Error())
				continue
			} else {
				break
			}
		}

		nc_uri := fmt.Sprintf("%s@%s", nl.network, nc_sock.RemoteAddr())
		nc, err := NewNetConnection(nc_uri, nl.commLogger, nil)
		if err != nil {
			nl.comm_log("CONN_ERROR -- %s", err.Error())
			continue
		}
		nc.sock = nc_sock
		nc.parent = nl
		nc.uriLogging = true
		nc.comm_log("CONNECT")

		nl.op_wg.Add(1)
		go func() {
			defer nl.op_wg.Done()
			defer nc.Close()
			nl.connHandler(nc)
		}()
	}

	nl.op_wg.Wait()
	nl.comm_log("CLOSED -- %s", nl.uri)
	return nil
}

func (nl *NetListener) Stop() {
	nl.op_mux.Lock()
	defer nl.op_mux.Unlock()

	nl.evtKill.Set()
	if nl.sock != nil {
		nl.sock.Close()
		nl.sock = nil
	}
}
