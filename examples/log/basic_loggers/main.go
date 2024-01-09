package main

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

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

	fmt.Println("\n* logging level: TRACE")
	logger.Level = xlog.TRACE2
	log_messages(logger)

	fmt.Println("\n* logging level: DEBUG")
	logger.Level = xlog.DEBUG
	log_messages(logger)

	// adjust formatters
	logger.SetFormatter(xlog.NewStdFormatter(
		"({time}) {level} [{source}] {message}",
		"2006/01/02 15:04:05.000000"))

	fmt.Println("\n-- with custom logging format --")
	fmt.Println("\n* logging level: ERROR")
	logger.Level = xlog.ERROR
	log_messages(logger)

	fmt.Println("\n* logging level: FATAL")
	logger.Level = xlog.FATAL
	log_messages(logger)

	fmt.Println("\n* logging level: PANIC")
	logger.Level = xlog.PANIC
	log_messages(logger)

	fmt.Println()
}
