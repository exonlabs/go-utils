package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/xcomm"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

var exit_event = xevent.NewEvent()

func run(cli xcomm.Connection) {
	if err := cli.Open(); err != nil {
		if errors.Is(err, xcomm.ErrBreak) {
			return
		}
		panic(err)
	}
	defer cli.Close()

	if data, err := cli.RecvWait(5); err == nil {
		fmt.Printf("received: %d bytes  %s", len(data), string(data))
		fmt.Println("--------------------------------")
		exit_event.Wait(1)
	}

	num_of_msgs := 5
	for i := 1; i <= num_of_msgs; i++ {
		if exit_event.IsSet() || !cli.IsOpened() {
			break
		}
		msg := []byte(fmt.Sprintf("msg: %d\n", i))
		fmt.Printf("sending: %d bytes  %s", len(msg), string(msg))
		err := cli.Send(msg)
		if err == nil {
			data, err := cli.RecvWait(5)
			if err == nil {
				fmt.Printf("received: %d bytes  %s", len(data), string(data))
				fmt.Println("--------------------------------")
			}
		}
		if err != nil {
			if err == xcomm.ErrClosed {
				break
			} else {
				fmt.Println(err)
			}
		}
		if i < num_of_msgs {
			exit_event.Wait(1)
		}
	}

	fmt.Println("end connection")
}

func main() {
	uri := flag.String(
		"uri", "", "connection uri\n"+
			"serial:  serial@/dev/ttyUSB0:115200:8N1\n"+
			"tcp:     tcp@127.0.0.1:1234\n")
	flag.Parse()

	fmt.Println("***** Starting Connection Client *****")

	logger := xlog.NewStdoutLogger("main")

	cli, err := xcomm.NewConnection(*uri, logger, nil)
	if err != nil {
		panic(err)
	}

	// register callback for close signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		s := <-sigCh
		fmt.Printf("\nreceived signal: %s\n", s)
		cli.Cancel()
		cli.Close()
		exit_event.Set()
	}()

	run(cli)
	fmt.Println("** exit **")
}
