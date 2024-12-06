// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package console_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exonlabs/go-utils/pkg/console"
)

type MockHandler struct {
	input    string
	writeErr error
	readErr  error
	closeErr error
	writeBuf bytes.Buffer
}

func (m *MockHandler) Read(msg string) (string, error) {
	m.writeBuf.WriteString(msg)
	if m.readErr != nil {
		return "", m.readErr
	}
	return m.input, nil
}

func (m *MockHandler) ReadHidden(msg string) (string, error) {
	m.writeBuf.WriteString(msg)
	if m.readErr != nil {
		return "", m.readErr
	}
	return m.input, nil
}

func (m *MockHandler) Write(msg string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.writeBuf.WriteString(msg)
	return nil
}

func (m *MockHandler) Close() error {
	return m.closeErr
}

func TestConsole_ReadValue(t *testing.T) {
	mockHandler := &MockHandler{input: "test value"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadValue("Enter value", "")
	require.NoError(t, err)
	assert.Equal(t, "test value", val, "Expected value to match")
}

func TestConsole_ReadValue_Default(t *testing.T) {
	mockHandler := &MockHandler{input: ""}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadValue("Enter value", "default")
	require.NoError(t, err)
	assert.Equal(t, "default", val, "Expected default value to be returned")
}

func TestConsole_Required(t *testing.T) {
	mockHandler := &MockHandler{input: ""}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	_, err = con.Required().ReadValue("Enter value", "")
	require.Error(t, err)
	assert.Equal(t, "failed to get a valid input", err.Error())
}

func TestConsole_ReadHidden(t *testing.T) {
	mockHandler := &MockHandler{input: "secret"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.Hidden().ReadValue("Enter password", "")
	require.NoError(t, err)
	assert.Equal(t, "secret", val)
}

func TestConsole_ReadNumber(t *testing.T) {
	mockHandler := &MockHandler{input: "42"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadNumber("Enter number", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(42), val)
}

func TestConsole_ReadDecimal(t *testing.T) {
	mockHandler := &MockHandler{input: "42.345"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadDecimal("Enter decimal", 3, 0)
	require.NoError(t, err)
	assert.Equal(t, float64(42.345), val)
}

func TestConsole_ReadDecimal_InvalidInput(t *testing.T) {
	mockHandler := &MockHandler{input: "invalid"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadDecimal("Enter decimal", 3, 1.23)
	require.Error(t, err, "Expected error for invalid decimal input")
	assert.Equal(t, float64(0), val, "Expected 0 value to be returned on error")
}

func TestConsole_ReadDecimal_Default(t *testing.T) {
	mockHandler := &MockHandler{input: ""}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	val, err := con.ReadDecimal("Enter decimal", 3, 9.99)
	require.NoError(t, err)
	assert.Equal(t, 9.99, val, "Expected default decimal value to be returned")
}

func TestConsole_ReadDecimal_Error(t *testing.T) {
	mockHandler := &MockHandler{readErr: errors.New("read error")}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	_, err = con.ReadDecimal("Enter decimal", 3, 0.0)
	require.Error(t, err, "Expected error on read failure")
	assert.Equal(t, "failed to get a valid input", err.Error(),
		"Expected error message to match")
}

func TestConsole_SelectValue(t *testing.T) {
	mockHandler := &MockHandler{input: "2"}
	options := []string{"option1", "option2", "option3"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	_, err = con.SelectValue("Choose option", options, "")
	require.Error(t, err, "Expected error for invalid selection")
}

func TestConsole_SelectValue_InvalidSelection(t *testing.T) {
	mockHandler := &MockHandler{input: "invalid"}
	options := []string{"option1", "option2", "option3"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	_, err = con.SelectValue("Choose option", options, "")
	require.Error(t, err, "Expected error for invalid selection")
}

func TestConsole_SelectValue_Default(t *testing.T) {
	mockHandler := &MockHandler{input: ""}
	options := []string{"option1", "option2", "option3"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	selection, err := con.SelectValue("Choose option", options, "option1")
	require.NoError(t, err)
	assert.Equal(t, "option1", selection, "Expected default option to be returned")
}

func TestConsole_SelectValue_EmptyOptions(t *testing.T) {
	mockHandler := &MockHandler{input: "1"}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	_, err = con.SelectValue("Choose option", []string{}, "")
	require.Error(t, err, "Expected error when no options are provided")
}

func TestConsole_WriteError(t *testing.T) {
	mockHandler := &MockHandler{writeErr: errors.New("write error")}
	_, err := console.New(mockHandler)
	require.NoError(t, err)

	err = mockHandler.Write("Test message")
	assert.Error(t, err, "write error")
}

func TestConsole_Close(t *testing.T) {
	mockHandler := &MockHandler{closeErr: nil}
	con, err := console.New(mockHandler)
	require.NoError(t, err)

	err = con.Close()
	assert.NoError(t, err)
}
