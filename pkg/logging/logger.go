// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"errors"
	"fmt"
	"time"
)

const (
	// Predefined log levels.
	TRACE = int(-2) // Trace level
	DEBUG = int(-1) // Debug level
	INFO  = int(0)  // Info level
	WARN  = int(1)  // Warning level
	ERROR = int(2)  // Error level
	FATAL = int(3)  // Fatal level
	PANIC = int(4)  // Panic level
)

// LEVEL returns the string representation of the log level.
func LEVEL(lvl int) string {
	switch {
	case lvl >= PANIC:
		return "PANIC"
	case lvl >= FATAL:
		return "FATAL"
	case lvl >= ERROR:
		return "ERROR"
	case lvl >= WARN:
		return "WARN"
	case lvl >= INFO:
		return "INFO"
	case lvl >= DEBUG:
		return "DEBUG"
	default:
		return "TRACE"
	}
}

// A Logger records structured information about each call to its methods.
// For each call, it creates a new log record and passes it to the logger
// handlers and to its parent logger.
type Logger struct {
	name      string    // Logger name
	Level     int       // Logger level
	Prefix    string    // an optional prefix for all logger records
	Suffix    string    // an optional suffix for all logger records
	formatter Formatter // Formatter for generating log messages
	handlers  []Handler // Handlers for processing log records
	parent    *Logger   // Parent logger for inheritance
}

// NewStdoutLogger creates a new logger that outputs to standard output.
func NewStdoutLogger(name string) *Logger {
	return &Logger{
		name:      name,
		Level:     INFO,
		formatter: StdFormatter,
		handlers:  []Handler{NewStdoutHandler()},
	}
}

// NewFileLogger creates a new logger that logs to a specified file.
func NewFileLogger(name, path string) *Logger {
	return &Logger{
		name:      name,
		Level:     INFO,
		formatter: StdFormatter,
		handlers:  []Handler{NewFileHandler(path)},
	}
}

// ChildLogger creates new named child logger.
// child logger inherits the parent logger [Formatter].
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO [child_name] log message
func (l *Logger) ChildLogger(name string) *Logger {
	return &Logger{
		parent:    l,
		name:      name,
		Level:     TRACE,
		Prefix:    l.Prefix,
		Suffix:    l.Suffix,
		formatter: l.formatter,
	}
}

// SubLogger creates a new child logger with name added between brackets before prefix.
// child sub logger inherits the parent logger [Formatter].
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO [parent_name] (child_name) log message
func (l *Logger) SubLogger(name string) *Logger {
	return &Logger{
		parent:    l,
		name:      l.name,
		Level:     TRACE,
		Prefix:    fmt.Sprintf("(%s) %s", name, l.Prefix),
		Suffix:    l.Suffix,
		formatter: l.formatter,
	}
}

// Name returns the logger name.
func (l *Logger) Name() string {
	return l.name
}

// SetFormatter sets a new formatter for the logger.
func (l *Logger) SetFormatter(f Formatter) {
	if f != nil {
		l.formatter = f
	}
}

// SetHandler clears all handler and set new one to the logger.
func (l *Logger) SetHandler(h Handler) {
	if h != nil {
		l.handlers = []Handler{h}
	}
}

// AddHandler adds a new handler to the logger.
func (l *Logger) AddHandler(h Handler) {
	if h != nil {
		l.handlers = append(l.handlers, h)
	}
}

// ClearHandlers removes all handlers from the logger.
func (l *Logger) ClearHandlers() {
	l.handlers = nil
}

// Log handles a log message, sending it to all handlers and parents.
func (l *Logger) Log(level int, msg string) error {
	var errAll error

	// process record by local handlers
	if level >= l.Level && l.handlers != nil {
		for _, h := range l.handlers {
			if err := h.HandleMessage(msg); err != nil {
				// Combine errors
				errAll = errors.Join(errAll, err)
			}
		}
	}

	// Propagate to parent logger
	if l.parent != nil {
		if err := l.parent.Log(level, msg); err != nil {
			errAll = errors.Join(errAll, err)
		}
	}

	return errAll
}

// Panic logs a record with Panic level.
func (l *Logger) Panic(msg string, args ...any) error {
	return l.Log(PANIC, l.formatter(
		time.Now().Local(), PANIC, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Fatal logs a record with Fatal level.
func (l *Logger) Fatal(msg string, args ...any) error {
	return l.Log(FATAL, l.formatter(
		time.Now().Local(), FATAL, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Error logs a record with Error level.
func (l *Logger) Error(msg string, args ...any) error {
	return l.Log(ERROR, l.formatter(
		time.Now().Local(), ERROR, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Warn logs a record with Warn level.
func (l *Logger) Warn(msg string, args ...any) error {
	return l.Log(WARN, l.formatter(
		time.Now().Local(), WARN, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Info logs a record with Info level.
func (l *Logger) Info(msg string, args ...any) error {
	return l.Log(INFO, l.formatter(
		time.Now().Local(), INFO, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Debug logs a record with Debug level.
func (l *Logger) Debug(msg string, args ...any) error {
	return l.Log(DEBUG, l.formatter(
		time.Now().Local(), DEBUG, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}

// Trace logs a record with Trace level.
func (l *Logger) Trace(msg string, args ...any) error {
	return l.Log(TRACE, l.formatter(
		time.Now().Local(), TRACE, l.name,
		fmt.Sprintf(l.Prefix+msg+l.Suffix, args...),
	))
}
