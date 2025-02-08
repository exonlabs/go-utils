// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/exonlabs/go-utils/pkg/comm/sockcomm"
	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/proc"
)

var (
	manage_sock = filepath.Join(os.TempDir(), "manage_sock")
)

type SampleProcess struct {
	*proc.Process
	counter int
}

func NewSampleProcess(log *logging.Logger) *SampleProcess {
	p := &SampleProcess{}
	p.Process = proc.NewProcessHandler(log, p)
	return p
}

func (p *SampleProcess) Initialize() error {
	p.Log.Info("initialized")
	return nil
}

func (p *SampleProcess) Execute() error {
	p.counter += 1
	p.Log.Info("running: ... %d", p.counter)
	p.Sleep(1)
	return nil
}

func (p *SampleProcess) Terminate() error {
	p.Log.Info("terminated")
	return nil
}

func (p *SampleProcess) HandleCommand(cmd string) string {
	p.Log.Info("received command: %s", cmd)

	reply := "done"

	switch cmd {
	case "exit":
		p.Log.Info("stopping after 5 sec")
		go func() {
			p.Sleep(5)
			p.Stop()
		}()
	case "reset":
		p.counter = 0
	default:
		reply = "invalid_command"
	}

	p.Log.Info("reply command: %s", reply)
	return reply
}

func main() {
	log := logging.NewStdoutLogger("main")

	commLog := logging.NewStdoutLogger("comm")
	commLog.SetFormatter(logging.RawFormatter)

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			log.Error("%s", r)
			log.Trace("\n----------\n%s----------", stack[indx:])
			os.Exit(1)
		}
	}()

	debug0 := flag.Bool("x", false, "\nenable debug logs")
	debug1 := flag.Bool("xx", false, "enable debug and trace logs")
	flag.Parse()

	switch {
	case *debug1:
		log.Level = logging.TRACE
	case *debug0:
		log.Level = logging.DEBUG
		commLog = nil
	default:
		commLog = nil
	}

	log.Info("**** starting ****")

	commListener, err := sockcomm.NewListener(
		fmt.Sprintf("sock@%s", manage_sock), commLog, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	p := NewSampleProcess(log)
	p.SetCmdHandler(commListener, p.HandleCommand)
	p.Start()

	log.Info("exit")
}
