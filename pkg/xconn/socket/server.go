package socket

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

type SockServer struct {
	*BaseSocket
	sock             net.Listener
	HandleConnection func(xconn.ClientConn)
}

func NewSockServer(ip string, port int, udp bool, log *xlog.Logger) *SockServer {
	return &SockServer{
		BaseSocket: NewBaseSocket(ip, port, udp, log),
	}

}

func (conn *SockServer) String() string {
	return fmt.Sprint(conn.sock.Addr())
}

func (conn *SockServer) IsOpened() bool {
	// chcek connection is realy UP
	if conn.sock != nil {
		return true
	}
	return false

}

func (conn *SockServer) Open() error {
	conn.Close()
	addr := fmt.Sprintf("%v:%v", conn.Ip, conn.Port)

	var err error
	conn.sock, err = net.Listen(strings.ToLower(conn.Type()), addr)
	if err != nil {
		return err
	}
	conn.Log.Info("Server listening on %s:%d", conn.Ip, conn.Port)
	conn.OpenPeer()
	return err
}

func (conn *SockServer) OpenPeer() error {

	for {
		logger := xlog.GetLogger()
		peer, err := conn.sock.Accept()
		if err != nil {
			fmt.Println("Error accepting connection ")
			return err
		}
		address := strings.Split(peer.RemoteAddr().String(), ":")
		ip := address[0]
		port, err := strconv.Atoi(address[1])
		PeerSock := NewSockClient(ip, port, conn.Type() == "UDP", logger)
		PeerSock.sock = peer
		if err != nil {
			return err
		}

		fmt.Println("Accepted new connection from", PeerSock.sock.RemoteAddr())
		conn.HandleConnection(PeerSock)

	}
}

func (conn *SockServer) Close() error {
	if conn.sock != nil {
		return conn.sock.Close()
	}
	return nil
}
