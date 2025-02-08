// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"errors"
)

// Level defines the severity of a log event.
// Higher values indicate more severe log levels.
type Level int

// Predefined log levels.
const (
	TRACE3 Level = -4 // Trace level 3
	TRACE2 Level = -3 // Trace level 2
	TRACE1 Level = -2 // Trace level 1
	DEBUG  Level = -1 // Debug level
	INFO   Level = 0  // Info level
	WARN   Level = 1  // Warning level
	ERROR  Level = 2  // Error level
	FATAL  Level = 3  // Fatal level
	PANIC  Level = 4  // Panic level
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch {
	case l < DEBUG:
		return "TRACE"
	case l == DEBUG:
		return "DEBUG"
	case l == INFO:
		return "INFO "
	case l == WARN:
		return "WARN "
	case l == ERROR:
		return "ERROR"
	case l == FATAL:
		return "FATAL"
	default:
		return "PANIC"
	}
}

// A Logger records structured information about each call to its methods.
// For each call, it creates a new log message formatted with [Formatter]
// and passes it to the logger handlers and to its parent logger.
type Logger struct {
	Name      string     // Logger name
	Level     Level      // Logger level
	parent    *Logger    // Parent logger for inheritance
	formatter *Formatter // Formatter for log messages
	handlers  []Handler  // Handlers for processing log records
}

// NewStdoutLogger creates a new logger that outputs to standard output.
func NewStdoutLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: NewStdFormatter(),
		handlers:  []Handler{NewStdoutHandler()},
	}
}

// NewFileLogger creates a new logger that logs to a specified file.
func NewFileLogger(name, path string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: NewStdFormatter(),
		handlers:  []Handler{NewFileHandler(path)},
	}
}

// ChildLogger creates new named child logger from parent logger.
// child logger inherits the parent log [Level] and [Formatter].
func (l *Logger) ChildLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		parent:    l,
		Level:     l.Level,
		formatter: l.formatter,
	}
}

// SubLogger creates a new child logger with an added prefix in its messages.
func (l *Logger) SubLogger(prefix string) *Logger {
	return &Logger{
		Name:   l.Name,
		parent: l,
		Level:  l.Level,
		formatter: &Formatter{ // Inherits and modifies the formatter
			MsgPrefix:    prefix,
			RecordFormat: l.formatter.RecordFormat,
			TimeFormat:   l.formatter.TimeFormat,
			EscapeMsg:    l.formatter.EscapeMsg,
		},
	}
}

// SetFormatter sets a new formatter for the logger.
func (l *Logger) SetFormatter(f *Formatter) {
	if f != nil {
		l.formatter = f
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

// log processes the log message and sends it to all attached handlers.
func (l *Logger) log(r string) error {
	var errAll error
	for _, h := range l.handlers {
		if err := h.HandleRecord(r); err != nil {
			// Combine errors
			errAll = errors.Join(errAll, err)
		}
	}
	// Propagate to parent logger
	if l.parent != nil {
		if err := l.parent.log(r); err != nil {
			errAll = errors.Join(errAll, err)
		}
	}
	return errAll
}

// Panic logs a message with Panic severity level.
func (l *Logger) Panic(msg string, args ...any) error {
	if l.Level <= PANIC {
		return l.log(l.formatter.Emit(PANIC, l.Name, msg, args...))
	}
	return nil
}

// Fatal logs a message with Fatal severity level.
func (l *Logger) Fatal(msg string, args ...any) error {
	if l.Level <= FATAL {
		return l.log(l.formatter.Emit(FATAL, l.Name, msg, args...))
	}
	return nil
}

// Error logs a message with Error severity level.
func (l *Logger) Error(msg string, args ...any) error {
	if l.Level <= ERROR {
		return l.log(l.formatter.Emit(ERROR, l.Name, msg, args...))
	}
	return nil
}

// Warn logs a message with Warn severity level.
func (l *Logger) Warn(msg string, args ...any) error {
	if l.Level <= WARN {
		return l.log(l.formatter.Emit(WARN, l.Name, msg, args...))
	}
	return nil
}

// Info logs a message with Info severity level.
func (l *Logger) Info(msg string, args ...any) error {
	if l.Level <= INFO {
		return l.log(l.formatter.Emit(INFO, l.Name, msg, args...))
	}
	return nil
}

// Debug logs a message with Debug severity level.
func (l *Logger) Debug(msg string, args ...any) error {
	if l.Level <= DEBUG {
		return l.log(l.formatter.Emit(DEBUG, l.Name, msg, args...))
	}
	return nil
}

// Trace1 logs a message with Trace1 severity level.
func (l *Logger) Trace1(msg string, args ...any) error {
	if l.Level <= TRACE1 {
		return l.log(l.formatter.Emit(TRACE1, l.Name, msg, args...))
	}
	return nil
}

// Trace2 logs a message with Trace2 severity level.
func (l *Logger) Trace2(msg string, args ...any) error {
	if l.Level <= TRACE2 {
		return l.log(l.formatter.Emit(TRACE2, l.Name, msg, args...))
	}
	return nil
}

// Trace3 logs a message with Trace3 severity level.
func (l *Logger) Trace3(msg string, args ...any) error {
	if l.Level <= TRACE3 {
		return l.log(l.formatter.Emit(TRACE3, l.Name, msg, args...))
	}
	return nil
}
