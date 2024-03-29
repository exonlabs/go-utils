package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

var (
	workers atomic.Int32
	wrkIndx atomic.Int32
)

type Worker struct {
	*xputil.BaseRoutine
	close bool
}

func NewWorker() *Worker {
	wk := &Worker{}
	wk.BaseRoutine = xputil.NewBaseRoutine(wk)
	return wk
}

func (wk *Worker) Initialize() error {
	wk.Log.Info("initialized")
	return nil
}

func (wk *Worker) Execute() error {
	if rand.Intn(10) >= 8 {
		wk.Log.Info("closing myself")
		wk.close = true
		wk.Stop()
		return nil
	}
	wk.Log.Info("running")
	wk.Sleep(2)
	return nil
}

func (wk *Worker) Terminate() error {
	if !wk.close && rand.Intn(10) >= 5 {
		wk.Log.Info("i will not exit")
		for {
			time.Sleep(time.Second * 10)
			break
		}
	}
	wk.Log.Info("terminated")
	return nil
}

func main() {
	logger := xlog.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			logger.Panic("%s", r)
			logger.Trace1("\n-------------\n%s-------------", stack[indx:])
			logger.Warn("exit ... due to last error")
		} else {
			logger.Info("exit")
		}
	}()

	debugOpt := flag.Int("x", 0, "set debug modes, (default: 0)")
	flag.Parse()

	if *debugOpt > 0 {
		switch *debugOpt {
		case 1:
			logger.Level = xlog.DEBUG
		case 2:
			logger.Level = xlog.TRACE1
		default:
			logger.Level = xlog.TRACE2
		}
	}

	logger.Info("**** starting ****")

	workers.Store(3)
	wrkIndx.Store(1)

	rm := xputil.NewRtManager(logger)
	rm.ProcTitle = "WrkManager"

	for i := int32(1); i <= workers.Load(); i++ {
		wname := fmt.Sprintf("wrk%d", i)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			logger.Error(err.Error())
			return
		}
	}

	rm.SetSignal(syscall.SIGUSR1, func() {
		logger.Info("adding worker")
		if (workers.Load() - wrkIndx.Load() + 1) >= 10 {
			logger.Info("max concurrent workers")
			return
		}
		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
		logger.Info("adding new worker: %s", wname)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			logger.Error(
				"failed adding worker: %s - %s", wname, err.Error())
			return
		}
		workers.Add(1)
	})

	rm.SetSignal(syscall.SIGUSR2, func() {
		logger.Info("deleting worker")
		if wrkIndx.Load() <= workers.Load() {
			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
			logger.Info("deleting worker: %s", wname)
			if err := rm.DelRoutine(wname); err != nil {
				logger.Error(
					"failed deleting worker: %s - %s", wname, err.Error())
				return
			}
			wrkIndx.Add(1)
		} else {
			logger.Info("no workers to delete")
		}
	})

	rm.Start()
}
