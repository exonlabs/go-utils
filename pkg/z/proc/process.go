package xputil

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/exonlabs/go-utils/pkg/unix/xpipe"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

// signal handling callback function
type SignalHandler func()

// command handling callback function
// takes received command and return reply message
type CommandHandler func(string) string

type BaseProcess struct {
	*BaseTasklet

	// process title as shown in OS process table
	ProcTitle string

	// buffer holding defined signal handlers
	sigHandlers map[os.Signal]SignalHandler

	// input and output pipes
	inputPipe  *xpipe.Pipe
	outputPipe *xpipe.Pipe
	// command handler
	cmdHandler CommandHandler
}

// create new base process
func NewBaseProcess(log *xlog.Logger, tsk Tasklet) *BaseProcess {
	pr := &BaseProcess{
		BaseTasklet: NewBaseTasklet(log, tsk),
	}
	pr.sigHandlers = map[os.Signal]SignalHandler{
		syscall.SIGINT:  pr.Stop,
		syscall.SIGTERM: pr.Stop,
		syscall.SIGKILL: pr.Stop,
		syscall.SIGQUIT: pr.Stop,
		syscall.SIGHUP:  pr.Stop,
	}
	return pr
}

// initialize process management
func (pr *BaseProcess) InitManagement(
	ipipe, opipe *xpipe.Pipe, hnd CommandHandler) {
	pr.inputPipe = ipipe
	pr.outputPipe = opipe
	pr.cmdHandler = hnd
}

// add signal handler
func (pr *BaseProcess) SetSignal(sig os.Signal, fn SignalHandler) {
	if sig != nil {
		pr.sigHandlers[sig] = fn
	}
}

func (pr *BaseProcess) Start() {
	// start signal handling routine
	go pr.handleSignals()

	// set process title in OS process table
	pr.ProcTitle = strings.TrimSpace(pr.ProcTitle)
	if len(pr.ProcTitle) > 0 {
		if err := SetProcTitle(pr.ProcTitle); err != nil {
			pr.Log.Warn("failed setting proctitle, %s", err.Error())
		}
	}

	// start command handling routine
	if pr.inputPipe != nil && pr.cmdHandler != nil {
		go pr.handleCommands()
	}

	pr.BaseTasklet.Start()
}

func (pr *BaseProcess) Stop() {
	if pr.inputPipe != nil {
		pr.inputPipe.Cancel()
	}
	if pr.outputPipe != nil {
		pr.outputPipe.Cancel()
	}
	pr.BaseTasklet.Stop()
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
		pr.SafeExecute(func() error {
			pr.sigHandlers[sig]()
			return nil
		})
	}
}

func (pr *BaseProcess) handleCommands() {
	for !pr.IsTermEvent() {
		b, err := pr.inputPipe.ReadWait(-1)
		if pr.IsTermEvent() {
			return
		} else if err != nil && !errors.Is(err, xpipe.ErrTimeout) {
			pr.Log.Error(err.Error())
			continue
		}

		cmd := strings.TrimSpace(string(b))
		if len(cmd) == 0 {
			continue
		}

		var reply string
		pr.SafeExecute(func() error {
			reply = pr.cmdHandler(cmd)
			return nil
		})
		reply = strings.TrimSpace(reply)

		if pr.IsKillEvent() {
			return
		} else if pr.outputPipe != nil && len(reply) > 0 {
			pr.outputPipe.WriteWait([]byte(reply+"\n"), -1)
		}
	}
}
