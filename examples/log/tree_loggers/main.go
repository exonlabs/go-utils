package main

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

func main() {
	logger := xlog.NewStdoutLogger("main")
	logger.Level = xlog.DEBUG
	logger.SetFormatter(xlog.CustomMsgFormatter(
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
	log2.Level = xlog.WARN
	log2.SetFormatter(xlog.CustomMsgFormatter(
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
