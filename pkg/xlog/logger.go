package xlog

import (
	"errors"
)

type Logger struct {
	Name      string
	Level     Level
	parent    *Logger
	formatter *Formatter
	handlers  []Handler
}

func NewLogger(name string) *Logger {
	return &Logger{
		Name:  name,
		Level: INFO,
	}
}

func (l *Logger) NewChildLogger(name string) *Logger {
	return &Logger{
		Name:   name,
		parent: l,
		Level:  l.Level,
	}
}

func (l *Logger) SetFormatter(f *Formatter) {
	l.formatter = f
	if f != nil && l.handlers != nil {
		for _, h := range l.handlers {
			h.SetFormatter(f)
		}
	}
}

func (l *Logger) AddHandler(h Handler) {
	if l.formatter != nil {
		h.SetFormatter(l.formatter)
	}
	if l.handlers == nil {
		l.handlers = []Handler{h}
	} else {
		l.handlers = append(l.handlers, h)
	}
}

func (l *Logger) ClearHandlers() {
	l.handlers = nil
}

func (l *Logger) Log(r *Record) error {
	var retErr error
	// handle record with loaded handlers
	if l.handlers != nil && r.Level >= l.Level {
		for _, h := range l.handlers {
			if err := h.HandleRecord(r); err != nil {
				retErr = errors.Join(retErr, err)
			}
		}
	}
	// propagate to parent logger
	if l.parent != nil {
		if err := l.parent.Log(r); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}
	return retErr
}

func (l *Logger) Panic(msg string, args ...any) error {
	return l.Log(NewRecord(PANIC, l.Name, msg, args...))
}
func (l *Logger) Fatal(msg string, args ...any) error {
	return l.Log(NewRecord(FATAL, l.Name, msg, args...))
}
func (l *Logger) Error(msg string, args ...any) error {
	return l.Log(NewRecord(ERROR, l.Name, msg, args...))
}
func (l *Logger) Warn(msg string, args ...any) error {
	return l.Log(NewRecord(WARN, l.Name, msg, args...))
}
func (l *Logger) Info(msg string, args ...any) error {
	return l.Log(NewRecord(INFO, l.Name, msg, args...))
}
func (l *Logger) Debug(msg string, args ...any) error {
	return l.Log(NewRecord(DEBUG, l.Name, msg, args...))
}
func (l *Logger) Trace1(msg string, args ...any) error {
	return l.Log(NewRecord(TRACE1, l.Name, msg, args...))
}
func (l *Logger) Trace2(msg string, args ...any) error {
	return l.Log(NewRecord(TRACE2, l.Name, msg, args...))
}
func (l *Logger) Trace3(msg string, args ...any) error {
	return l.Log(NewRecord(TRACE3, l.Name, msg, args...))
}
func (l *Logger) Trace4(msg string, args ...any) error {
	return l.Log(NewRecord(TRACE4, l.Name, msg, args...))
}

// ///////////////////// creator functions

func NewStdoutLogger(name string) *Logger {
	return &Logger{
		Name:     name,
		Level:    INFO,
		handlers: []Handler{NewStdoutHandler()},
	}
}

func NewFileLogger(name, path string) *Logger {
	return &Logger{
		Name:     name,
		Level:    INFO,
		handlers: []Handler{NewFileHandler(path)},
	}
}
