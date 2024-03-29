package main

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

var Formatter1 = &xlog.Formatter{
	RecordFormat: "{time} {level} -- {message}",
	TimeFormat:   "2006/01/02 15:04:05",
}

var Formatter2 = &xlog.Formatter{
	RecordFormat: "{time} -- [{level}] -- {message}",
	TimeFormat:   "2006-01-02 15:04:05",
}

var Formatter3 = &xlog.Formatter{
	RecordFormat: `{"time":"{time}", "level":"{level}", "message":"{message}"}`,
	TimeFormat:   "2006-01-02 15:04:05",
	EscapeMsg:    true,
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
	logger := xlog.NewStdoutLogger("main")
	logger.Level = xlog.DEBUG

	fmt.Println("\n* with default formatter:")
	log_messages(logger)

	fmt.Println("\n* with custom formatter1:")
	logger.SetFormatter(Formatter1)
	log_messages(logger)

	fmt.Println("\n* with custom formatter2:")
	logger.SetFormatter(Formatter2)
	log_messages(logger)

	fmt.Println("\n* with custom json formatter:")
	logger.SetFormatter(Formatter3)
	log_messages(logger)

	fmt.Println()
}
