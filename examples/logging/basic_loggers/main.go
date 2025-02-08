// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/logging"
)

func log_messages(logger *logging.Logger) {
	logger.Panic("logging message type: %s", "panic")
	logger.Fatal("logging message type: %s", "fatal")
	logger.Error("logging message type: %s", "error")
	logger.Warn("logging message type: %s", "warn")
	logger.Info("logging message type: %s", "info")
	logger.Debug("logging message type: %s", "debug")
	logger.Trace("logging message type: %s", "trace")
}

func main() {
	logger := logging.NewStdoutLogger("main")

	fmt.Println("\n-- logging with default formatter --")

	fmt.Println("\n* logging level: TRACE")
	logger.Level = logging.TRACE
	log_messages(logger)

	fmt.Println("\n* logging level: INFO")
	logger.Level = logging.INFO
	log_messages(logger)

	// adjust formatters
	fmt.Println("\n-- logging with simple formatter --")
	logger.SetFormatter(logging.BasicFormatter)

	fmt.Println("\n* logging level: ERROR")
	logger.Level = logging.ERROR
	log_messages(logger)

	fmt.Println("\n* logging level: FATAL")
	logger.Level = logging.FATAL
	log_messages(logger)

	fmt.Println("\n* logging level: PANIC")
	logger.Level = logging.PANIC
	log_messages(logger)

	// adjust formatters
	fmt.Println("\n-- logging with json formatter --")
	logger.SetFormatter(logging.JsonFormatter)

	fmt.Println("\n* logging level: INFO")
	logger.Level = logging.INFO
	log_messages(logger)

	fmt.Println()
}
