package xputil

import (
	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	defaultErrorDelay = float64(1)
)

type Tasklet interface {
	Initialize() error
	Execute() error
	Terminate() error
}

type BaseTasklet struct {
	Log     *xlog.Logger
	tasklet Tasklet
	evtTerm *xevent.Event
	evtKill *xevent.Event

	// error delay for execution loop
	ErrorDelay float64
}

func NewTaskletManager(log *xlog.Logger, tsk Tasklet) *BaseTasklet {
	return &BaseTasklet{
		Log:        log,
		tasklet:    tsk,
		evtTerm:    xevent.NewEvent(),
		evtKill:    xevent.NewEvent(),
		ErrorDelay: defaultErrorDelay,
	}
}

func (tl *BaseTasklet) IsTermEvent() bool {
	return tl.evtTerm.IsSet()
}

func (tl *BaseTasklet) IsKillEvent() bool {
	return tl.evtKill.IsSet()
}

func (tl *BaseTasklet) Start() {
	tl.evtTerm.Clear()
	tl.evtKill.Clear()

	// initialize
	if err := tl.tasklet.Initialize(); err != nil {
		tl.Log.Fatal("initialize failed, %s", err.Error())
		return
	}

	// execute operation loop forever till term event
	for !tl.evtTerm.IsSet() {
		if !tl.safeExec(tl.tasklet.Execute) {
			tl.Sleep(tl.ErrorDelay)
		}
	}

	// graceful terminate
	if !tl.evtKill.IsSet() {
		tl.safeExec(tl.tasklet.Terminate)
	}
}

func (tl *BaseTasklet) Stop() {
	if tl.evtTerm.IsSet() {
		tl.evtKill.Set()
	} else {
		tl.evtTerm.Set()
	}
}

func (tl *BaseTasklet) Kill() {
	tl.evtKill.Set()
	tl.evtTerm.Set()
}

func (tl *BaseTasklet) Sleep(timeout float64) bool {
	if tl.evtTerm.IsSet() {
		return tl.evtKill.Wait(timeout)
	} else {
		return tl.evtTerm.Wait(timeout)
	}
}

// run function with error and panic handling
func (tl *BaseTasklet) safeExec(f func() error) bool {
	err, trace := PanicExcept(f)
	if trace != "" {
		tl.Log.Panic(err.Error())
		tl.Log.Trace1("\n-------------\n%s-------------", trace)
		return false
	} else if err != nil {
		tl.Log.Error(err.Error())
		return false
	}
	return true
}
