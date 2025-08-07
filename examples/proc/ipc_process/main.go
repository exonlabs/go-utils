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
	"github.com/exonlabs/go-utils/pkg/tasklet"
)

var (
	ipc_sock = filepath.Join(os.TempDir(), "ipc_sock")

	tskManager *proc.TaskletManager
)

type MainWorker struct {
	*tasklet.SyncTasklet
}

func NewWorker(log *logging.Logger) *MainWorker {
	wk := &MainWorker{}
	wk.SyncTasklet = tasklet.NewSyncTasklet(log, wk)
	return wk
}

func (wk *MainWorker) Initialize() error {
	wk.Log.Info("initialized")
	return nil
}

func (wk *MainWorker) Execute() error {
	wk.Log.Info("running")
	wk.Sleep(2)
	return nil
}

func (wk *MainWorker) Terminate() error {
	wk.Log.Info("terminated")
	return nil
}

func DataHandler([]byte) ([]byte, error) {
	// 	p := strings.Split(cmd, ":")

	// 	switch strings.TrimSpace(p[0]) {
	// 	case "exit":
	// 		tskManager.Stop()

	// 	case "list_workers":
	// 		workers := tskManager.ListTasklets()
	// 		sort.Strings(workers)
	// 		res := strings.Join(workers, ",")
	// 		if len(res) > 0 {
	// 			return res
	// 		}
	// 		return "<empty>"

	// 	case "add_worker":
	// 		if (workers.Load() - wrkIndx.Load() + 1) >= 10 {
	// 			return "MAX_REACHED"
	// 		}
	// 		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
	// 		wrk := NewWorker(tskManager.Log.ChildLogger(wname))
	// 		if err := tskManager.AddTasklet(wname, wrk); err != nil {
	// 			fmt.Println(err.Error())
	// 			return "FAILED"
	// 		}
	// 		workers.Add(1)
	// 		fmt.Printf("added worker: %s\n", wname)

	// 	case "del_worker":
	// 		if wrkIndx.Load() <= workers.Load() {
	// 			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
	// 			if err := tskManager.DeleteTasklet(wname); err != nil {
	// 				fmt.Println(err.Error())
	// 				return "FAILED"
	// 			}
	// 			wrkIndx.Add(1)
	// 			fmt.Printf("deleted worker: %s\n", wname)
	// 		} else {
	// 			return "NO_WORKERS"
	// 		}

	// 	case "start_worker":
	// 		if len(p) < 2 {
	// 			return "MISSING_PARAM"
	// 		}
	// 		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
	// 		if err := tskManager.StartTasklet(wname); err != nil {
	// 			fmt.Println(err.Error())
	// 			return "FAILED"
	// 		}

	// 	case "stop_worker":
	// 		if len(p) < 2 {
	// 			return "MISSING_PARAM"
	// 		}
	// 		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
	// 		if err := tskManager.StopTasklet(wname); err != nil {
	// 			fmt.Println(err.Error())
	// 			return "FAILED"
	// 		}

	// 	case "restart_worker":
	// 		if len(p) < 2 {
	// 			return "MISSING_PARAM"
	// 		}
	// 		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
	// 		if err := tskManager.RestartTasklet(wname); err != nil {
	// 			fmt.Println(err.Error())
	// 			return "FAILED"
	// 		}

	// 	default:
	// 		return "INVALID_COMMAND"
	// 	}

	// return "DONE"
}

func main() {
	log := logging.NewStdoutLogger("main")

	commLog := logging.NewStdoutLogger("comm")
	commLog.SetFormatter(logging.RawFormatter)

	defer func() {
		if tskManager != nil {
			tskManager.Stop()
		}
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
		commLog = nil
	default:
		commLog = nil
	}

	log.Info("**** starting ****")

	tskManager = proc.NewTaskletManager(log)
	tskManager.StoppingDelay = 5

	wrk := NewWorker(log.ChildLogger("wrk"))
	wrk.Enable()
	if err := tskManager.AddTasklet("wrk", wrk); err != nil {
		log.Error(err.Error())
		return
	}

	commListener, err := sockcomm.NewListener(
		fmt.Sprintf("sock@%s", ipc_sock), commLog, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ipcListener, err := proc.NewIpcListener(
		log.ChildLogger("ipc"), commListener, DataHandler)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ipcListener.Enable()
	if err := tskManager.AddTasklet("ipc", ipcListener); err != nil {
		log.Error(err.Error())
		return
	}

	if err := tskManager.Start(); err != nil {
		log.Error(err.Error())
		return
	}

	if !tskManager.Join(5) {
		log.Warn("timeout waiting to stop ... exit anyway")
	}
	log.Info("exit")
}
