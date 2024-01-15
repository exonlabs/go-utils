package main

import (
	"bytes"
	"flag"
	"runtime/debug"
	"sync/atomic"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

var (
	counter   atomic.Int32
	exitcount atomic.Int32
)

type Routine1 struct {
	*xputil.BaseRoutine
}

func NewRoutine1() *Routine1 {
	rt := &Routine1{}
	rt.BaseRoutine = xputil.NewBaseRoutine(rt)
	return rt
}

func (r *Routine1) Initialize() error {
	r.Log.Info("initialized")
	return nil
}

func (rt *Routine1) Execute() error {
	counter.Add(1)

	count := counter.Load()
	rt.Log.Info("new counter = %d", count)

	switch count {
	case 5:
		rt.Log.Info("stopping rt2 at count=%d", count)
		rt.Parent().StopRoutine("rt2")
	case 10:
		rt.Log.Info("starting rt2 at count=%d", count)
		rt.Parent().StartRoutine("rt2")
	}

	rt.Sleep(1)
	return nil
}

func (rt *Routine1) Terminate() error {
	rt.Log.Info("terminating")

	// terminate activity after 3sec
	exitSec := 3
	rt.Log.Info("exit after %d sec", exitSec)
	for i := 0; i < exitSec && !rt.IsKillEvent(); i++ {
		rt.Sleep(1)
		rt.Log.Info("term ... %d", (i + 1))
	}

	rt.Log.Info("terminated")
	return nil
}

type Routine2 struct {
	*xputil.BaseRoutine
}

func NewRoutine2() *Routine2 {
	rt := &Routine2{}
	rt.BaseRoutine = xputil.NewBaseRoutine(rt)
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
		rt.Parent().Stop()
	}

	rt.Sleep(0.5)
	return nil
}

func (rt *Routine2) Terminate() error {
	rt.Log.Info("terminated")
	return nil
}

func main() {
	logger := xlog.GetRootLogger()
	logger.Name = "main"
	logger.SetFormatter(xlog.NewStdFormatter(
		"{time} {level} [{source}] {message}",
		"2006-01-02 15:04:05.000000"))

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

	counter.Store(0)

	rm := xputil.NewRtManager(logger)
	rm.ProcTitle = "RtManager"
	rm.MonInterval = 10
	rm.TermDelay = 5

	if err := rm.AddRoutine("rt1", NewRoutine1(), true); err != nil {
		logger.Error(err.Error())
		return
	}
	if err := rm.AddRoutine("rt2", NewRoutine2(), true); err != nil {
		logger.Error(err.Error())
		return
	}

	rm.Start()
}
