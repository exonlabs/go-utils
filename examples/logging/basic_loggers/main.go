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
	logger.Trace1("logging message type: %s", "trace1")
	logger.Trace2("logging message type: %s", "trace2")
	logger.Trace3("logging message type: %s", "trace3")
}

func main() {
	logger := logging.NewStdoutLogger("main")

	fmt.Println("\n* logging level: TRACE2")
	logger.Level = logging.TRACE2
	log_messages(logger)

	fmt.Println("\n* logging level: DEBUG")
	logger.Level = logging.DEBUG
	log_messages(logger)

	// adjust formatters
	fmt.Println("\n-- logging without source formatter --")
	logger.SetFormatter(logging.NewSimpleFormatter())

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
	fmt.Println("\n-- logging json formatter --")
	logger.SetFormatter(logging.NewJsonFormatter())
	logger.Level = logging.TRACE3
	log_messages(logger)

	fmt.Println()
}
