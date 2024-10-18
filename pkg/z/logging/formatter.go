// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xlog

import (
	"fmt"
	"strings"
	"time"
)

// A Formatter handler formats the log record structure. it controls the
// record format fields "time", "level", "source" and "message".
// also a msg prefix can be added to each logged message.
type Formatter struct {
	MsgPrefix    string
	RecordFormat string
	TimeFormat   string
	EscapeMsg    bool
}

// Emit generates a formatted log record message
func (f *Formatter) Emit(lvl Level, src, msg string, args ...any) string {
	now := time.Now().Local()

	var t string
	if len(f.TimeFormat) == 0 {
		t = now.Format("2006-01-02 15:04:05.000000")
	} else {
		t = now.Format(f.TimeFormat)
	}

	m := fmt.Sprintf(f.MsgPrefix+msg, args...)
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

// Standard text formatted log record
func NewStdFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} [{source}] {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// Simple text formatted log record, without source
func NewSimpleFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// Basic text formatted log record, just timestamp and message
func NewBasicFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// Raw text formatted log record, just the message
func NewRawFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{message}",
		TimeFormat:   "",
	}
}

// Customized message text formatter
func NewCustomMsgFormatter(recFmt string) *Formatter {
	return &Formatter{
		RecordFormat: recFmt,
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// Custom timestamp text formatter
func NewCustomTimeFormatter(tsFmt string) *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   tsFmt,
	}
}

// Json formatted log record.
func NewJsonFormatter() *Formatter {
	return &Formatter{
		RecordFormat: `{"ts":"{time}","lvl":"{level}",` +
			`"src":"{source}","msg":"{message}"}`,
		TimeFormat: "2006-01-02 15:04:05.000000",
		EscapeMsg:  true,
	}
}
