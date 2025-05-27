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
	"github.com/exonlabs/go-utils/pkg/tasklet"
)

var (
	ErrInvalidTaskletName = errors.New("invalid tasklet name")
	ErrNoLoadedTaslets    = errors.New("no loaded tasklets")
)

// TaskletManager runs and manages the tasklet lifecycle inside a routine.
type TaskletManager struct {
	*Process

	// tBuff holds the mapping of tasklet names to their instances.
	tBuff map[string]tasklet.Tasklet
	// tBuffMutex is used to synchronize access to tBuff.
	tBuffMutex sync.Mutex

	// MonitoringInterval specifies the tasklets monitoring interval in sec.
	// default 30 seconds.
	MonitoringInterval float64
	// StoppingDelay specifies the duration to wait for tasklets to stop.
	// default 5 seconds.
	StoppingDelay float64
}

// NewTaskletManager constructs a TaskletManager with sane defaults and an
// embedded Process for signal handling.
func NewTaskletManager(log *logging.Logger) *TaskletManager {
	m := &TaskletManager{
		tBuff:              make(map[string]tasklet.Tasklet),
		MonitoringInterval: 30,
		StoppingDelay:      5,
	}
	m.Process = NewProcess(log, m)
	return m
}

// start a tasklet with error logging
func (m *TaskletManager) start_tasklet(name string, tsk tasklet.Tasklet) {
	if err := tsk.Start(); err != nil {
		if m.Log != nil {
			m.Log.Error("start tasklet: %s failed - %s", name, err)
		} else {
			fmt.Printf("start tasklet: %s failed - %s\n", name, err)
		}
	}
}

// stop a tasklet with error logging
func (m *TaskletManager) stop_tasklet(name string, tsk tasklet.Tasklet) {
	if err := tsk.Stop(); err != nil {
		if m.Log != nil {
			m.Log.Error("stop tasklet: %s failed - %s", name, err)
		} else {
			fmt.Printf("stop tasklet: %s failed - %s\n", name, err)
		}
	}
}

// ListTasklets returns a sorted slice of all registered tasklet names.
func (m *TaskletManager) ListTasklets() []string {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	names := []string{}
	for name := range m.tBuff {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CheckTasklets reconciles each tasklet’s enabled flag with its liveness,
// starting or stopping tasklets as necessary.
func (m *TaskletManager) CheckTasklets() error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	for name := range m.tBuff {
		if m.tBuff[name].IsEnabled() && !m.tBuff[name].IsActive() {
			if m.Log != nil {
				m.Log.Warn("starting dead-enabled tasklet: %s", name)
			}
			go m.start_tasklet(name, m.tBuff[name])
		} else if !m.tBuff[name].IsEnabled() && m.tBuff[name].IsActive() {
			if m.Log != nil {
				m.Log.Warn("stopping alive-disabled tasklet: %s", name)
			}
			go m.stop_tasklet(name, m.tBuff[name])
		}
	}
	return nil
}

// AddTasklet registers a new tasklet by name.
// It returns an error if the name is empty.
func (m *TaskletManager) AddTasklet(name string, rt tasklet.Tasklet) error {
	if len(name) == 0 {
		return ErrInvalidTaskletName
	}

	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	m.tBuff[name] = rt
	return nil
}

// DeleteTasklet removes a tasklet by nmae. It returns
// error if the name is invalid or the tasklet is running.
func (m *TaskletManager) DeleteTasklet(name string) error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	if _, ok := m.tBuff[name]; !ok {
		return ErrInvalidTaskletName
	}

	if m.tBuff[name].IsActive() {
		return tasklet.ErrIsRunning
	}
	delete(m.tBuff, name)
	return nil
}

// IsTaskletAlive reports whether a tasklet by name exists and is currently
// running or not.
func (m *TaskletManager) IsTaskletAlive(name string) bool {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	if _, ok := m.tBuff[name]; !ok {
		return false
	}
	return m.tBuff[name].IsActive()
}

// StartTasklet enables and starts a tasklet by name.
func (m *TaskletManager) StartTasklet(name string) error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	if _, ok := m.tBuff[name]; !ok {
		return ErrInvalidTaskletName
	}

	m.tBuff[name].Enable()
	go m.start_tasklet(name, m.tBuff[name])
	return nil
}

// StopTasklet disables and stops a tasklet by name.
func (m *TaskletManager) StopTasklet(name string) error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	if _, ok := m.tBuff[name]; !ok {
		return ErrInvalidTaskletName
	}

	m.tBuff[name].Disable()
	go m.stop_tasklet(name, m.tBuff[name])
	return nil
}

// RestartTasklet (re)enables a tasklet by name, then if tasklet is not
// alive it is started and if running it is restarted.
func (m *TaskletManager) RestartTasklet(name string) error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	if _, ok := m.tBuff[name]; !ok {
		return ErrInvalidTaskletName
	}

	m.tBuff[name].Enable()
	if m.tBuff[name].IsActive() {
		go m.stop_tasklet(name, m.tBuff[name])
	} else {
		go m.start_tasklet(name, m.tBuff[name])
	}
	return nil
}

// Initialize verifies that at least one tasklet has been loaded and logs
// their names. It should be called once before Start().
func (m *TaskletManager) Initialize() error {
	names := m.ListTasklets()
	if len(names) == 0 {
		return ErrNoLoadedTaslets
	}

	if m.Log != nil {
		m.Log.Info("loaded tasklets: %s", strings.Join(names, ", "))
	}

	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	for name := range m.tBuff {
		if m.tBuff[name].IsEnabled() {
			if m.Log != nil {
				m.Log.Info("starting tasklet: %s", name)
			}
			go m.start_tasklet(name, m.tBuff[name])
		}
	}
	return nil
}

// Execute performs one monitoring cycle: it reconciles desired tasklet state
// (enabled/disabled) with actual liveness, then sleeps for the configured
// MonitoringInterval.
func (m *TaskletManager) Execute() error {
	m.Sleep(m.MonitoringInterval)
	return m.CheckTasklets()
}

// Terminate disables every tasklet, waits up to StoppingDelay seconds for each
// to shut down, and logs any that fail to stop in time.
func (m *TaskletManager) Terminate() error {
	m.tBuffMutex.Lock()
	defer m.tBuffMutex.Unlock()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var nonStopped []string

	if m.Log != nil {
		m.Log.Info("stopping all activate tasklets")
	}
	for name := range m.tBuff {
		m.tBuff[name].Disable()
		if m.tBuff[name].IsActive() {
			if m.Log != nil {
				m.Log.Info("stopping tasklet: %s", name)
			}
			go m.stop_tasklet(name, m.tBuff[name])

			// if wait stop is required, run routines monitoring for tasklets
			// termination untill StoppingDelay timeout.
			if m.StoppingDelay > 0 {
				wg.Add(1)
				go func(n string) {
					defer wg.Done()
					if !m.tBuff[n].Join(m.StoppingDelay) {
						mu.Lock()
						defer mu.Unlock()
						nonStopped = append(nonStopped, n)
					}
				}(name)
			}
		}
	}

	// if wait stop is required, log non-stopped tasklets after StoppingDelay timeout.
	if m.StoppingDelay > 0 {
		wg.Wait()
		if m.Log != nil && len(nonStopped) > 0 {
			sort.Strings(nonStopped)
			m.Log.Warn("failed stopping tasklets: %s",
				strings.Join(nonStopped, ", "))
		}
	}

	return nil
}
