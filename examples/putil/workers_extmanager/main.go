package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

var (
	workers atomic.Int32
	wrkIndx atomic.Int32
)

type Worker struct {
	*xputil.BaseRoutine
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
	wk.Log.Info("running")
	wk.Sleep(2)
	return nil
}

func (wk *Worker) Terminate() error {
	wk.Log.Info("terminated")
	return nil
}

func CommandHandler(rm *xputil.RtManager, cmd string) (string, error) {
	p := strings.Split(cmd, ":")

	switch strings.TrimSpace(p[0]) {
	case "EXIT":
		rm.Stop()

	case "LIST_WORKERS":
		return strings.Join(rm.ListRoutines(), ","), nil

	case "ADD_WORKER":
		if (workers.Load() - wrkIndx.Load() + 1) >= 10 {
			return "MAX_REACHED", nil
		}
		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			return "FAILED", err
		}
		workers.Add(1)

	case "DEL_WORKER":
		if wrkIndx.Load() <= workers.Load() {
			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
			if err := rm.DelRoutine(wname); err != nil {
				return "FAILED", err
			}
			wrkIndx.Add(1)
		} else {
			return "NO_WORKERS", nil
		}

	case "START_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM", nil
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.StartRoutine(wname); err != nil {
			return "FAILED", err
		}

	case "STOP_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM", nil
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.StopRoutine(wname); err != nil {
			return "FAILED", err
		}

	case "RESTART_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM", nil
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.RestartRoutine(wname); err != nil {
			return "FAILED", err
		}

	default:
		return "INVALID_COMMAND", nil
	}

	return "DONE", nil
}

func main() {
	logger := xlog.NewStdoutLogger("main")
	logger.SetFormatter(xlog.NewStdFrmtWithSrc())

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

	tmpPath := filepath.Join(os.TempDir(), "foobar")
	logger.Info("Using Pipes Dir: %s", tmpPath)

	syscall.Umask(0)
	os.RemoveAll(tmpPath)
	os.MkdirAll(tmpPath, 0o777)
	defer os.RemoveAll(tmpPath)

	rm := xputil.NewExtRtManager(logger)
	rm.ProcTitle = "WrkManager"
	rm.PipeDir = tmpPath
	rm.CommandHandler = CommandHandler

	for i := int32(1); i <= workers.Load(); i++ {
		wname := fmt.Sprintf("wrk%d", i)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			logger.Error(err.Error())
			return
		}
	}

	rm.Start()
}
