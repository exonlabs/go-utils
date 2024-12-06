// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package logging_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/exonlabs/go-utils/pkg/logging"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) HandleRecord(record string) error {
	args := m.Called(record)
	return args.Error(0)
}

func TestLogger(t *testing.T) {
	handler := new(MockHandler)
	logger := &logging.Logger{Name: "TestLogger"}
	logger.AddHandler(handler)

	// Test setting a new formatter
	formatter := logging.NewStdFormatter()
	logger.SetFormatter(formatter)

	// Test logging at different levels
	handler.On("HandleRecord", mock.Anything).Return(nil).Once()

	assert.NoError(t, logger.Info("Info message"))
	assert.NoError(t, logger.Debug("Error message"))
	assert.NoError(t, logger.Trace1("Warning message"))

	// Test panic logging
	handler.On("HandleRecord", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Panic("Panic message"))

	// Test log level filtering
	logger.Level = logging.ERROR
	handler.On("HandleRecord", mock.Anything).Return(nil).Once()
	assert.NoError(t, logger.Warn("Should log warning"))
	assert.NoError(t, logger.Error("Should log error"))
	assert.NoError(t, logger.Info("Should not log info"))
}

func TestChildAndSubLoggers(t *testing.T) {
	parentLogger := &logging.Logger{Name: "Parent"}
	parentLogger.Level = logging.DEBUG
	parentLogger.SetFormatter(logging.NewStdFormatter())

	childLogger := parentLogger.ChildLogger("Child")
	subLogger := parentLogger.SubLogger("Sub")

	// Mock handler for child logger
	handler := new(MockHandler)
	childLogger.AddHandler(handler)
	handler.On("HandleRecord", mock.Anything).Return(nil).Once()

	// Test logging from child logger
	assert.NoError(t, childLogger.Debug("Debug message from child"))

	// Test logging from sub logger
	handler.On("HandleRecord", mock.Anything).Return(nil).Once()
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
	err = handler.HandleRecord(message)
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
		formatter *logging.Formatter
		level     logging.Level
		source    string
		message   string
		expected  string
	}{
		{
			name:      "Standard Formatter",
			formatter: logging.NewStdFormatter(),
			level:     logging.INFO,
			source:    "TestSource",
			message:   "Test message",
			expected:  "2006-01-02 15:04:05.000000 INFO [TestSource] Test message",
		},
		{
			name:      "Simple Formatter",
			formatter: logging.NewSimpleFormatter(),
			level:     logging.DEBUG,
			source:    "",
			message:   "Debugging info",
			expected:  "2006-01-02 15:04:05.000000 DEBUG Debugging info",
		},
		{
			name:      "Basic Formatter",
			formatter: logging.NewBasicFormatter(),
			level:     logging.ERROR,
			source:    "",
			message:   "An error occurred",
			expected:  "2006-01-02 15:04:05.000000 An error occurred",
		},
		{
			name:      "Raw Formatter",
			formatter: logging.NewRawFormatter(),
			level:     logging.FATAL,
			source:    "",
			message:   "Fatal error!",
			expected:  "Fatal error!",
		},
		{
			name:      "JSON Formatter",
			formatter: logging.NewJsonFormatter(),
			level:     logging.WARN,
			source:    "JsonSource",
			message:   "Warning occurred",
			expected:  `{"ts":"2006-01-02 15:04:05.000000","lvl":"WARN","src":"JsonSource","msg":"Warning occurred"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := tt.formatter.Emit(tt.level, tt.source, tt.message)
			assert.Contains(t, formatted, tt.message)

			// For expected timestamp, we need to check that it contains a valid time format
			if tt.formatter.TimeFormat != "" {
				assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}`, formatted)
			}
		})
	}
}
