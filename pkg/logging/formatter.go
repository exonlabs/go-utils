// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"strings"
	"time"
)

const (
	// standard time format for log messages
	STD_TIME_FORMAT = "2006-01-02 15:04:05.000000"
)

// Formatter defines function for formatting log record into messages.
type Formatter func(ts time.Time, lvl int, src, msg string) string

// StdFormatter generates a standard text formatted log message.
// Format: {time} {level} [{source}] {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO [logger_name] log message
func StdFormatter(ts time.Time, lvl int, src, msg string) string {
	return fmt.Sprintf("%s %-5s [%s] %s",
		ts.Format(STD_TIME_FORMAT), LEVEL(lvl), src, msg)
}

// BasicFormatter generates a basic formatted text log message.
// Format: {time} {level} {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 INFO log message
func BasicFormatter(ts time.Time, lvl int, src, msg string) string {
	return fmt.Sprintf("%s %-5s %s",
		ts.Format(STD_TIME_FORMAT), LEVEL(lvl), msg)
}

// RawFormatter generates a minimal formatted text log message.
// Format: {time} {message}
//
// Example:
//
//	2006-01-02 15:04:05.000000 log message
func RawFormatter(ts time.Time, lvl int, src, msg string) string {
	return fmt.Sprintf("%s %s",
		ts.Format(STD_TIME_FORMAT), msg)
}

// JsonFormatter generates a JSON formatted text log message.
//
// Example:
//
//	{"time": "2006-01-02 15:04:05.000000", "level": "INFO", "source": "logger_name", "message": "log message"}
func JsonFormatter(ts time.Time, lvl int, src, msg string) string {
	msg = strings.ReplaceAll(msg, `\`, `\\`)
	msg = strings.ReplaceAll(msg, `"`, `\"`)
	return fmt.Sprintf(
		`{"time": "%s", "level": "%s", "source": "%s", "message": "%s"}`,
		ts.Format(STD_TIME_FORMAT), LEVEL(lvl), src, msg)
}
