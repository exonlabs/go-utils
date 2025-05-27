// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"os"
	"runtime/debug"
	"sync/atomic"

	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/proc"
	"github.com/exonlabs/go-utils/pkg/tasklet"
)

// global counter
var counter atomic.Int32

type Routine1 struct {
	*tasklet.SyncTasklet
	Manager *proc.TaskletManager
}

func NewRoutine1(log *logging.Logger, rm *proc.TaskletManager) *Routine1 {
	rt := &Routine1{
		Manager: rm,
	}
	rt.SyncTasklet = tasklet.NewSyncTasklet(log, rt)
	return rt
}

func (rt *Routine1) Initialize() error {
	rt.Log.Info("initialized")
	return nil
}

func (rt *Routine1) Execute() error {
	counter.Add(1)

	count := counter.Load()
	rt.Log.Info("new counter = %d", count)

	switch count {
	case 5:
		rt.Log.Info("stop rt2 at count=%d", count)
		rt.Manager.StopTasklet("rt2")
	case 10:
		rt.Log.Info("start rt2 at count=%d", count)
		rt.Manager.StartTasklet("rt2")
	}

	rt.Sleep(1)
	return nil
}

func (rt *Routine1) Terminate() error {
	rt.Log.Info("terminating")

	// terminate activity after few seconds
	exitSec := 3
	for i := exitSec; i > 0; i-- {
		rt.Log.Info("exit after %d sec", i)
		rt.Manager.Sleep(1)
	}

	rt.Log.Info("terminated")
	return nil
}

type Routine2 struct {
	*tasklet.SyncTasklet
	Manager *proc.TaskletManager
}

func NewRoutine2(log *logging.Logger, rm *proc.TaskletManager) *Routine2 {
	rt := &Routine2{
		Manager: rm,
	}
	rt.SyncTasklet = tasklet.NewSyncTasklet(log, rt)
	return rt
}

func (rt *Routine2) Initialize() error {
	rt.Log.Info("initialized")
	return nil
}

func (rt *Routine2) Execute() error {
	count := counter.Load()
	rt.Log.Info("monitoring: counter = %d", count)

	switch count {
	case 15:
		rt.Log.Info("reset myself at count=%d", count)
		rt.Sleep(1)
		go rt.Stop()
	case 20:
		rt.Log.Info("stopping process at count=%d", count)
		go rt.Manager.Stop()
	}

	rt.Sleep(0.5)
	return nil
}

func (rt *Routine2) Terminate() error {
	rt.Log.Info("terminated")
	return nil
}

func main() {
	log := logging.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			log.Panic("%s", r)
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
	}

	log.Info("**** starting ****")

	counter.Store(0)

	rm := proc.NewTaskletManager(log)
	rm.StoppingDelay = 10

	rt1 := NewRoutine1(log.ChildLogger("rt1"), rm)
	rt1.Enable()
	if err := rm.AddTasklet("rt1", rt1); err != nil {
		log.Error(err.Error())
		return
	}

	rt2 := NewRoutine2(log.ChildLogger("rt2"), rm)
	rt2.Enable()
	if err := rm.AddTasklet("rt2", rt2); err != nil {
		log.Error(err.Error())
		return
	}

	if err := rm.Start(); err != nil {
		log.Error(err.Error())
		return
	}

	if !rm.Join(10) {
		log.Warn("timeout waiting to stop ... exit anyway")
	}
	log.Info("exit")
}
