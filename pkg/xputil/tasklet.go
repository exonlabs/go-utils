package xputil

import (
	"bytes"
	"runtime/debug"

	"github.com/exonlabs/go-utils/pkg/sync/xevent"
	"github.com/exonlabs/go-utils/pkg/xlog"
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

func NewBaseTasklet(log *xlog.Logger, tsk Tasklet) *BaseTasklet {
	return &BaseTasklet{
		Log:        log,
		tasklet:    tsk,
		evtTerm:    xevent.NewEvent(),
		evtKill:    xevent.NewEvent(),
		ErrorDelay: float64(1),
	}
}

// check if terminate event is set
func (tl *BaseTasklet) IsTermEvent() bool {
	return tl.evtTerm.IsSet()
}

// check if kill event is set
func (tl *BaseTasklet) IsKillEvent() bool {
	return tl.evtKill.IsSet()
}

// start tasklet execution
func (tl *BaseTasklet) Start() {
	tl.evtTerm.Clear()
	tl.evtKill.Clear()

	// initialize
	if !tl.SafeExecute(tl.tasklet.Initialize) {
		return
	}

	// execute operation loop forever till term event
	for !tl.evtTerm.IsSet() {
		if !tl.SafeExecute(tl.tasklet.Execute) {
			tl.Sleep(tl.ErrorDelay)
		}
	}

	// graceful terminate
	if !tl.evtKill.IsSet() {
		tl.SafeExecute(tl.tasklet.Terminate)
	}
}

// terminate then kill tasklet execution
func (tl *BaseTasklet) Stop() {
	if tl.evtTerm.IsSet() {
		tl.evtKill.Set()
	} else {
		tl.evtTerm.Set()
	}
}

// kill tasklet execution
func (tl *BaseTasklet) Kill() {
	tl.evtKill.Set()
	tl.evtTerm.Set()
}

// non-blocking sleep
func (tl *BaseTasklet) Sleep(timeout float64) bool {
	if tl.evtTerm.IsSet() {
		return tl.evtKill.Wait(timeout)
	} else {
		return tl.evtTerm.Wait(timeout)
	}
}

// run function with error and panic handling
func (tl *BaseTasklet) SafeExecute(f func() error) bool {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			indx := bytes.Index(stack, []byte("panic({"))
			tl.Log.Panic("%s", r)
			tl.Log.Trace1("\n----------\n%s----------", stack[indx:])
		}
	}()
	if err := f(); err != nil {
		tl.Log.Error(err.Error())
		return false
	}
	return true
}
