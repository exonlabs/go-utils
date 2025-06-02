// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"strings"
	"time"
)

// Formatter defines function for formatting log record into messages.
type Formatter func(ts time.Time, lvl Level, src, msg string) string

// FormatStdTime generates standard time format for log messages
// time format: 2006-01-02 15:04:05.000000
func FormatStdTime(t time.Time) string {
	y, m, d := t.Date()
	h, min, s := t.Clock()
	us := t.Nanosecond() / 1000 // microseconds

	return fmt.Sprintf(
		"%04d-%02d-%02d %02d:%02d:%02d.%06d", y, int(m), d, h, min, s, us)
}

// StdFormatter generates a standard text formatted log message.
// Format: {time} {level} [{source}] {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO [logger_name] log message
func StdFormatter(ts time.Time, lvl Level, src, msg string) string {
	return fmt.Sprintf("%s %-5s [%s] %s",
		FormatStdTime(ts), lvl, src, msg)
}

// BasicFormatter generates a basic formatted text log message.
// Format: {time} {level} {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO log message
func BasicFormatter(ts time.Time, lvl Level, src, msg string) string {
	return fmt.Sprintf("%s %-5s %s",
		FormatStdTime(ts), lvl, msg)
}

// RawFormatter generates a minimal formatted text log message.
// Format: {time} {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 log message
func RawFormatter(ts time.Time, lvl Level, src, msg string) string {
	return fmt.Sprintf("%s %s",
		FormatStdTime(ts), msg)
}

// JsonFormatter generates a JSON formatted text log message.
//
// Example:
//
//	{"time": "2006-01-02 15:04:05.000000", "level": "INFO", "source": "logger_name", "message": "log message"}
func JsonFormatter(ts time.Time, lvl Level, src, msg string) string {
	msg = strings.ReplaceAll(msg, `\`, `\\`)
	msg = strings.ReplaceAll(msg, `"`, `\"`)
	return fmt.Sprintf(
		`{"time": "%s", "level": "%s", "source": "%s", "message": "%s"}`,
		FormatStdTime(ts), lvl, src, msg)
}
