package main

import (
	"fmt"
	"strings"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

type Formatter1 struct{}

func (f Formatter1) ParseRecord(r xlog.Record) string {
	return strings.NewReplacer(
		"{time}", r.Time.Format("2006/01/02 15:04:05"),
		"{level}", xlog.StringLevel(r.Level),
		"{source}", r.Source,
		"{message}", fmt.Sprintf(r.Message, r.MsgArgs...),
	).Replace("{time} {level} -- {message}")
}

type Formatter2 struct{}

func (f Formatter2) ParseRecord(r xlog.Record) string {
	return strings.NewReplacer(
		"{time}", r.Time.Format("2006-01-02 15:04:05"),
		"{level}", xlog.StringLevel(r.Level),
		"{source}", r.Source,
		"{message}", fmt.Sprintf(r.Message, r.MsgArgs...),
	).Replace("{time} -- [{level}] -- {message}")
}

func log_messages(logger *xlog.Logger) {
	logger.Panic("logging message type: %s", "panic")
	logger.Fatal("logging message type: %s", "fatal")
	logger.Error("logging message type: %s", "error")
	logger.Warn("logging message type: %s", "warn")
	logger.Info("logging message type: %s", "info")
	logger.Debug("logging message type: %s", "debug")
	logger.Trace1("logging message type: %s", "trace1")
	logger.Trace2("logging message type: %s", "trace2")
	logger.Trace3("logging message type: %s", "trace3")
	logger.Trace4("logging message type: %s", "trace4")
}

func main() {
	logger := xlog.GetLogger()
	logger.Level = xlog.DEBUG

	fmt.Println("\n* with default formatter:")
	log_messages(logger)

	hnd1 := xlog.NewStdoutHandler()
	hnd1.SetFormatter(Formatter1{})
	logger.AddHandler(hnd1)

	fmt.Println("\n* with 1 handler using custom formatter:")
	log_messages(logger)

	hnd2 := xlog.NewStdoutHandler()
	hnd2.SetFormatter(Formatter2{})
	logger.AddHandler(hnd2)

	fmt.Println("\n* logging with 2 handlers using custom formatters:")
	log_messages(logger)

	fmt.Println()
}
