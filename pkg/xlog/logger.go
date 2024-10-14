// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xlog

import (
	"errors"
)

// Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level int

// Names for common log levels.
const (
	TRACE3 Level = -4
	TRACE2 Level = -3
	TRACE1 Level = -2
	DEBUG  Level = -1
	INFO   Level = 0
	WARN   Level = 1
	ERROR  Level = 2
	FATAL  Level = 3
	PANIC  Level = 4
)

// String returns a name for the level in uppercase.
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
	}
	return "PANIC"
}

// A Logger records structured information about each call to its methods.
// For each call, it creates a new log message formatted with [Formatter]
// and passes it to the logger handlers and to its parent logger.
type Logger struct {
	Name      string // logger name
	Level     Level  // logger level
	parent    *Logger
	formatter *Formatter
	handlers  []Handler
}

// NewLogger creates new named logger instance. new logger has default
// level=INFO and has the standard message [Formatter].
// no handlers are defined by default.
func NewLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: NewStdFormatter(),
	}
}

// NewStdoutLogger creates new logger instance same as [NewLogger]
// but with Stdout handler defined.
func NewStdoutLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: NewStdFormatter(),
		handlers:  []Handler{NewStdoutHandler()},
	}
}

// NewFileLogger creates new logger instance same as [NewLogger]
// but with file handler defined.
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

// SubLogger creates new named child logger same as [ChildLogger]
// but also modifies the inherited [Formatter] adding prefix to messages.
func (l *Logger) SubLogger(prefix string) *Logger {
	return &Logger{
		Name:   l.Name,
		parent: l,
		Level:  l.Level,
		formatter: &Formatter{
			MsgPrefix:    prefix,
			RecordFormat: l.formatter.RecordFormat,
			TimeFormat:   l.formatter.TimeFormat,
			EscapeMsg:    l.formatter.EscapeMsg,
		},
	}
}

// SetFormatter set a new [Formatter] f on the logger instance.
func (l *Logger) SetFormatter(f *Formatter) {
	if f != nil {
		l.formatter = f
	}
}

// AddHandler adds a new [Handler] h on the logger instance.
func (l *Logger) AddHandler(h Handler) {
	if h != nil {
		if l.handlers == nil {
			l.handlers = []Handler{h}
		} else {
			l.handlers = append(l.handlers, h)
		}
	}
}

// ClearHandlers removes all attached handlers on the logger instance.
func (l *Logger) ClearHandlers() {
	l.handlers = nil
}

// process log message record r, it sends the log record to all logger
// handlers, then forwards the record to parent logger.
func (l *Logger) log(r string) error {
	var errAll error
	// handle record with loaded handlers
	for _, h := range l.handlers {
		if err := h.HandleRecord(r); err != nil {
			errAll = errors.Join(errAll, err)
		}
	}
	// propagate to parent logger
	if l.parent != nil {
		if err := l.parent.log(r); err != nil {
			errAll = errors.Join(errAll, err)
		}
	}
	return errAll
}

// logs a new message with Panic severity level.
func (l *Logger) Panic(msg string, args ...any) error {
	if l.Level <= PANIC {
		return l.log(l.formatter.Emit(PANIC, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Fatal severity level.
func (l *Logger) Fatal(msg string, args ...any) error {
	if l.Level <= FATAL {
		return l.log(l.formatter.Emit(FATAL, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Error severity level.
func (l *Logger) Error(msg string, args ...any) error {
	if l.Level <= ERROR {
		return l.log(l.formatter.Emit(ERROR, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Warn severity level.
func (l *Logger) Warn(msg string, args ...any) error {
	if l.Level <= WARN {
		return l.log(l.formatter.Emit(WARN, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Info level.
func (l *Logger) Info(msg string, args ...any) error {
	if l.Level <= INFO {
		return l.log(l.formatter.Emit(INFO, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Debug level.
func (l *Logger) Debug(msg string, args ...any) error {
	if l.Level <= DEBUG {
		return l.log(l.formatter.Emit(DEBUG, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Trace1  level.
func (l *Logger) Trace1(msg string, args ...any) error {
	if l.Level <= TRACE1 {
		return l.log(l.formatter.Emit(TRACE1, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Trace2 level.
func (l *Logger) Trace2(msg string, args ...any) error {
	if l.Level <= TRACE2 {
		return l.log(l.formatter.Emit(TRACE2, l.Name, msg, args...))
	}
	return nil
}

// logs a new message with Trace3 level.
func (l *Logger) Trace3(msg string, args ...any) error {
	if l.Level <= TRACE3 {
		return l.log(l.formatter.Emit(TRACE3, l.Name, msg, args...))
	}
	return nil
}
