// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"fmt"
	"strings"
	"sync"

	"github.com/exonlabs/go-utils/pkg/logging"
)

// Routine defines the methods that must be implemented by any routine
// managed by routine manager.
type Routine interface {
	IsAlive() bool
	Start()
	Stop()
	Kill()
}

type RoutineHandler = TaskletHandler

func NewRoutineHandler(log *logging.Logger, tsk Tasklet) *RoutineHandler {
	return NewTaskletHandler(log, tsk)
}

// rtInfo holds the information about a routine's activation state.
type rtInfo struct {
	rt        Routine
	activated bool
}

// RoutineManager manages the lifecycle of routines, allowing them to be
// started, stopped, and monitored.
type RoutineManager struct {
	*Process

	// rtBuffer holds the mapping of routine names to their information.
	rtBuffer map[string]*rtInfo

	// rtBuffLock is used to synchronize access to rtBuffer.
	rtBuffLock sync.Mutex

	// MonitorInterval defines the time interval to check the status of routines.
	MonitorInterval float64

	// StoppingDelay specifies the duration to wait for routines to stop
	// before considering them terminated.
	StoppingDelay float64
}

// New creates a new routine manager instance.
func NewRoutineManager(log *logging.Logger) *RoutineManager {
	rm := &RoutineManager{
		rtBuffer:        make(map[string]*rtInfo),
		MonitorInterval: 2,
		StoppingDelay:   3,
	}
	rm.Process = NewProcessHandler(log, rm)
	return rm
}

// Initialize prepares the routine manager.
func (m *RoutineManager) Initialize() error {
	if len(m.rtBuffer) == 0 {
		return fmt.Errorf("no routines loaded")
	}
	m.Log.Debug("loaded routines: %s", strings.Join(m.ListRoutines(), ", "))
	return nil
}

// Execute runs the routine check and waits for the specified monitor interval.
func (m *RoutineManager) Execute() error {
	m.CheckRoutines()
	m.Sleep(m.MonitorInterval)
	return nil
}

// Terminate stops all activated routines and waits for them to finish.
func (m *RoutineManager) Terminate() error {
	defer func() {
		if !m.rtBuffLock.TryLock() {
			m.rtBuffLock.Unlock()
		}
	}()

	m.Log.Info("stopping all activated routines")
	m.rtBuffLock.Lock()
	for n := range m.rtBuffer {
		m.rtBuffer[n].activated = false
		if m.rtBuffer[n].rt.IsAlive() {
			m.Log.Info("stopping routine: %s", n)
			m.rtBuffer[n].rt.Stop()
		}
	}
	m.rtBuffLock.Unlock()

	// no wait required
	if m.StoppingDelay <= 0 {
		return nil
	}

	// check and wait all routines exit
	tPoll := float64(0.1)
	for t := m.StoppingDelay; t > 0 && !m.KillEvent.IsSet(); t -= tPoll {
		m.Sleep(tPoll)
		chk := true
		m.rtBuffLock.Lock()
		for n := range m.rtBuffer {
			if m.rtBuffer[n].rt.IsAlive() {
				chk = false
				break
			}
		}
		m.rtBuffLock.Unlock()
		if chk {
			return nil
		}
	}

	names := []string{}
	m.rtBuffLock.Lock()
	for n := range m.rtBuffer {
		if m.rtBuffer[n].rt.IsAlive() {
			names = append(names, n)
		}
	}
	m.rtBuffLock.Unlock()
	m.Log.Error("failed stopping routines: %s", strings.Join(names, ", "))
	return nil
}

// ListRoutines returns a slice of names of all routines managed by routine manager.
func (m *RoutineManager) ListRoutines() []string {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	names := []string{}
	for n := range m.rtBuffer {
		names = append(names, n)
	}
	return names
}

// AddRoutine adds a new routine to the routine manager.
func (m *RoutineManager) AddRoutine(name string, rt Routine, activated bool) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; ok {
		return fmt.Errorf("duplicate routine name")
	}

	m.rtBuffer[name] = &rtInfo{rt, activated}
	m.Log.Trace1("added routine: %s", name)
	return nil
}

// DelRoutine removes a routine from the routine manager.
func (m *RoutineManager) DelRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.rtBuffer[name].activated = false
	m.rtBuffer[name].rt.Kill()
	m.Sleep(0.2)
	if m.rtBuffer[name].rt.IsAlive() {
		m.Sleep(1)
		if m.rtBuffer[name].rt.IsAlive() {
			return fmt.Errorf("failed to stop routine: %s", name)
		}
	}

	m.Log.Trace1("deleting routine: %s", name)
	delete(m.rtBuffer, name)
	return nil
}

// StartRoutine activates a routine, allowing it to run.
func (m *RoutineManager) StartRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer func() {
		m.rtBuffLock.Unlock()
		m.CheckRoutines()
	}()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.Log.Trace1("activating routine: %s", name)
	m.rtBuffer[name].activated = true
	return nil
}

// StopRoutine deactivates a routine, preventing it from running.
func (m *RoutineManager) StopRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer func() {
		m.rtBuffLock.Unlock()
		m.CheckRoutines()
	}()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.Log.Trace1("deactivating routine: %s", name)
	m.rtBuffer[name].activated = false
	return nil
}

// RestartRoutine restarts a routine, stopping it if it's currently running.
func (m *RoutineManager) RestartRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer func() {
		m.rtBuffLock.Unlock()
		m.CheckRoutines()
	}()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.Log.Trace1("restarting routine: %s", name)
	m.rtBuffer[name].activated = true
	if m.rtBuffer[name].rt.IsAlive() {
		m.rtBuffer[name].rt.Stop()
	}
	return nil
}

// CheckRoutines checks the status of each routine and starts or stops them as needed.
func (m *RoutineManager) CheckRoutines() error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	m.Log.Trace1("checking routines ...")
	for n := range m.rtBuffer {
		if m.rtBuffer[n].activated {
			if !m.rtBuffer[n].rt.IsAlive() {
				m.Log.Info("starting routine: %s", n)
				go m.rtBuffer[n].rt.Start()
			}
		} else {
			if m.rtBuffer[n].rt.IsAlive() {
				m.Log.Info("stopping routine: %s", n)
				m.rtBuffer[n].rt.Stop()
			}
		}
	}
	return nil
}
