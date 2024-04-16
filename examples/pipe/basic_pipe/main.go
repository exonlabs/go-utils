package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/unix/xpipe"
)

var (
	tmp_path  = filepath.Join(os.TempDir(), "foobar")
	pipe_file = filepath.Join(tmp_path, "foo.pipe")
)

func sender_print(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
}

func receiver_print(msg string, args ...any) {
	fmt.Printf("                         "+msg+"\n", args...)
}

func CheckPeerPipe(path string, timeout float64) {
	sender_print("-- starting sender")
	p := xpipe.NewPipe(path)
	defer sender_print("-- stop sender")

	msg := "HELLO"
	if err := p.WriteWait([]byte(msg), timeout); err == nil {
		sender_print("-- CONNECTED")
		sender_print("-- sending >> " + msg)
	} else if errors.Is(err, xpipe.ErrTimeout) {
		sender_print("-- TIMEOUT: no peer connected")
	} else {
		sender_print("-- FAILED: %s", err.Error())
	}
}

func StartInputPipe(path string, timeout float64) {
	receiver_print("-- starting receiver")
	p := xpipe.NewPipe(path)
	defer receiver_print("-- stop receiver")

	b, err := p.ReadWait(timeout)
	if err != nil {
		receiver_print("-- %s", err.Error())
		return
	}
	receiver_print("-- received << %s", b)
}

func SendingMsgs(path string, timeout float64) {
	sender_print("-- starting sender")
	p := xpipe.NewPipe(path)
	defer sender_print("-- stop sender")

	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("MESSAGE_%d", i)
		sender_print("-- sending >> " + msg)
		if err := p.WriteWait([]byte(msg), timeout); err != nil {
			sender_print("-- FAILED: %s", err.Error())
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func StartCmdHandler(path string, timeout float64) {
	receiver_print("-- starting receiver")
	p := xpipe.NewPipe(path)
	defer receiver_print("-- stop receiver")

	var wg sync.WaitGroup
	evtClose := xevent.NewEvent()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for !evtClose.IsSet() {
			b, err := p.ReadWait(timeout)
			if err == nil {
				receiver_print("-- received << %s", b)
			} else if !errors.Is(err, xpipe.ErrBreak) {
				receiver_print("-- %s", err.Error())
			}
		}
	}()

	SendingMsgs(path, 2)
	evtClose.Set()
	p.Cancel()
	wg.Wait()
}

func main() {
	fmt.Printf("\n**** starting ****\n")

	fmt.Printf("\nUsing Pipe: %s\n", pipe_file)

	syscall.Umask(0)
	os.MkdirAll(tmp_path, 0o775)
	defer os.RemoveAll(tmp_path)

	if err := xpipe.CreatePipe(pipe_file, 0o666); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer xpipe.DeletePipe(pipe_file)

	// check without peer
	fmt.Printf("\n\n* checking with no peer:\n")
	CheckPeerPipe(pipe_file, 2)

	// check with peer
	fmt.Printf("\n\n* checking with peer:\n")
	go CheckPeerPipe(pipe_file, 2)
	StartInputPipe(pipe_file, 5)

	// start read handler
	fmt.Printf("\n\n* running command handler:\n")
	StartCmdHandler(pipe_file, 5)

	fmt.Printf("\n**** exit ****\n\n")
}
