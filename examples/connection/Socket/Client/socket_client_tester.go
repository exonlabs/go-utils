package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/exonlabs/go-utils/pkg/xconn"
	"github.com/exonlabs/go-utils/pkg/xconn/utils"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

func run(cli xconn.ClientConn) {
	num_of_msgs := 10
	err := cli.Open()
	if err != nil {
		log.Println(err)
		return
	}
	// data, err := cli.RecvWait(1)
	// if err == nil {
	// 	fmt.Printf("received: %s, %d\n", string(data), len(data))
	// 	fmt.Println("--------------------------------")
	// }
	// }

	for i := 1; i <= num_of_msgs; i++ {
		msg := []byte(fmt.Sprintf("msg: %d\n", i))
		fmt.Printf("sending: %s, %d\n", string(msg), len(msg))
		cli.Send(msg)

		data, err := cli.RecvWait(5)
		if err == nil {
			fmt.Printf("received: %s, %d\n", string(data), len(data))
			fmt.Println("--------------------------------")
		} else {
			fmt.Println(err)
		}

		time.Sleep(time.Second / 2)
	}

	cli.Close()
}

func main() {
	var debugLevel int
	var uri string

	flag.IntVar(&debugLevel, "x", 0, "set debug modes:\n"+
		"-x      debug ON\n"+
		"-xx     debug ON + comm logs\n")
	flag.StringVar(&uri, "uri", "", "connection uri\n"+
		"serial:  SER:/dev/ttyUSB0:115200:8N1\n"+
		"tcp:     TCP:127.0.0.1:1234\n")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if debugLevel > 0 {
		log.SetOutput(os.Stdout)
	}

	if debugLevel > 1 {
		commLog := log.New(os.Stdout, "", log.LstdFlags)
		log.SetOutput(os.Stdout)
		log.Printf("commLog: %v\n", commLog)
	}

	log.Println("***** Starting Connection Client *****")
	logger := xlog.NewLogger("Client")
	cli, err := utils.NewClientConn("TCP:127.0.0.1:8080", logger)
	if err != nil {
		panic("Failed To Create New Client")
	}
	run(cli)
}
