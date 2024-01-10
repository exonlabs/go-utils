package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xconn/utils"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

// func newPeerHandler(server *server.SockServer) {
// 	fmt.Println("new peer:", server.String())
// 	server.Send([]byte(fmt.Sprintf("HELLO: %s\n")))
// }

// func endPeerHandler(server *server.SockServer) {
// 	fmt.Println("end peer:", server.String())
// }

func HandleConnection2(peer xconn.ClientConn) {

	defer peer.Close()

	for {

		data, err := peer.RecvWait(0)
		if data == nil {
			break
		}
		if err != nil {
			fmt.Println("error receiving data: ", err)
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
func HandleConnection3(peer xconn.ClientConn) {
	go HandleConnection2(peer)

	// go func() {
	// 	fmt.Println("new conn")
	// 	for i := 0; i < 10; i++ {
	// 		fmt.Println("loop conn -----------------")
	// 		peer.Send([]byte("SHehab welcomes you"))
	// 		time.Sleep(time.Second)
	// 	}
	// 	fmt.Println("end conn")
	// }()
}

func main() {
	var debug int
	flag.IntVar(&debug, "x", 0, "set debug modes:\n"+
		"-x      debug ON\n"+
		"-xx     debug ON + comm logs\n")

	var uri string
	flag.StringVar(&uri, "uri", "", "connection uri\n"+
		"serial:  SER:/dev/ttyUSB0:115200:8N1\n"+
		"tcp:     TCP:0.0.0.0:1234\n")
	flag.Parse()

	Logger := xlog.GetLogger()

	fmt.Println("***** Starting Connection Server *****")

	srv, err := utils.NewServerConn("TCP:0.0.0.0:8080", HandleConnection3, Logger)
	println(err)
	// if srv.Type == "tcp" || srv.Type == "udp" {
	// srv.SetPeerConnectHandler(newPeerHandler)
	// srv.SetPeerDisconnectHandler(endPeerHandler)
	// }
	err = srv.Open()
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		Logger.Panic("Server start error: %v", err)
	}

	// Allow time for the server to run
	time.Sleep(time.Second * 10)

	if err := srv.Close(); err != nil {
		Logger.Error("Error stopping server: %v", err)
	}
}
