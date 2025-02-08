// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"sync/atomic"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/proc"
)

var (
	workers atomic.Int32
	wrkIndx atomic.Int32

	MaxWorkers = int32(10)
)

type Worker struct {
	*proc.RoutineHandler
}

func NewWorker(log *logging.Logger) *Worker {
	wk := &Worker{}
	wk.RoutineHandler = proc.NewRoutineHandler(log, wk)
	return wk
}

func (wk *Worker) Initialize() error {
	wk.Log.Info("initialized")
	return nil
}

func (wk *Worker) Execute() error {
	wk.Log.Info("running")
	if rand.Intn(10) >= 8 {
		wk.Log.Info("closing myself")
		wk.Stop()
		return nil
	}
	wk.Sleep(2)
	return nil
}

func (wk *Worker) Terminate() error {
	wk.Log.Info("terminated")
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

	workers.Store(3)
	wrkIndx.Store(1)

	rm := proc.NewRoutineManager(log)

	for i := int32(1); i <= workers.Load(); i++ {
		wname := fmt.Sprintf("wrk%d", i)
		wrk := NewWorker(log.ChildLogger(wname))
		if err := rm.AddRoutine(wname, wrk, true); err != nil {
			log.Error(err.Error())
			return
		}
	}

	rm.SetSignalHandler(syscall.SIGUSR1, func() {
		log.Info("adding worker")
		if (workers.Load() - wrkIndx.Load() + 1) >= MaxWorkers {
			log.Info("max concurrent workers")
			return
		}
		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
		wrk := NewWorker(log.ChildLogger(wname))
		log.Info("adding new worker: %s", wname)
		if err := rm.AddRoutine(wname, wrk, true); err != nil {
			log.Error(
				"failed adding worker: %s - %s", wname, err.Error())
			return
		}
		workers.Add(1)
	})

	rm.SetSignalHandler(syscall.SIGUSR2, func() {
		log.Info("deleting worker")
		if wrkIndx.Load() <= workers.Load() {
			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
			log.Info("deleting worker: %s", wname)
			if err := rm.DelRoutine(wname); err != nil {
				log.Error(
					"failed deleting worker: %s - %s", wname, err.Error())
				return
			}
			wrkIndx.Add(1)
		} else {
			log.Info("no workers to delete")
		}
	})

	rm.Start()

	log.Info("exit")
}
