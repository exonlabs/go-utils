// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging_test

import (
	"os"
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/logging"
)

func ExampleNewStdoutLogger() {
	logger := logging.NewStdoutLogger("main")
	logger.Level = logging.DEBUG

	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug")
}

func ExampleNewFileLogger() {
	log_path := filepath.Join(os.TempDir(), "foo.log")

	logger := logging.NewFileLogger("main", log_path)
	logger.Level = logging.DEBUG

	logger.Warn("logging message type: warn")
	logger.Info("logging message type: info")
	logger.Debug("logging message type: debug")
}
