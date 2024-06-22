package xlog

import (
	"fmt"
	"strings"
	"time"
)

type Formatter struct {
	RecordFormat string
	TimeFormat   string
	EscapeMsg    bool
}

// generate new formatted log record
func (f *Formatter) Emit(
	lvl Level, src string, msg string, args ...any) string {
	now := time.Now().Local()

	var t string
	if len(f.TimeFormat) == 0 {
		t = now.Format("2006-01-02 15:04:05.000000")
	} else {
		t = now.Format(f.TimeFormat)
	}

	m := fmt.Sprintf(msg, args...)
	if f.EscapeMsg {
		m = strings.ReplaceAll(m, `\`, `\\`)
		m = strings.ReplaceAll(m, `"`, `\"`)
	}

	return strings.NewReplacer(
		"{time}", t,
		"{level}", lvl.String(),
		"{source}", src,
		"{message}", m,
	).Replace(f.RecordFormat)
}

// standard text formatted log record
func StdFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} [{source}] {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// simple text formatted log record, without source
func SimpleFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// basic text formatted log record, just timestamp and message
func BasicFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// raw text formatted log record, just the message
func RawFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{message}",
		TimeFormat:   "",
	}
}

// customized message text formatter
func CustomMsgFormatter(recFmt string) *Formatter {
	return &Formatter{
		RecordFormat: recFmt,
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// customized timestamp text formatter
func CustomTimeFormatter(tsFmt string) *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   tsFmt,
	}
}

// standard json formatted log record
func JsonFormatter() *Formatter {
	return &Formatter{
		RecordFormat: `{"ts":"{time}","lvl":"{level}",` +
			`"src":"{source}","msg":"{message}"}`,
		TimeFormat: "2006-01-02 15:04:05.000000",
		EscapeMsg:  true,
	}
}
