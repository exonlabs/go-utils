// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xlog_test

import (
	"os"
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/xlog"
)

func ExampleNewLogger() {
	logger := xlog.NewLogger("main")

	stdhnd := xlog.NewStdoutHandler()
	logger.AddHandler(stdhnd)

	logger.Level = xlog.INFO
	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug") // dropped as Level=INFO

	logger.Level = xlog.DEBUG
	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug")
}

func ExampleNewStdoutLogger() {
	logger := xlog.NewStdoutLogger("main")
	logger.Level = xlog.DEBUG

	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug")
}

func ExampleNewFileLogger() {
	log_path := filepath.Join(os.TempDir(), "foo.log")

	logger := xlog.NewFileLogger("main", log_path)
	logger.Level = xlog.DEBUG

	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug")
}
