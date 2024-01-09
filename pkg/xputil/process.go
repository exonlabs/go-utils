package xputil

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

type SignalHandler func()

type BaseProcess struct {
	*BaseTasklet
	sigHandlers map[os.Signal]SignalHandler

	// process title as shown in OS process table
	ProcTitle string
}

func NewBaseProcess(log *xlog.Logger, tsk Tasklet) *BaseProcess {
	pr := &BaseProcess{
		BaseTasklet: NewTaskletManager(log, tsk),
	}
	pr.sigHandlers = map[os.Signal]SignalHandler{
		syscall.SIGINT:  pr.Stop,
		syscall.SIGTERM: pr.Stop,
		syscall.SIGQUIT: pr.Stop,
		syscall.SIGHUP:  pr.Stop,
	}
	return pr
}

func (pr *BaseProcess) SetSignal(sig os.Signal, fn SignalHandler) {
	pr.sigHandlers[sig] = fn
}

func (pr *BaseProcess) Start() {
	// start signal handler routine
	go pr.handleSignals()

	// set process title in OS process table
	pr.ProcTitle = strings.TrimSpace(pr.ProcTitle)
	if len(pr.ProcTitle) > 0 {
		if err := SetProcTitle(pr.ProcTitle); err != nil {
			pr.Log.Warn("failed setting proctitle, %s", err.Error())
		}
	}

	// start the tasklet manager
	pr.BaseTasklet.Start()
}

func (pr *BaseProcess) handleSignals() {
	// register callback for defined signals
	sigCh := make(chan os.Signal, 1)
	for sig, fn := range pr.sigHandlers {
		if fn != nil {
			signal.Notify(sigCh, sig)
		}
	}

	// handle signals
	for {
		sig := <-sigCh
		pr.Log.Debug("<received signal: %s>", sig)
		pr.safeExec(func() error {
			pr.sigHandlers[sig]()
			return nil
		})
	}
}
