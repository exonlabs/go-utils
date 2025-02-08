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
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/exonlabs/go-utils/pkg/comm/commutils"
	"github.com/exonlabs/go-utils/pkg/logging"
	"github.com/exonlabs/go-utils/pkg/proc"
)

var (
	wrkManager *proc.RoutineManager
	cmdSock    = filepath.Join(os.TempDir(), "wrkmanager_sock")
	workers    atomic.Int32
	wrkIndx    atomic.Int32
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
	wk.Sleep(2)
	return nil
}

func (wk *Worker) Terminate() error {
	wk.Log.Info("terminated")
	return nil
}

func HandleCommand(cmd string) string {
	p := strings.Split(cmd, ":")

	switch strings.TrimSpace(p[0]) {
	case "exit":
		go func() {
			time.Sleep(3 * time.Second)
			wrkManager.Stop()
		}()

	case "list_workers":
		workers := wrkManager.ListRoutines()
		sort.Strings(workers)
		res := strings.Join(workers, ",")
		if len(res) > 0 {
			return res
		}
		return "<empty>"

	case "add_worker":
		if (workers.Load() - wrkIndx.Load() + 1) >= 10 {
			return "MAX_REACHED"
		}
		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
		wrk := NewWorker(wrkManager.Log.ChildLogger(wname))
		if err := wrkManager.AddRoutine(wname, wrk, true); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}
		workers.Add(1)
		fmt.Printf("added worker: %s\n", wname)

	case "del_worker":
		if wrkIndx.Load() <= workers.Load() {
			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
			if err := wrkManager.DelRoutine(wname); err != nil {
				fmt.Println(err.Error())
				return "FAILED"
			}
			wrkIndx.Add(1)
			fmt.Printf("deleted worker: %s\n", wname)
		} else {
			return "NO_WORKERS"
		}

	case "start_worker":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := wrkManager.StartRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	case "stop_worker":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := wrkManager.StopRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	case "restart_worker":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := wrkManager.RestartRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	default:
		return "INVALID_COMMAND"
	}

	return "DONE"
}

func main() {
	log := logging.NewStdoutLogger("main")

	defer func() {
		if wrkManager != nil {
			wrkManager.Stop()
		}
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
	default:
	}

	log.Info("**** starting ****")

	workers.Store(3)
	wrkIndx.Store(1)

	commListener, err := commutils.NewListener(
		fmt.Sprintf("sock@%s", cmdSock), nil, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	wrkManager = proc.NewRoutineManager(log)
	wrkManager.SetCmdHandler(commListener, HandleCommand)

	for i := int32(1); i <= workers.Load(); i++ {
		wname := fmt.Sprintf("wrk%d", i)
		wrk := NewWorker(log.ChildLogger(wname))
		if err := wrkManager.AddRoutine(wname, wrk, true); err != nil {
			log.Error(err.Error())
			return
		}
	}

	wrkManager.Start()
}
