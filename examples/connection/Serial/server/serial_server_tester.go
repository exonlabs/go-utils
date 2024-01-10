package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xconn/utils"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

// HandleConnection2 handles the communication with the connected client
func HandleConnection2(peer xconn.ClientConn) {
	defer peer.Close()

	for {
		data, err := peer.RecvWait(0)
		if data == nil {
			break
		}
		if err != nil {
			fmt.Println("error receiving data:", err)
			break
		}
		fmt.Println("received:", string(data), len(data))
		if strings.Contains(string(data), "STOP_PEER") {
			peer.Send([]byte("peer stopped by server"))
			return
		} else if strings.Contains(string(data), "STOP_SERVER") {
			peer.Send([]byte("server stopped"))
			return
		} else {
			peer.Send([]byte("echo: " + string(data)))
		}
		fmt.Println("--------------------------------")
		time.Sleep(time.Second)
	}
}

// HandleConnection3 sets up a goroutine to handle the connection
func HandleConnection3(peer xconn.ClientConn) {
	go HandleConnection2(peer)
}

func main() {
	var debug int
	flag.IntVar(&debug, "x", 0, "set debug modes:\n"+
		"-x      debug ON\n"+
		"-xx     debug ON + comm logs\n")

	var uri string
	flag.StringVar(&uri, "uri", "/dev/ttyUSB0", "serial port path, e.g., /dev/ttyUSB0")
	flag.Parse()

	Logger := xlog.GetLogger()

	fmt.Println("***** Starting Serial Server *****")

	// Use the provided URI from the command-line arguments
	srv, _ := utils.NewServerConn("SER:"+uri+":115200:8N1", HandleConnection3, Logger)

	err := srv.Open()
	if err != nil {
		Logger.Panic("Serial server start error: %v", err)
	}

	// Allow time for the server to run
	time.Sleep(time.Second * 1000)

	if err := srv.Close(); err != nil {
		Logger.Error("Error stopping serial server: %v", err)
	}
}
