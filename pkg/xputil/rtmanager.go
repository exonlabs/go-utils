package xputil

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

const (
	defaultMonInterval = float64(5)
	defaultTermDelay   = float64(3)
)

var (
	ErrRtError          = errors.New("")
	ErrNoRoutines       = fmt.Errorf("%wno routines loaded", ErrRtError)
	ErrInvalidRoutine   = fmt.Errorf("%winvalid routine name", ErrRtError)
	ErrDuplicateRoutine = fmt.Errorf("%wduplicate routine name", ErrRtError)
	ErrRoutineActive    = fmt.Errorf("%wroutine is active", ErrRtError)
)

type rtInfo struct {
	rt     Routine
	active bool
}

type RtManager struct {
	*BaseProcess
	rtBuffer   map[string]*rtInfo
	rtBuffLock sync.Mutex

	// routines monitoring interval in sec
	MonInterval float64
	// delay to wait for routine terminate in sec
	TermDelay float64
}

func NewRtManager(log *xlog.Logger) *RtManager {
	rm := &RtManager{
		rtBuffer:    make(map[string]*rtInfo),
		MonInterval: defaultMonInterval,
		TermDelay:   defaultTermDelay,
	}
	rm.BaseProcess = NewBaseProcess(log, rm)
	return rm
}

// initialize manager
func (rm *RtManager) Initialize() error {
	if len(rm.rtBuffer) == 0 {
		return ErrNoRoutines
	}
	rm.Log.Debug("loaded routines: %s",
		strings.Join(rm.ListRoutines(), ","))
	return nil
}

// run manager operations
func (rm *RtManager) Execute() error {
	rm.CheckRoutines()
	rm.Sleep(rm.MonInterval)
	return nil
}

// terminate manager
func (rm *RtManager) Terminate() error {
	defer func() {
		if !rm.rtBuffLock.TryLock() {
			rm.rtBuffLock.Unlock()
		}
	}()

	rm.Log.Info("stopping all active routines")
	rm.rtBuffLock.Lock()
	for n := range rm.rtBuffer {
		rm.rtBuffer[n].active = false
		if rm.rtBuffer[n].rt.IsAlive() {
			rm.Log.Info("stopping routine: %s", n)
			rm.rtBuffer[n].rt.Stop()
		}
	}
	rm.rtBuffLock.Unlock()

	// no wait required
	if rm.TermDelay <= 0 {
		return nil
	}

	// check and wait all routines exit
	tPoll := float64(0.1)
	for t := rm.TermDelay; t > 0 && !rm.IsKillEvent(); t -= tPoll {
		rm.Sleep(tPoll)
		chk := true
		rm.rtBuffLock.Lock()
		for n := range rm.rtBuffer {
			if rm.rtBuffer[n].rt.IsAlive() {
				chk = false
				break
			}
		}
		rm.rtBuffLock.Unlock()
		if chk {
			return nil
		}
	}

	names := []string{}
	rm.rtBuffLock.Lock()
	for n := range rm.rtBuffer {
		if rm.rtBuffer[n].rt.IsAlive() {
			names = append(names, n)
		}
	}
	rm.rtBuffLock.Unlock()
	rm.Log.Error("failed stopping routines: %s", strings.Join(names, ","))
	return nil
}

// return name list of all current loaded routines
func (rm *RtManager) ListRoutines() []string {
	rm.rtBuffLock.Lock()
	defer rm.rtBuffLock.Unlock()

	names := []string{}
	for n := range rm.rtBuffer {
		names = append(names, n)
	}
	return names
}

// add new routine handler to manager
func (rm *RtManager) AddRoutine(name string, rt Routine, active bool) error {
	rm.rtBuffLock.Lock()
	defer rm.rtBuffLock.Unlock()

	if _, ok := rm.rtBuffer[name]; ok {
		return ErrDuplicateRoutine
	}
	if err := rt.Setup(name, rm); err != nil {
		return fmt.Errorf(
			"%wroutine setup failed, %s", ErrRtError, err.Error())
	}
	rm.Log.Trace1("adding routine: %s", name)
	rm.rtBuffer[name] = &rtInfo{rt, active}
	return nil
}

// delete routine handler from manager
func (rm *RtManager) DelRoutine(name string) error {
	rm.rtBuffLock.Lock()
	defer rm.rtBuffLock.Unlock()

	if _, ok := rm.rtBuffer[name]; !ok {
		return ErrInvalidRoutine
	}
	rm.rtBuffer[name].active = false
	rm.rtBuffer[name].rt.Kill()
	rm.Sleep(0.2)
	if rm.rtBuffer[name].rt.IsAlive() {
		rm.Sleep(1)
		if rm.rtBuffer[name].rt.IsAlive() {
			return ErrRoutineActive
		}
	}
	rm.Log.Trace1("deleting routine: %s", name)
	delete(rm.rtBuffer, name)
	return nil
}

// start routine
func (rm *RtManager) StartRoutine(name string) error {
	rm.rtBuffLock.Lock()
	defer func() {
		rm.rtBuffLock.Unlock()
		rm.CheckRoutines()
	}()

	if _, ok := rm.rtBuffer[name]; !ok {
		return ErrInvalidRoutine
	}
	rm.Log.Trace1("activating routine: %s", name)
	rm.rtBuffer[name].active = true
	return nil
}

// stop routine
func (rm *RtManager) StopRoutine(name string) error {
	rm.rtBuffLock.Lock()
	defer func() {
		rm.rtBuffLock.Unlock()
		rm.CheckRoutines()
	}()

	if _, ok := rm.rtBuffer[name]; !ok {
		return ErrInvalidRoutine
	}
	rm.Log.Trace1("deactivating routine: %s", name)
	rm.rtBuffer[name].active = false
	return nil
}

// restart routine
func (rm *RtManager) RestartRoutine(name string) error {
	rm.rtBuffLock.Lock()
	defer func() {
		rm.rtBuffLock.Unlock()
		rm.CheckRoutines()
	}()

	if _, ok := rm.rtBuffer[name]; !ok {
		return ErrInvalidRoutine
	}
	rm.Log.Trace1("restarting routine: %s", name)
	rm.rtBuffer[name].active = true
	if rm.rtBuffer[name].rt.IsAlive() {
		rm.rtBuffer[name].rt.Stop()
	}
	return nil
}

// monitor routines and start/stop as per each routine status
func (rm *RtManager) CheckRoutines() error {
	rm.rtBuffLock.Lock()
	defer rm.rtBuffLock.Unlock()

	rm.Log.Trace1("checking routines ...")
	for n := range rm.rtBuffer {
		if rm.rtBuffer[n].active {
			if !rm.rtBuffer[n].rt.IsAlive() {
				rm.Log.Info("starting routine: %s", n)
				go rm.rtBuffer[n].rt.Start()
			}
		} else {
			if rm.rtBuffer[n].rt.IsAlive() {
				rm.Log.Info("stopping routine: %s", n)
				rm.rtBuffer[n].rt.Stop()
			}
		}
	}
	return nil
}
