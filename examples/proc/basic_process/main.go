// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"os"
	"runtime/debug"
	"syscall"

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
	p.Process.SetSignalHandler(syscall.SIGQUIT, p.handleSigQuit)
	return p
}

func (p *SampleProcess) Initialize() error {
	p.Log.Info("initialized")
	return nil
}

func (p *SampleProcess) Execute() error {
	p.counter += 1
	p.Log.Info("running: ... %d", p.counter)

	// stop after n counts
	if p.counter >= 60 {
		p.Log.Info("exit process at count %d", p.counter)
		p.Stop()
		return nil
	}

	p.Sleep(1)
	return nil
}

func (p *SampleProcess) Terminate() error {
	p.Log.Info("terminating")

	// terminate activity after 3sec
	exitSec := 3
	p.Log.Info("exit after %d sec", exitSec)
	for i := 0; i < exitSec && !p.KillEvent.IsSet(); i++ {
		p.Sleep(1)
		p.Log.Info("term ... %d", (i + 1))
	}

	p.Log.Info("terminated")
	return nil
}

func (p *SampleProcess) handleSigQuit() {
	p.Log.Info("exit overwrite .. no wait counts")
	p.Kill()
}

func main() {
	log := logging.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			log.Panic("%s", r)
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
	case *debug0:
		log.Level = logging.DEBUG
	}

	log.Info("**** starting ****")

	p := NewSampleProcess(log)
	p.Start()
}
