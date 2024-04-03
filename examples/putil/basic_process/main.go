package main

import (
	"bytes"
	"flag"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/xlog"
	"github.com/exonlabs/go-utils/pkg/xputil"
)

type SampleProcess struct {
	*xputil.BaseProcess
	counter int
}

func NewSampleProcess(log *xlog.Logger) *SampleProcess {
	pr := &SampleProcess{}
	pr.BaseProcess = xputil.NewBaseProcess(log, pr)
	pr.SetSignal(syscall.SIGUSR2, pr.handleSigUsr2)
	pr.SetSignal(syscall.SIGQUIT, pr.handleSigQuit)
	return pr
}

func (pr *SampleProcess) Initialize() error {
	pr.Log.Info("initialized")
	return nil
}

func (pr *SampleProcess) Execute() error {
	pr.counter += 1
	pr.Log.Info("running: ... %d", pr.counter)

	// stop after n counts
	if pr.counter >= 60 {
		pr.Log.Info("exit process at count %d", pr.counter)
		pr.Stop()
		return nil
	}

	pr.Sleep(1)
	return nil
}

func (pr *SampleProcess) Terminate() error {
	pr.Log.Info("terminating")

	// terminate activity after 3sec
	exitSec := 3
	pr.Log.Info("exit after %d sec", exitSec)
	for i := 0; i < exitSec && !pr.IsKillEvent(); i++ {
		pr.Sleep(1)
		pr.Log.Info("term ... %d", (i + 1))
	}

	pr.Log.Info("terminated")
	return nil
}

func (pr *SampleProcess) handleSigUsr2() {
	pr.counter = 0
	pr.Log.Info("counter reset")
}

func (pr *SampleProcess) handleSigQuit() {
	pr.Log.Info("exit overwrite .. no wait counts")
	pr.Kill()
}

func main() {
	logger := xlog.NewStdoutLogger("main")

	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			logger.Panic("%s", r)
			logger.Trace1("\n-------------\n%s-------------", stack[indx:])
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

	pr := NewSampleProcess(logger)
	pr.ProcTitle = "SampleProcess"

	pr.Start()
}
