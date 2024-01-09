package xevent

import (
	"sync/atomic"
	"time"
)

type Event struct {
	state atomic.Bool
	setch chan bool
}

func NewEvent() *Event {
	e := &Event{}
	e.state.Store(false)
	e.setch = make(chan bool, 1)
	return e
}

func (e *Event) Set() {
	e.state.Store(true)
	if len(e.setch) == 0 {
		e.setch <- true
	}
}

func (e *Event) Clear() {
	e.state.Store(false)
	if len(e.setch) > 0 {
		<-e.setch
	}
}

func (e *Event) IsSet() bool {
	return e.state.Load()
}

func (e *Event) Wait(timeout float64) bool {
	select {
	case <-time.After(time.Duration(timeout * 1000000000)):
		return true
	case <-e.setch:
	}
	return false
}
