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
	p.Process = proc.NewProcess(log, p)

	// set custom signal handler
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

	p.Sleep(0.2)
	return nil
}

func (p *SampleProcess) Terminate() error {
	p.Log.Info("terminating")

	// terminate activity after few seconds
	exitSec := 5
	for i := exitSec; i > 0; i-- {
		p.Log.Info("exit after %d sec", i)
		p.Sleep(1)
	}

	p.Log.Info("terminated")
	return nil
}

func (p *SampleProcess) handleSigQuit() error {
	p.Log.Info("exit .. no wait counts")
	os.Exit(0)
	return nil
}

func main() {
	log := logging.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			log.Panic("%v\n----------\n%s----------", r, stack[indx:])
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
	}

	log.Info("**** starting ****")

	p := NewSampleProcess(log)
	if err := p.Start(); err != nil {
		log.Error(err.Error())
		return
	}

	if !p.Join(10) {
		log.Warn("timeout waiting to stop ... exit anyway")
	}
	log.Info("exit")
}
