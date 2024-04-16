package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/unix/xpipe"
	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

var (
	tmp_path     = filepath.Join(os.TempDir(), "WrkManager")
	inpipe_file  = filepath.Join(tmp_path, "in.pipe")
	outpipe_file = filepath.Join(tmp_path, "out.pipe")
	workers      atomic.Int32
	wrkIndx      atomic.Int32
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

type RtManager struct {
	*xputil.RtManager
}

func NewRtManager(log *xlog.Logger) *RtManager {
	return &RtManager{xputil.NewRtManager(log)}
}

func (rm *RtManager) HandleCommand(cmd string) string {
	p := strings.Split(cmd, ":")

	switch strings.TrimSpace(p[0]) {
	case "EXIT":
		rm.Stop()

	case "LIST_WORKERS":
		workers := rm.ListRoutines()
		slices.Sort(workers)
		res := strings.Join(workers, ",")
		if len(res) > 0 {
			return res
		}
		return "<empty>"

	case "ADD_WORKER":
		if (workers.Load() - wrkIndx.Load() + 1) >= 10 {
			return "MAX_REACHED"
		}
		wname := fmt.Sprintf("wrk%d", workers.Load()+1)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}
		workers.Add(1)

	case "DEL_WORKER":
		if wrkIndx.Load() <= workers.Load() {
			wname := fmt.Sprintf("wrk%d", wrkIndx.Load())
			if err := rm.DelRoutine(wname); err != nil {
				fmt.Println(err.Error())
				return "FAILED"
			}
			wrkIndx.Add(1)
		} else {
			return "NO_WORKERS"
		}

	case "START_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.StartRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	case "STOP_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.StopRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	case "RESTART_WORKER":
		if len(p) < 2 {
			return "MISSING_PARAM"
		}
		wname := fmt.Sprintf("wrk%s", strings.TrimSpace(p[1]))
		if err := rm.RestartRoutine(wname); err != nil {
			fmt.Println(err.Error())
			return "FAILED"
		}

	default:
		return "INVALID_COMMAND"
	}

	return "DONE"
}

func main() {
	logger := xlog.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			logger.Panic("%s", r)
			logger.Trace1("\n----------\n%s----------", stack[indx:])
			os.Exit(1)
		} else {
			logger.Info("exit")
			os.Exit(0)
		}
	}()

	debug := flag.Int("x", 0, "set debug modes, (default: 0)")
	flag.Parse()

	switch {
	case *debug >= 5:
		logger.Level = xlog.TRACE4
	case *debug >= 4:
		logger.Level = xlog.TRACE3
	case *debug >= 3:
		logger.Level = xlog.TRACE2
	case *debug >= 2:
		logger.Level = xlog.TRACE1
	case *debug >= 1:
		logger.Level = xlog.DEBUG
	}

	logger.Info("**** starting ****")

	syscall.Umask(0)
	os.MkdirAll(tmp_path, 0o775)
	defer os.RemoveAll(tmp_path)

	logger.Info("Using Input Pipe: %s", inpipe_file)
	if err := xpipe.CreatePipe(inpipe_file, 0o666); err != nil {
		fmt.Println(err.Error())
		return
	}
	logger.Info("Using Output Pipe: %s", outpipe_file)
	if err := xpipe.CreatePipe(outpipe_file, 0o666); err != nil {
		fmt.Println(err.Error())
		return
	}

	workers.Store(3)
	wrkIndx.Store(1)

	rm := NewRtManager(logger)
	rm.ProcTitle = "WrkManager"
	rm.InitManagement(
		xpipe.NewPipe(inpipe_file),
		xpipe.NewPipe(outpipe_file),
		rm.HandleCommand,
	)

	for i := int32(1); i <= workers.Load(); i++ {
		wname := fmt.Sprintf("wrk%d", i)
		if err := rm.AddRoutine(wname, NewWorker(), true); err != nil {
			logger.Error(err.Error())
			return
		}
	}

	rm.Start()
}
