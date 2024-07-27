package xlog

import (
	"errors"
)

type Level int

// logging levels
const (
	TRACE4 Level = -50
	TRACE3 Level = -40
	TRACE2 Level = -30
	TRACE1 Level = -20
	DEBUG  Level = -10
	INFO   Level = 0
	WARN   Level = 10
	ERROR  Level = 20
	FATAL  Level = 30
	PANIC  Level = 40
)

// Stringer interface, returns string of log level
func (l Level) String() string {
	switch {
	case l >= PANIC:
		return "PANIC"
	case l >= FATAL:
		return "FATAL"
	case l >= ERROR:
		return "ERROR"
	case l >= WARN:
		return "WARN "
	case l >= INFO:
		return "INFO "
	case l >= DEBUG:
		return "DEBUG"
	case l >= TRACE1:
		return "TRACE1"
	case l >= TRACE2:
		return "TRACE2"
	case l >= TRACE3:
		return "TRACE3"
	case l >= TRACE4:
		return "TRACE4"
	}
	return "TRACE"
}

type Logger struct {
	Name      string
	Level     Level
	parent    *Logger
	formatter *Formatter
	handlers  []Handler
}

func NewLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: StdFormatter(),
	}
}

func (l *Logger) ChildLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		parent:    l,
		Level:     l.Level,
		formatter: l.formatter,
	}
}

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

func (l *Logger) SetFormatter(f *Formatter) {
	if f != nil {
		l.formatter = f
	}
}

func (l *Logger) AddHandler(h Handler) {
	if h != nil {
		if l.handlers == nil {
			l.handlers = []Handler{h}
		} else {
			l.handlers = append(l.handlers, h)
		}
	}
}

func (l *Logger) ClearHandlers() {
	l.handlers = nil
}

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

func (l *Logger) Panic(msg string, args ...any) error {
	if l.Level <= PANIC {
		return l.log(l.formatter.Emit(PANIC, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Fatal(msg string, args ...any) error {
	if l.Level <= FATAL {
		return l.log(l.formatter.Emit(FATAL, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Error(msg string, args ...any) error {
	if l.Level <= ERROR {
		return l.log(l.formatter.Emit(ERROR, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Warn(msg string, args ...any) error {
	if l.Level <= WARN {
		return l.log(l.formatter.Emit(WARN, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Info(msg string, args ...any) error {
	if l.Level <= INFO {
		return l.log(l.formatter.Emit(INFO, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Debug(msg string, args ...any) error {
	if l.Level <= DEBUG {
		return l.log(l.formatter.Emit(DEBUG, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Trace1(msg string, args ...any) error {
	if l.Level <= TRACE1 {
		return l.log(l.formatter.Emit(TRACE1, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Trace2(msg string, args ...any) error {
	if l.Level <= TRACE2 {
		return l.log(l.formatter.Emit(TRACE2, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Trace3(msg string, args ...any) error {
	if l.Level <= TRACE3 {
		return l.log(l.formatter.Emit(TRACE3, l.Name, msg, args...))
	}
	return nil
}
func (l *Logger) Trace4(msg string, args ...any) error {
	if l.Level <= TRACE4 {
		return l.log(l.formatter.Emit(TRACE4, l.Name, msg, args...))
	}
	return nil
}

// ///////////////////// creator functions

func NewStdoutLogger(name string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: StdFormatter(),
		handlers:  []Handler{NewStdoutHandler()},
	}
}

func NewFileLogger(name, path string) *Logger {
	return &Logger{
		Name:      name,
		Level:     INFO,
		formatter: StdFormatter(),
		handlers:  []Handler{NewFileHandler(path)},
	}
}
