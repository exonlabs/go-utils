package xputil

import (
	"sync/atomic"
)

type Routine interface {
	Setup(string, *RtManager) error
	Parent() *RtManager
	IsAlive() bool
	Start()
	Stop()
	Kill()
}

type BaseRoutine struct {
	*BaseTasklet
	Name   string
	parent *RtManager
	alive  atomic.Bool
}

func NewBaseRoutine(tsk Tasklet) *BaseRoutine {
	return &BaseRoutine{
		BaseTasklet: NewTaskletManager(nil, tsk),
	}
}

func (rt *BaseRoutine) Setup(name string, rm *RtManager) error {
	rt.Name = name
	rt.parent = rm
	rt.Log = rm.Log.ChildLogger(name)
	return nil
}

func (rt *BaseRoutine) Parent() *RtManager {
	return rt.parent
}

func (rt *BaseRoutine) IsAlive() bool {
	return rt.alive.Load()
}

func (rt *BaseRoutine) Start() {
	// set alive status flag and clear when exit
	rt.alive.Store(true)
	defer rt.alive.Store(false)

	// start the tasklet manager
	rt.BaseTasklet.Start()
}
