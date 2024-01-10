package socket

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

type SockClient struct {
	*BaseSocket
	sock net.Conn
}

// NewSockClient creates a new SockClient instance
func NewSockClient(ip string, port int, udp bool, log *xlog.Logger) *SockClient {
	return &SockClient{
		BaseSocket: NewBaseSocket(ip, port, udp, log),
	}
}

// String returns a string representation of SockClient
func (conn *SockClient) String() string {
	return fmt.Sprintf("<%v: %v %v>",
		conn.Type(), conn.Ip, conn.Port)
}

//	func (conn *SockClient) Info() string {
//		return fmt.Sprintf("<%v: %v %v>",
//			conn.Type(), conn.Ip, conn.Port)
//	}
func (conn *SockClient) IsOpened() bool {
	// TODO check connection is really UP
	if conn.sock != nil {
		return true
	}
	return false
}

// Opens the socket connection
func (conn *SockClient) Open() error {
	addr := fmt.Sprintf("%v:%v", conn.Ip, conn.Port)
	// close existing sock if peer changed
	if conn.sock != nil {
		if conn.sock.RemoteAddr().String() != addr {
			conn.sock.Close()
		}
	}

	var err error
	conn.sock, err = net.DialTimeout(strings.ToLower(conn.Type()),
		addr, time.Millisecond*time.Duration(conn.ConnectTimeout*1000))
	if err != nil {
		return err
	}

	switch strings.ToUpper(conn.Type()) {
	case "UDP":
	case "TCP":
		tcpConn := conn.sock.(*net.TCPConn)
		if err := tcpConn.SetKeepAlive(conn.KeepAlive); err != nil {
			return err
		}

		if conn.KeepAlive {
			if err := tcpConn.SetKeepAlivePeriod(
				time.Second * time.Duration(conn.KeepAliveInterval)); err != nil {
				return err
			}
		}

		conn.sock = tcpConn
	default:
		return errors.New("INVALID_SOCK_TYPE")
	}

	if conn.Log != nil {
		conn.Log.Info("OPEN -- %v", conn)
	}

	return nil
}

// Closes the socket connection
func (conn *SockClient) Close() error {
	if conn.sock != nil {
		if err := conn.sock.Close(); err != nil {
			return err
		}

		conn.EvtKill.Set()

		conn.sock = nil

		if conn.Log != nil {
			conn.Log.Info("CLOSE -- %v", conn)
		}
	}

	return nil
}

// Sends data over the socket connection
func (conn *SockClient) Send(data []byte) error {
	if conn.sock == nil {
		return xconn.ErrNotOpend
	}

	if len(data) == 0 {
		return errors.New("EMPTY_DATA")
	}

	conn.TxLog(data)

	_, err := conn.sock.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Recv data from the socket connection
func (conn *SockClient) Recv() ([]byte, error) {
	if conn.sock == nil {
		return nil, xconn.ErrNotOpend
	}

	var res []byte
	var errStr string
	chunkRead := 1024

	for {
		buff := make([]byte, chunkRead)

		// Similar to select in pyUSI
		conn.sock.SetReadDeadline(
			time.Now().Add(time.Millisecond * time.Duration(0.1*1000)))
		n, err := conn.sock.Read(buff)

		if err != nil {
			if err, ok := err.(net.Error); ok &&
				err.Timeout() {
				break
			}
			errStr = err.Error()
			break
		}

		if n > 0 {
			res = append(res, buff[:n]...)
		}

		if conn.EvtBreak.IsSet() {
			errStr = "BREAK_REQUEST"
			break
		}

		if conn.PollMaxSize != 0 && len(res) > conn.PollMaxSize {
			errStr = "MAX_DATA_LIMIT - reached max receive limit"
			break
		}

		if n > chunkRead {
			break
		}
	}

	if len(errStr) > 0 {
		// if errStr != io.EOF.Error() {
		return nil, errors.New(errStr)
		// }
	}

	if len(res) > 0 {
		conn.RxLog(res)
	}

	return res, nil
}

// Receives data with a specified timeout from the socket connection
func (conn *SockClient) RecvWait(timeout float64) ([]byte, error) {
	conn.EvtBreak.Clear()
	tlimit := time.Now().Add(time.Millisecond * time.Duration(timeout*1000))

	for {
		data, err := conn.Recv()
		if err != nil {
			return nil, err
		}

		if len(data) > 0 {
			return data, nil
		}

		if conn.EvtKill.IsSet() {
			return nil, xconn.ErrClosed
		}

		if timeout > 0 {
			if time.Now().After(tlimit) {
				return nil, xconn.ErrTimeout
			}
		}
	}
}
