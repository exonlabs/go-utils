// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exonlabs/go-utils/pkg/logging"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) HandleMessage(msg string) error {
	args := m.Called(msg)
	return args.Error(0)
}

func TestLogger(t *testing.T) {
	handler := new(MockHandler)
	logger := logging.NewStdoutLogger("TestLogger")
	logger.SetHandler(handler)

	assert.Equal(t, logger.Name(), "TestLogger")

	// Test logging at different levels
	handler.On("HandleMessage", mock.Anything).Return(nil).Times(0).Once()

	assert.NoError(t, logger.Info("Info message"))
	assert.NoError(t, logger.Debug("Error message"))
	assert.NoError(t, logger.Trace("Warning message"))

	// Test panic logging
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Panic("Panic message"))

	// Test log level filtering
	logger.Level = logging.ERROR
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Fatal("Should log fatal"))
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Error("Should log error"))

	assert.NoError(t, logger.Warn("Should not log warning"))
	assert.NoError(t, logger.Info("Should not log info"))
	assert.NoError(t, logger.Debug("Should not log debug"))
	assert.NoError(t, logger.Trace("Should not log trace"))
}

func TestFileLogger(t *testing.T) {
	handler := new(MockHandler)
	logger := logging.NewFileLogger("TestLogger", "")
	logger.SetHandler(handler)

	assert.Equal(t, logger.Name(), "TestLogger")

	// Test logging at different levels
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()

	assert.NoError(t, logger.Info("Info message"))
	assert.NoError(t, logger.Debug("Error message"))
	assert.NoError(t, logger.Trace("Warning message"))

	// Test panic logging
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Panic("Panic message"))

	// Test log level filtering
	logger.Level = logging.ERROR
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Fatal("Should log fatal"))
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Error("Should log error"))

	assert.NoError(t, logger.Warn("Should not log warning"))
	assert.NoError(t, logger.Info("Should not log info"))
	assert.NoError(t, logger.Debug("Should not log debug"))
	assert.NoError(t, logger.Trace("Should not log trace"))
}

func TestChildAndSubLoggers(t *testing.T) {
	parentLogger := logging.NewStdoutLogger("Parent")
	parentLogger.Level = logging.DEBUG

	assert.Equal(t, parentLogger.Name(), "Parent")

	childLogger := parentLogger.ChildLogger("Child")
	assert.Equal(t, childLogger.Name(), "Child")

	subLogger := parentLogger.SubLogger("Sub")
	assert.Equal(t, subLogger.Name(), "Parent")

	// Mock handler for child logger
	handler := new(MockHandler)
	childLogger.SetHandler(handler)
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()

	// Test logging from child logger
	assert.NoError(t, childLogger.Debug("Debug message from child"))

	// Test logging from sub logger
	handler.On("HandleMessage", mock.Anything).Return(nil).Once()
	assert.NoError(t, subLogger.Info("Info message from sublogger"))
}

// TestFileHandler tests writing log messages to a file.
func TestFileHandler(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_log_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Clean up after test

	handler := logging.NewFileHandler(tempFile.Name())

	// Test writing a log record to the file
	message := "This is a test log message."
	err = handler.HandleMessage(message)
	assert.NoError(t, err)

	// Read back the contents of the file
	content, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	// Assert that the file contains the expected message
	assert.Contains(t, string(content), message)
}

func TestFormatterEmit(t *testing.T) {
	tests := []struct {
		name      string
		formatter logging.Formatter
		level     logging.Level
		source    string
		message   string
		expected  string
	}{
		{
			name:      "Standard Formatter",
			formatter: logging.StdFormatter,
			level:     logging.INFO,
			source:    "TestSource",
			message:   "Test message",
			expected:  "2006-01-02 15:04:05.000000 INFO  [TestSource] Test message",
		},
		{
			name:      "Basic Formatter",
			formatter: logging.BasicFormatter,
			level:     logging.DEBUG,
			source:    "",
			message:   "Debugging info",
			expected:  "2006-01-02 15:04:05.000000 DEBUG Debugging info",
		},
		{
			name:      "Raw Formatter",
			formatter: logging.RawFormatter,
			level:     logging.ERROR,
			source:    "",
			message:   "An error occurred",
			expected:  "2006-01-02 15:04:05.000000 An error occurred",
		},
		{
			name:      "JSON Formatter",
			formatter: logging.JsonFormatter,
			level:     logging.WARN,
			source:    "JsonSource",
			message:   "Warning occurred",
			expected:  `{"time": "2006-01-02 15:04:05.000000", "level": "WARN", "source": "JsonSource", "message": "Warning occurred"}`,
		},
	}

	ts, _ := time.Parse("2006-01-02 15:04:05.000000", "2006-01-02 15:04:05.000000")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.formatter(ts, tt.level, tt.source, tt.message)
			assert.Equal(t, formatted, tt.expected)
		})
	}
}
