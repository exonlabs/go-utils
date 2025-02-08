// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/logging"
)

func main() {
	logger := logging.NewStdoutLogger("main")
	logger.Level = logging.INFO
	logger.Prefix = "-- root handler -- "

	fmt.Println("\n* logging parent logger:")
	logger.Warn("logging root message type: %s", "warn")
	logger.Info("logging root message type: %s", "info")

	log1 := logger.ChildLogger("child1")
	fmt.Println("\n* logging child1 logger:")
	logger.Warn("logging root message type: %s", "warn")
	log1.Warn("logging child 1 message type: %s", "warn")
	logger.Info("logging root message type: %s", "info")
	log1.Info("logging child 1 message type: %s", "info")

	log2 := logger.ChildLogger("child2")
	log2.Level = logging.WARN
	log2.Prefix = "-- child2 -- "
	log2.SetFormatter(logging.BasicFormatter)
	log2.SetHandler(logging.NewStdoutHandler())

	fmt.Println("\n* logging child2 logger level:WARN (+handlers):")
	logger.Warn("logging root message type: %s", "warn")
	log1.Warn("logging child 1 message type: %s", "warn")
	log2.Warn("logging child 2 message type: %s (should print twice)", "warn")
	logger.Info("logging root message type: %s", "info")
	log1.Info("logging child 1 message type: %s", "info")
	log2.Info("logging child 2 message type: %s", "info (should print once)")

	fmt.Println()
}
