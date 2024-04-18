package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xcomm"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

func run(cli xcomm.Connection) {
	err := cli.Open()
	if err != nil {
		if errors.Is(err, xcomm.ErrBreak) {
			return
		}
		panic(err)
	}
	defer cli.Close()

	t := cli.Type()
	if t != "serial" {
		if data, err := cli.RecvWait(5); err == nil {
			fmt.Printf("received: %d  %s", len(data), string(data))
			fmt.Println("--------------------------------")
			cli.Sleep(1)
		}
	}

	num_of_msgs := 5
	for i := 1; i <= num_of_msgs; i++ {
		msg := []byte(fmt.Sprintf("msg: %d\n", i))
		fmt.Printf("sending: %d  %s", len(msg), string(msg))
		err = cli.Send(msg)
		if err == nil {
			data, err := cli.RecvWait(5)
			if err == nil {
				fmt.Printf("received: %d  %s", len(data), string(data))
				fmt.Println("--------------------------------")
			}
		}
		if err != nil {
			fmt.Println(err)
		}
		if !cli.IsOpened() {
			fmt.Println("connection closed")
			return
		}

		cli.Sleep(1)
	}
}

func main() {
	uri := flag.String("uri", "", "connection uri\n"+
		"serial:  serial@/dev/ttyUSB0:115200:8N1\n"+
		"tcp:     tcp@127.0.0.1:1234\n")
	flag.Parse()

	fmt.Println("***** Starting Connection Client *****")

	logger := xlog.NewStdoutLogger("main")

	cli, err := xcomm.NewConnection(*uri, nil, logger)
	if err != nil {
		panic(err)
	}

	run(cli)
}
