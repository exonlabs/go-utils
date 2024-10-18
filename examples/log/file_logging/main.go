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
}

func main() {
	log_path := filepath.Join(os.TempDir(), "foobar.log")
	defer os.Remove(log_path)

	logger := xlog.NewLogger("main")
	logger.Level = xlog.DEBUG

	hnd1 := xlog.NewStdoutHandler()
	logger.AddHandler(hnd1)

	hnd2 := xlog.NewFileHandler(log_path)
	logger.AddHandler(hnd2)

	fmt.Println("\n* logging stdout")
	log_messages(logger)

	fmt.Println("\n* logs in file:", log_path)
	f, err := os.OpenFile(log_path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Error!! failed open log file, %s", err.Error())
	}
	defer f.Close()
	b := make([]byte, 10240)
	_, err = f.Read(b)
	if err != nil {
		fmt.Printf("Error!! failed read log file, %s", err.Error())
	}
	fmt.Println(string(b))

	fmt.Println()
}
