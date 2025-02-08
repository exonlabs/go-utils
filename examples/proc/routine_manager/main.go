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
)

var (
	counter atomic.Int32
)

type Routine1 struct {
	*proc.RoutineHandler
	Parent *proc.RoutineManager
}

func NewRoutine1(log *logging.Logger, parent *proc.RoutineManager) *Routine1 {
	rt := &Routine1{
		Parent: parent,
	}
	rt.RoutineHandler = proc.NewRoutineHandler(log, rt)
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
		rt.Log.Info("stopping rt2 at count=%d", count)
		rt.Parent.StopRoutine("rt2")
	case 10:
		rt.Log.Info("starting rt2 at count=%d", count)
		rt.Parent.StartRoutine("rt2")
	}

	rt.Sleep(1)
	return nil
}

func (rt *Routine1) Terminate() error {
	rt.Log.Info("terminating")

	// terminate activity after 3sec
	exitSec := 3
	rt.Log.Info("exit after %d sec", exitSec)
	for i := 0; i < exitSec; i++ {
		if !rt.Sleep(1) {
			break
		}
		rt.Log.Info("term ... %d", (i + 1))
	}

	rt.Log.Info("terminated")
	return nil
}

type Routine2 struct {
	*proc.RoutineHandler
	Parent *proc.RoutineManager
}

func NewRoutine2(log *logging.Logger, parent *proc.RoutineManager) *Routine2 {
	rt := &Routine2{
		Parent: parent,
	}
	rt.RoutineHandler = proc.NewRoutineHandler(log, rt)
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
		rt.Log.Info("stopping myself at count=%d", count)
		rt.Stop()
	case 20:
		rt.Log.Info("stopping process at count=%d", count)
		rt.Parent.Stop()
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
	case *debug0:
		log.Level = logging.DEBUG
	case *debug1:
		log.Level = logging.TRACE
	}

	log.Info("**** starting ****")

	counter.Store(0)

	rm := proc.NewRoutineManager(log)
	rm.StoppingDelay = 5

	rt1 := NewRoutine1(log.ChildLogger("rt1"), rm)
	if err := rm.AddRoutine("rt1", rt1, true); err != nil {
		log.Error(err.Error())
		return
	}

	rt2 := NewRoutine2(log.ChildLogger("rt2"), rm)
	if err := rm.AddRoutine("rt2", rt2, true); err != nil {
		log.Error(err.Error())
		return
	}

	rm.Start()

	log.Info("exit")
}
