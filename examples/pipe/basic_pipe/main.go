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

var wg sync.WaitGroup

func SenderPrint(msg string, args ...any) {
	fmt.Printf("                         "+msg+"\n", args...)
}

func CheckPeerPipe(path string, timeout float64) {
	SenderPrint("-- starting sender")
	p := xpipe.NewPipe(path)
	msg := "HELLO"
	SenderPrint("<< sending: " + msg)
	if err := p.WriteWait([]byte(msg), timeout); err == nil {
		SenderPrint("-- CONNECTED")
	} else if errors.Is(err, xpipe.ErrTimeout) {
		SenderPrint("-- TIMEOUT: no peer connected")
	} else {
		SenderPrint("-- FAILED: %s", err.Error())
	}
}

func StartInputPipe(path string, timeout float64) {
	fmt.Printf("open input pipe\n")
	p := xpipe.NewPipe(path)
	if err := p.Create(0o666); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer p.Delete()

	b, err := p.ReadWait(timeout)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf(">> received: %s\n", b)
}

func SendingMsgs(path string, timeout float64) {
	SenderPrint("-- starting message sender")
	p := xpipe.NewPipe(path)
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("MESSAGE_%d", i)
		SenderPrint("<< sending: " + msg)
		if err := p.WriteWait([]byte(msg), timeout); err == nil {
			SenderPrint("-- DONE")
		} else {
			SenderPrint("-- FAILED: %s", err.Error())
		}
		time.Sleep(time.Millisecond * 500)
	}
	SenderPrint("-- finished sending")
}

func StartCmdHandler(path string, timeout float64) {
	fmt.Printf("open input pipe\n")
	p := xpipe.NewPipe(path)
	if err := p.Create(0o666); err != nil {
		fmt.Println(err.Error())
		return
	}
	defer p.Delete()

	evtClose := xevent.NewEvent()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for !evtClose.IsSet() {
			b, err := p.ReadWait(timeout)
			if err == nil {
				fmt.Printf(">> received: %s\n", b)
			} else if !errors.Is(err, xpipe.ErrBreak) {
				fmt.Println(err.Error())
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

	tmpPath := filepath.Join(os.TempDir(), "foobar")
	pipeFile := filepath.Join(tmpPath, "foo.pipe")

	fmt.Printf("\nUsing Pipe: %s\n", pipeFile)

	syscall.Umask(0)
	os.RemoveAll(tmpPath)
	os.MkdirAll(tmpPath, 0o777)
	defer os.RemoveAll(tmpPath)

	// check without peer
	fmt.Printf("\n* checking with no peer:\n")
	CheckPeerPipe(pipeFile, 2)

	// check with peer
	fmt.Printf("\n* checking with peer:\n")
	go CheckPeerPipe(pipeFile, 2)
	StartInputPipe(pipeFile, 5)

	// start read handler
	fmt.Printf("\nrunning command handler:\n")
	StartCmdHandler(pipeFile, 5)

	fmt.Printf("\n**** exit ****\n\n")
}
