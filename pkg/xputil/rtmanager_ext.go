package xputil

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exonlabs/go-utils/pkg/unix/xpipe"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	defaultPollInterval = float64(0.5)
)

// command handling callback function definition
// args: manager handler, received command
// return: reply message and error
type CommandHandler func(string) (string, error)

// extended routine manager with management through named pipes
// manager uses 2 pipes: 1 for input and 1 for output
type ExtRtManager struct {
	*RtManager

	// management pipes path and handlers
	PipeDir string
	inPipe  *xpipe.Pipe
	outPipe *xpipe.Pipe

	// last routine monitoring timestamp
	lastMonTs float64
	// pipe read polling delay
	PollInterval float64

	// command handler callback function
	CommandHandler CommandHandler
}

func NewExtRtManager(log *xlog.Logger) *ExtRtManager {
	rm := &ExtRtManager{
		RtManager:    NewRtManager(log),
		PollInterval: defaultPollInterval,
	}
	rm.BaseProcess = NewBaseProcess(log, rm)
	return rm
}

func (rm *ExtRtManager) Initialize() error {
	if rm.PipeDir != "" && rm.CommandHandler != nil {
		if _, err := os.Stat(rm.PipeDir); os.IsNotExist(err) {
			return fmt.Errorf("pipes dir does not exist")
		}
		rm.inPipe = xpipe.NewPipe(fmt.Sprintf("%s/in.pipe", rm.PipeDir))
		if err := rm.inPipe.Create(0o666); err != nil {
			return fmt.Errorf(
				"failed creating input pipe, %s", err.Error())
		}
		rm.outPipe = xpipe.NewPipe(fmt.Sprintf("%s/out.pipe", rm.PipeDir))
		if err := rm.outPipe.Create(0o666); err != nil {
			rm.inPipe.Delete()
			return fmt.Errorf(
				"failed creating output pipe, %s", err.Error())
		}
		rm.inPipe.OpenRead()
		rm.outPipe.OpenWrite()
	}
	return rm.RtManager.Initialize()
}

func (rm *ExtRtManager) Execute() error {
	// if no command handler or pipes defined
	if rm.PipeDir == "" || rm.CommandHandler == nil {
		return rm.RtManager.Execute()
	}

	// check routines at defined intervals
	if (float64(time.Now().Unix()) - rm.lastMonTs) >= rm.MonInterval {
		rm.CheckRoutines()
		rm.lastMonTs = float64(time.Now().Unix())
	}

	// checking commands
	rm.CheckCommand()

	return nil
}

func (rm *ExtRtManager) Terminate() error {
	if rm.inPipe != nil {
		rm.inPipe.Close()
		rm.inPipe.Delete()
	}
	if rm.outPipe != nil {
		rm.outPipe.Close()
		rm.outPipe.Delete()
	}
	return rm.RtManager.Terminate()
}

func (rm *ExtRtManager) CheckCommand() error {
	rm.Log.Trace1("checking commands ...")

	b, err := rm.inPipe.ReadWait(rm.PollInterval)
	if err != nil && !errors.Is(err, xpipe.ErrTimeout) {
		rm.Log.Error(err.Error())
		return nil
	}
	if len(b) == 0 {
		return nil
	}

	cmd := strings.TrimSpace(string(b))
	rm.Log.Info("CMD_REQ: %s", cmd)

	reply, err := rm.CommandHandler(cmd)
	if err != nil {
		rm.Log.Error(err.Error())
		if len(reply) == 0 {
			reply = "INTERNAL_ERROR"
		}
	} else if len(reply) == 0 {
		reply = "DONE"
	}
	reply = strings.TrimSpace(reply)
	rm.Log.Info("CMD_RES: %s", reply)

	rm.outPipe.WriteWait([]byte(reply+"\n"), rm.PollInterval)
	return nil
}
