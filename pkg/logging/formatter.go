// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"strings"
	"time"
)

// Formatter formats the log record structure. It controls the
// record format fields "time", "level", "source", and "message".
// A message prefix can also be added to each logged message.
type Formatter struct {
	MsgPrefix    string // Prefix to prepend to the message
	RecordFormat string // Template for the log record format
	TimeFormat   string // Custom time format
	EscapeMsg    bool   // Flag to escape special characters in messages
}

// Emit generates a formatted log record message.
func (f *Formatter) Emit(lvl Level, src, msg string, args ...any) string {
	now := time.Now().Local()

	// Determine the time string based on the specified format
	var t string
	if len(f.TimeFormat) == 0 {
		t = now.Format("2006-01-02 15:04:05.000000")
	} else {
		t = now.Format(f.TimeFormat)
	}

	// Format the message with optional prefix and arguments
	m := fmt.Sprintf(f.MsgPrefix+msg, args...)
	if f.EscapeMsg {
		m = strings.ReplaceAll(m, `\`, `\\`)
		m = strings.ReplaceAll(m, `"`, `\"`)
	}

	// Replace placeholders in the record format with actual values
	return strings.NewReplacer(
		"{time}", t,
		"{level}", lvl.String(),
		"{source}", src,
		"{message}", m,
	).Replace(f.RecordFormat)
}

// NewStdFormatter creates a standard text formatted log record.
func NewStdFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} [{source}] {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// NewSimpleFormatter creates a simple text formatted log record, without source.
func NewSimpleFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// NewBasicFormatter creates a basic text formatted log record, just timestamp and message.
func NewBasicFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{time} {message}",
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// NewRawFormatter creates a raw text formatted log record, just the message.
func NewRawFormatter() *Formatter {
	return &Formatter{
		RecordFormat: "{message}",
		TimeFormat:   "",
	}
}

// NewCustomMsgFormatter creates a customized message text formatter.
func NewCustomMsgFormatter(recFmt string) *Formatter {
	return &Formatter{
		RecordFormat: recFmt,
		TimeFormat:   "2006-01-02 15:04:05.000000",
	}
}

// NewCustomTimeFormatter creates a custom timestamp text formatter.
func NewCustomTimeFormatter(tsFmt string) *Formatter {
	return &Formatter{
		RecordFormat: "{time} {level} {message}",
		TimeFormat:   tsFmt,
	}
}

// NewJsonFormatter creates a JSON formatted log record.
func NewJsonFormatter() *Formatter {
	return &Formatter{
		RecordFormat: `{"ts":"{time}","lvl":"{level}",` +
			`"src":"{source}","msg":"{message}"}`,
		TimeFormat: "2006-01-02 15:04:05.000000",
		EscapeMsg:  true,
	}
}
