package main

import (
	"bytes"
	"flag"
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
	logger := xlog.GetLogger()
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

	pr := NewSampleProcess(logger)
	pr.ProcTitle = "SampleProcess"

	pr.Start()
}
