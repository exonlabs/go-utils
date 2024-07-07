package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/xcomm"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

func HandleConnection(conn xcomm.Connection) {
	if conn.Type() != "serial" {
		defer fmt.Println("End peer connection", conn)
		fmt.Println("New peer connection", conn)
	}

	if err := conn.Send([]byte("HELLO\n")); err != nil {
		fmt.Println(err.Error())
		return
	}

	for conn.IsOpened() {
		data, err := conn.RecvWait(0)
		if err != nil {
			if conn.IsOpened() && err != xcomm.ErrClosed {
				fmt.Println("error receiving:", err)
				continue
			} else {
				break
			}
		}

		msg := strings.TrimSpace(string(data))
		fmt.Println("received:", msg)

		switch msg {
		case "STOP_PEER":
			conn.Send([]byte("peer stopped by server"))
			conn.Close()
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

func main() {
	uri := flag.String(
		"uri", "", "connection uri\n"+
			"serial:  serial@/dev/ttyUSB0:115200:8N1\n"+
			"tcp:     tcp@0.0.0.0:1234\n")
	multi := flag.Bool(
		"multi", false, "allow multiple sessions for tcp connections")
	flag.Parse()

	fmt.Println("***** Starting Connection Server *****")

	logger := xlog.NewStdoutLogger("main")

	conn_opts := xcomm.Options{"connections_limit": 1}
	if *multi {
		conn_opts.Set("connections_limit", 5)
	}
	srv, err := xcomm.NewListener(*uri, logger, conn_opts)
	if err != nil {
		panic(err)
	}
	srv.SetConnHandler(HandleConnection)

	// register callback for close signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		s := <-sigCh
		fmt.Printf("\nreceived signal: %s\n", s)
		srv.Stop()
	}()

	if err := srv.Start(); err != nil {
		panic(err)
	}
	fmt.Println("** exit **")
}
