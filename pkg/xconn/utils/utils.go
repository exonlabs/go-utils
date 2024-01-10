package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xconn/serial"
	"github.com/exonlabs/go-utils/pkg/xconn/socket"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

func NewClientConn(uri string, log *xlog.Logger) (xconn.ClientConn, error) {
	p := strings.Split(uri, ":")
	if p[0] == "SER" && len(p) >= 3 {
		baudrate, _ := strconv.Atoi(p[2])
		client := serial.NewSerialClient(p[1], baudrate, p[3], log)
		client.Uri = uri
		return client, nil
	} else if (p[0] == "TCP" || p[0] == "UDP") && len(p) >= 3 {
		port, _ := strconv.Atoi(p[2])
		client := socket.NewSockClient(p[1], port, p[0] == "UDP", log)
		client.Uri = uri
		return client, nil
	} else {
		return nil, errors.New("INVALID_CONNECTION_URI")
	}
}

func NewServerConn(uri string, handler func(xconn.ClientConn), log *xlog.Logger) (xconn.ServerConn, error) {
	p := strings.Split(uri, ":")
	if p[0] == "SER" && len(p) >= 3 {
		baudrate, _ := strconv.Atoi(p[2])
		server := serial.NewSerialServer(p[1], baudrate, p[3], log)
		server.HandleConnection = handler
		return server, nil
	} else if (p[0] == "TCP" || p[0] == "UDP") && len(p) >= 3 {
		port, _ := strconv.Atoi(p[2])
		server := socket.NewSockServer(p[1], port, p[0] == "UDP", log)
		server.HandleConnection = handler
		return server, nil
	} else {
		return nil, nil
	}
}
