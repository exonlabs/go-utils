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
	logger := xlog.GetRootLogger()
	logger.Level = xlog.DEBUG
	logger.SetFormatter(xlog.NewStdFormatter(
		"{time} {level} [{source}] -- root handler, {message}",
		"2006-01-02 15:04:05.000000"))

	fmt.Println("\n* logging parent logger:", logger.Name)
	log_messages(logger)

	log1 := logger.NewChildLogger("child1")
	fmt.Println("\n* logging child logger:", log1.Name)
	log_messages(log1)

	log2 := logger.NewChildLogger("child2")
	log2.Level = xlog.WARN
	log2.SetFormatter(xlog.NewStdFormatter(
		"{time} {level} ({source}) ----- child2 handler, {message}",
		"2006-01-02 15:04:05.000000"))
	log2.AddHandler(xlog.NewStdoutHandler())
	fmt.Println("\n* logging child logger (+handlers):", log2.Name)
	log_messages(log2)

	log21 := log2.NewChildLogger("child21")
	log21.Level = xlog.INFO
	log21.SetFormatter(xlog.NewStdFormatter(
		"{time} {level} ({source}) -------- child21 handler, {message}",
		"2006-01-02 15:04:05.000000"))
	log21.AddHandler(xlog.NewStdoutHandler())
	fmt.Println("\n* logging subchild logger (+handlers):", log21.Name)
	log_messages(log21)

	fmt.Println()
}
