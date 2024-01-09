package xlog

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Formatter interface {
	ParseRecord(Record) string
}

var (
	defaultFormatter atomic.Value
)

func init() {
	SetDefaultFormatter(NewStdFormatter(
		"{time} {level} {message}", "2006-01-02 15:04:05.000000"))
}

func GetDefaultFormatter() Formatter {
	return defaultFormatter.Load().(Formatter)
}

func SetDefaultFormatter(f Formatter) {
	defaultFormatter.Store(f)
}

type StdFormatter struct {
	RecordFormat string
	TimeFormat   string
}

func NewStdFormatter(recFmt, tsFmt string) StdFormatter {
	return StdFormatter{recFmt, tsFmt}
}

func (f StdFormatter) ParseRecord(r Record) string {
	return strings.NewReplacer(
		"{time}", r.Time.Format(f.TimeFormat),
		"{level}", StringLevel(r.Level),
		"{source}", r.Source,
		"{message}", fmt.Sprintf(r.Message, r.MsgArgs...),
	).Replace(f.RecordFormat)
}
