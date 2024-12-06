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
	"time"

	"github.com/exonlabs/go-utils/pkg/comm/commutils"
	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/proc"
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
		go func() {
			time.Sleep(10 * time.Millisecond)
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

	uri := fmt.Sprintf("sock@%s", filepath.Join(os.TempDir(), "process_sock"))
	commLog := logging.NewStdoutLogger("comm")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			log.Error("%s", r)
			log.Trace1("\n----------\n%s----------", stack[indx:])
			os.Exit(1)
		} else {
			log.Info("exit")
			os.Exit(0)
		}
	}()

	debug0 := flag.Bool("x", false, "\nenable debug logs")
	debug1 := flag.Bool("xx", false, "enable debug and trace1 logs")
	debug2 := flag.Bool("xxx", false, "enable debug and trace2 logs")
	debug3 := flag.Bool("xxxx", false, "enable debug and trace3 logs")
	flag.Parse()

	switch {
	case *debug3:
		log.Level = logging.TRACE3
	case *debug2:
		log.Level = logging.TRACE2
	case *debug1:
		log.Level = logging.TRACE1
		commLog = nil
	case *debug0:
		log.Level = logging.DEBUG
		commLog = nil
	default:
		commLog = nil
	}

	log.Info("**** starting ****")

	commListener, err := commutils.NewListener(uri, commLog, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	p := NewSampleProcess(log)
	p.SetCmdHandler(commListener, p.HandleCommand)
	p.Start()
}
