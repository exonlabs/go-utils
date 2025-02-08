// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/exonlabs/go-utils/pkg/logging"
)

func log_messages(logger *logging.Logger) {
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

	logger := logging.NewStdoutLogger("main")
	logger.Level = logging.DEBUG

	hnd := logging.NewFileHandler(log_path)
	logger.AddHandler(hnd)

	fmt.Println("\n* logging stdout")
	log_messages(logger)

	fmt.Println("\n* logs in file:", log_path)
	b, err := os.ReadFile(log_path)
	if err != nil {
		fmt.Printf("Error!! failed read log file, %s", err.Error())
	}
	fmt.Println(strings.TrimSpace(string(b)))

	fmt.Println()
}
