// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package proc

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/exonlabs/go-utils/pkg/logging"
)

var (
	ErrRtInvalidName = errors.New("invalid routine name")
	ErrRtIsRunning   = errors.New("routine is still running")
)

// Routine represents a long‑running job managed by a RoutineManager.
type Routine interface {
	IsEnabled() bool
	IsAlive() bool
	Enable()
	Disable()
	Start() error
	Stop() error
	WaitTerm(float64) bool
}

// RoutineHandler is an alias for TaskletHandler.
type RoutineHandler = TaskletHandler

// NewRoutineHandler is an alias for NewTaskletHandler.
var NewRoutineHandler = NewTaskletHandler

// RoutineManager starts, stops, and monitors multiple routines, and integrates
// with Process to honor OS‑level shutdown signals.
type RoutineManager struct {
	*Process

	// rtBuffer holds the mapping of routine names to their information.
	rtBuffer map[string]Routine
	// rtBuffLock is used to synchronize access to rtBuffer.
	rtBuffLock sync.Mutex

	// MonitoringInterval specifies the routines monitoring interval in sec.
	// default 60 seconds.
	MonitoringInterval float64
	// StoppingDelay specifies the duration to wait for routines to stop.
	// default 5 seconds.
	StoppingDelay float64
}

// NewRoutineManager constructs a RoutineManager with sane defaults and an
// embedded Process for signal handling.
func NewRoutineManager(log *logging.Logger) *RoutineManager {
	rm := &RoutineManager{
		rtBuffer:           make(map[string]Routine),
		MonitoringInterval: 5,
		StoppingDelay:      5,
	}
	rm.Process = NewProcessHandler(log, rm)
	return rm
}

// start routine with error logging
func (m *RoutineManager) start_rt(name string, rt Routine) {
	if err := rt.Start(); err != nil {
		if m.Log != nil {
			m.Log.Error("start routine: %s failed - %s", name, err.Error())
		} else {
			fmt.Printf("start routine: %s failed - %s\n", name, err.Error())
		}
	}
}

// stop routine with error logging
func (m *RoutineManager) stop_rt(name string, rt Routine) {
	if err := rt.Stop(); err != nil {
		if m.Log != nil {
			m.Log.Error("stop routine: %s failed - %s", name, err.Error())
		} else {
			fmt.Printf("stop routine: %s failed - %s\n", name, err.Error())
		}
	}
}

// Initialize verifies that at least one routine has been loaded and logs their
// names. It should be called once before Start().
func (m *RoutineManager) Initialize() error {
	names := m.ListRoutines()
	if len(names) == 0 {
		return errors.New("no routines loaded")
	}

	if m.Log != nil {
		m.Log.Info("loaded routines: %s", strings.Join(names, ", "))
	}

	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	for name := range m.rtBuffer {
		if m.rtBuffer[name].IsEnabled() {
			if m.Log != nil {
				m.Log.Info("starting routine: %s", name)
			}
			go m.start_rt(name, m.rtBuffer[name])
		}
	}
	return nil
}

// Execute performs one monitoring cycle: it reconciles desired routine state
// (enabled/disabled) with actual liveness, then sleeps for the configured
// MonitoringInterval.
func (m *RoutineManager) Execute() error {
	m.Sleep(m.MonitoringInterval)
	return m.CheckRoutines()
}

// Terminate disables every routine, waits up to StoppingDelay seconds for each
// to shut down, and logs any that fail to stop in time.
func (m *RoutineManager) Terminate() error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var nonStopped []string

	if m.Log != nil {
		m.Log.Info("stopping all activate routines")
	}
	for name := range m.rtBuffer {
		m.rtBuffer[name].Disable()
		if m.rtBuffer[name].IsAlive() {
			if m.Log != nil {
				m.Log.Info("stopping routine: %s", name)
			}
			go m.stop_rt(name, m.rtBuffer[name])
			if m.StoppingDelay > 0 {
				wg.Add(1)
				go func(n string) {
					defer wg.Done()
					if !m.rtBuffer[n].WaitTerm(m.StoppingDelay) {
						mu.Lock()
						nonStopped = append(nonStopped, n)
						mu.Unlock()
					}
				}(name)
			}
		}
	}

	// if wait stop required
	if m.StoppingDelay > 0 {
		wg.Wait()
		if m.Log != nil && len(nonStopped) > 0 {
			sort.Strings(nonStopped)
			m.Log.Warn("failed stopping routines: %s",
				strings.Join(nonStopped, ", "))
		}
	}

	return nil
}

// ListRoutines returns a sorted slice of all registered routine names.
func (m *RoutineManager) ListRoutines() []string {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	names := []string{}
	for name := range m.rtBuffer {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CheckRoutines reconciles each routine’s enabled flag with its liveness,
// starting or stopping routines as necessary.
func (m *RoutineManager) CheckRoutines() error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	for name := range m.rtBuffer {
		if m.rtBuffer[name].IsEnabled() && !m.rtBuffer[name].IsAlive() {
			if m.Log != nil {
				m.Log.Warn("starting dead routine: %s", name)
			}
			go m.start_rt(name, m.rtBuffer[name])
		} else if !m.rtBuffer[name].IsEnabled() && m.rtBuffer[name].IsAlive() {
			if m.Log != nil {
				m.Log.Warn("stopping disabled routine: %s", name)
			}
			go m.stop_rt(name, m.rtBuffer[name])
		}
	}
	return nil
}

// AddRoutine registers a new routine.
// It returns an error if the name already exists.
func (m *RoutineManager) AddRoutine(name string, rt Routine) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	m.rtBuffer[name] = rt
	return nil
}

// DelRoutine removes a routine that is not running. It returns an error if the
// name is unknown or the routine is still alive.
func (m *RoutineManager) DelRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return ErrRtInvalidName
	}

	if m.rtBuffer[name].IsAlive() {
		return ErrRtIsRunning
	}
	delete(m.rtBuffer, name)
	return nil
}

// IsRoutineAlive reports whether a named routine exists and is currently
// running.
func (m *RoutineManager) IsRoutineAlive(name string) bool {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return false
	}
	return m.rtBuffer[name].IsAlive()
}

// StartRoutine enables and asynchronously starts the named routine.
func (m *RoutineManager) StartRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return ErrRtInvalidName
	}

	m.rtBuffer[name].Enable()
	go m.start_rt(name, m.rtBuffer[name])
	return nil
}

// StopRoutine disables and asynchronously stops the named routine.
func (m *RoutineManager) StopRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return ErrRtInvalidName
	}

	m.rtBuffer[name].Disable()
	go m.stop_rt(name, m.rtBuffer[name])
	return nil
}

// RestartRoutine (re)enables the named routine and either stops or starts it
// depending on its current liveness.
func (m *RoutineManager) RestartRoutine(name string) error {
	m.rtBuffLock.Lock()
	defer m.rtBuffLock.Unlock()

	if _, ok := m.rtBuffer[name]; !ok {
		return ErrRtInvalidName
	}

	m.rtBuffer[name].Enable()
	if m.rtBuffer[name].IsAlive() {
		go m.stop_rt(name, m.rtBuffer[name])
	} else {
		go m.start_rt(name, m.rtBuffer[name])
	}
	return nil
}
