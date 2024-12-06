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
	IsEnabled() bool
	IsAlive() bool
	IsInitialized() bool
	Enable()
	Disable()
	Start()
	Stop()
	Kill()
}

type RoutineHandler = TaskletHandler

func NewRoutineHandler(log *logging.Logger, tsk Tasklet) *RoutineHandler {
	return NewTaskletHandler(log, tsk)
}

// RoutineManager manages the lifecycle of routines, allowing them to be
// started, stopped, and monitored.
type RoutineManager struct {
	*Process

	// rtBuffer holds the mapping of routine names to their information.
	rtBuffer map[string]Routine
	// rtBuffLock is used to synchronize access to rtBuffer.
	rtBuffLock sync.Mutex

	// MonitoringInterval specifies the routines monitoring interval in sec.
	MonitoringInterval float64
	// StoppingDelay specifies the duration to wait for routines to stop.
	StoppingDelay float64
}

// New creates a new routine manager instance.
func NewRoutineManager(log *logging.Logger) *RoutineManager {
	rm := &RoutineManager{
		rtBuffer:           make(map[string]Routine),
		MonitoringInterval: 300,
		StoppingDelay:      3,
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
	for n := range m.rtBuffer {
		if m.rtBuffer[n].IsEnabled() && !m.rtBuffer[n].IsAlive() {
			go m.rtBuffer[n].Start()
		}
	}
	m.Sleep(m.MonitoringInterval)
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
		m.rtBuffer[n].Disable()
		if m.rtBuffer[n].IsAlive() {
			m.Log.Info("stopping routine: %s", n)
			m.rtBuffer[n].Stop()
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
			if m.rtBuffer[n].IsAlive() {
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
		if m.rtBuffer[n].IsAlive() {
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
func (m *RoutineManager) AddRoutine(name string, rt Routine, enabled bool) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; ok {
		return fmt.Errorf("duplicate routine name")
	}

	m.rtBuffer[name] = rt
	if enabled {
		rt.Enable()
	}
	m.Log.Trace1("added routine: %s", name)

	if m.IsInitialized() {
		go rt.Start()
	}
	return nil
}

// DelRoutine removes a routine from the routine manager.
func (m *RoutineManager) DelRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.rtBuffer[name].Disable()
	if m.rtBuffer[name].IsAlive() {
		m.rtBuffer[name].Stop()
		m.Sleep(1)
		m.rtBuffer[name].Kill()
		m.Sleep(1)
		if m.rtBuffer[name].IsAlive() {
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
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.rtBuffer[name].Enable()
	if !m.rtBuffer[name].IsAlive() {
		m.Log.Trace1("activating routine: %s", name)
		go m.rtBuffer[name].Start()
	} else {
		m.Log.Trace1("already running routine: %s", name)
	}
	return nil
}

// StopRoutine deactivates a routine, preventing it from running.
func (m *RoutineManager) StopRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.Log.Trace1("deactivating routine: %s", name)
	m.rtBuffer[name].Disable()
	m.rtBuffer[name].Stop()
	return nil
}

// RestartRoutine restarts a routine, stopping it if it's currently running.
func (m *RoutineManager) RestartRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return fmt.Errorf("invalid routine name")
	}

	m.rtBuffer[name].Enable()
	if m.rtBuffer[name].IsAlive() {
		m.Log.Trace1("restarting routine: %s", name)
		m.rtBuffer[name].Stop()
	} else {
		m.Log.Trace1("starting routine: %s", name)
		go m.rtBuffer[name].Start()
	}
	return nil
}
