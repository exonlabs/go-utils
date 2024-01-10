package socket

import (
	"errors"
	"strconv"
	"strings"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

type BaseSocket struct {
	*xconn.BaseConnection
	Ip                string
	Port              int
	Udp               bool
	ConnectTimeout    float64
	ConnectionPool    int
	KeepAlive         bool
	KeepAliveIdleTime float64
	KeepAliveInterval float64
}

func NewBaseSocket(ip string, port int, udp bool, log *xlog.Logger) *BaseSocket {
	return &BaseSocket{

		BaseConnection: xconn.NewBaseConnection("", log),
		Ip:             ip,
		Port:           port,
		Udp:            udp,
	}
}

func (bs *BaseSocket) Type() string {
	if bs.Udp {
		return "UDP"
	} else {
		return "TCP"
	}
}

func (bs *BaseSocket) Init() error {
	p := strings.Split(bs.Uri, ":")
	if p[0] != "TCP" && p[0] != "UDP" || len(p) < 3 {
		return errors.New("INVALID_CONNECTION_URI")
	}

	bs.Ip = p[1]
	bs.Port, _ = strconv.Atoi(p[2])
	bs.Udp = strings.ToUpper(p[0]) == "UDP"

	return nil
}
