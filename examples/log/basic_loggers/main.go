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
	logger := xlog.NewStdoutLogger("main")

	fmt.Println("\n* logging level: TRACE2")
	logger.Level = xlog.TRACE2
	log_messages(logger)

	fmt.Println("\n* logging level: DEBUG")
	logger.Level = xlog.DEBUG
	log_messages(logger)

	// adjust formatters
	fmt.Println("\n-- logging without source formatter --")
	logger.SetFormatter(xlog.SimpleFormatter())

	fmt.Println("\n* logging level: ERROR")
	logger.Level = xlog.ERROR
	log_messages(logger)

	fmt.Println("\n* logging level: FATAL")
	logger.Level = xlog.FATAL
	log_messages(logger)

	fmt.Println("\n* logging level: PANIC")
	logger.Level = xlog.PANIC
	log_messages(logger)

	// adjust formatters
	fmt.Println("\n-- logging json formatter --")
	logger.SetFormatter(xlog.JsonFormatter())
	logger.Level = xlog.TRACE4
	log_messages(logger)

	fmt.Println()
}
