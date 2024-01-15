package main

import (
	"fmt"
	"os"
	"path/filepath"

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

	hnd1 := xlog.NewStdoutHandler()
	logger.AddHandler(hnd1)

	hnd2 := xlog.NewFileHandler(
		filepath.Join(os.TempDir(), "foobar.log"))
	hnd2.SetFormatter(xlog.NewStdFormatter(
		"{time} {level} [{source}] {message}",
		"2006-01-02 15:04:05.000000"))
	logger.AddHandler(hnd2)

	fmt.Println("\n* logging stdout and file:", hnd2.FilePath)
	log_messages(logger)

	fmt.Println()
}
