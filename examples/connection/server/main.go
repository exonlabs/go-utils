package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/exonlabs/go-utils/pkg/xcomm"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

func HandleConnection(conn xcomm.Connection) {
	defer func() {
		fmt.Println("End peer connection", conn)
		conn.Close()
	}()

	t := conn.Type()
	if t != "serial" {
		fmt.Println("New peer connection", conn)
		if err := conn.Send([]byte("HELLO\n")); err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	for {
		data, err := conn.RecvWait(0)
		if err != nil {
			if !conn.IsOpened() {
				return
			}
			fmt.Println("error receiving:", err)
			return
		}

		msg := strings.TrimSpace(string(data))
		fmt.Println("received:", msg)

		switch msg {
		case "STOP_PEER":
			conn.Send([]byte("peer stopped by server"))
			return
		case "STOP_SERVER":
			conn.Send([]byte("server stopped"))
			conn.Parent().Stop()
			return
		default:
			conn.Send([]byte("echo: " + msg + "\n"))
		}
		fmt.Println("--------------------------------")
	}
}

func HandleConnectionMulti(conn xcomm.Connection) {
	go HandleConnection(conn)
}

func main() {
	uri := flag.String("uri", "", "connection uri\n"+
		"serial:  serial@/dev/ttyUSB0:115200:8N1\n"+
		"tcp:     tcp@0.0.0.0:1234\n")
	multi := flag.Bool("multi", false, "handle multiple connections")
	flag.Parse()

	fmt.Println("***** Starting Connection Server *****")

	logger := xlog.NewStdoutLogger("main")

	srv, err := xcomm.NewListener(*uri, logger, nil)
	if err != nil {
		panic(err)
	}
	if *multi {
		srv.SetConnHandler(HandleConnectionMulti)
	} else {
		srv.SetConnHandler(HandleConnection)
	}
	if err := srv.Start(); err != nil {
		panic(err)
	}
}
