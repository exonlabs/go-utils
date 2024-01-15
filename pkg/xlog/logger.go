package xlog

import (
	"errors"
	"sync/atomic"
)

var rootLogger atomic.Value

func GetRootLogger() *Logger {
	if l := rootLogger.Load(); l != nil {
		return l.(*Logger)
	}
	SetRootLogger(NewLogger("root"))
	return rootLogger.Load().(*Logger)
}

func SetRootLogger(l *Logger) {
	rootLogger.Store(l)
}

type Logger struct {
	Name     string
	Level    Level
	Frmt     Formatter
	Handlers []Handler
	Parent   *Logger

	// default handlers state
	defHndlers bool
}

func NewLogger(name string) *Logger {
	return &Logger{
		Name:       name,
		Level:      INFO,
		Handlers:   []Handler{NewStdoutHandler()},
		defHndlers: true,
	}
}

func (l *Logger) NewChildLogger(name string) *Logger {
	return &Logger{
		Name:       name,
		Parent:     l,
		Level:      l.Level,
		Handlers:   []Handler{},
		defHndlers: false,
	}
}

func (l *Logger) SetFormatter(f Formatter) {
	l.Frmt = f
	for _, h := range l.Handlers {
		h.SetFormatter(l.Frmt)
	}
}

func (l *Logger) AddHandler(h Handler) {
	if l.Frmt != nil {
		h.SetFormatter(l.Frmt)
	}
	if l.defHndlers {
		l.Handlers = []Handler{}
		l.defHndlers = false
	}
	l.Handlers = append(l.Handlers, h)
}

func (l *Logger) ClearHandlers() {
	l.Handlers = []Handler{}
}

func (l *Logger) Log(r Record) error {
	var resErr error
	// handle record with loaded handlers
	if r.Level >= l.Level {
		for _, h := range l.Handlers {
			if err := h.HandleRecord(r); err != nil {
				resErr = errors.Join(resErr, err)
			}
		}
	}
	// propagate to parent logger
	if l.Parent != nil {
		if err := l.Parent.Log(r); err != nil {
			resErr = errors.Join(resErr, err)
		}
	}
	return resErr
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
