package xlog

import (
	"time"
)

type Record struct {
	Time    time.Time
	Level   Level
	Source  string
	Message string
	MsgArgs []any
}

// create new logging record
func NewRecord(lvl Level, src string, msg string, args ...any) Record {
	return Record{
		Time:    time.Now().Local(),
		Source:  src,
		Level:   lvl,
		Message: msg,
		MsgArgs: args,
	}
}
