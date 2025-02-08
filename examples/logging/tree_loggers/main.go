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
	logger.Level = logging.DEBUG
	logger.SetFormatter(logging.NewCustomMsgFormatter(
		"{time} {level} [{source}] -- root handler, {message}"))

	fmt.Println("\n* logging parent logger:", logger.Name)
	logger.Warn("logging root message type: %s", "warn")
	logger.Info("logging root message type: %s", "info")

	log1 := logger.ChildLogger("child1")
	fmt.Println("\n* logging child  logger:", log1.Name)
	logger.Warn("logging root message type: %s", "warn")
	log1.Warn("logging child 1 message type: %s", "warn")
	logger.Info("logging root message type: %s", "info")
	log1.Info("logging child 1 message type: %s", "info")

	log2 := logger.ChildLogger("child2")
	log2.Level = logging.WARN
	log2.SetFormatter(logging.NewCustomMsgFormatter(
		"{time} {level} ----- child2 handler, {message}"))
	fmt.Println("\n* logging child 2 logger (+handlers):", log2.Name)
	logger.Warn("logging root message type: %s", "warn")
	log1.Warn("logging child 1 message type: %s", "warn")
	log2.Warn("logging child 2 message type: %s", "warn")
	logger.Info("logging root message type: %s", "info")
	log1.Info("logging child 1 message type: %s", "info")
	log2.Info("logging child 2 message type: %s", "info")

	fmt.Println()
}
