// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build windows

package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// TermHandler is a terminal-based implementation of the Handler interface.
type TermHandler struct{}

// NewTermHandler creates and returns a new TermHandler for reading from
// and writing to the terminal.
func NewTermHandler() (*TermHandler, error) {
	return &TermHandler{}, nil
}

// Close implements the Handler interface but does not need to perform any
// action for TermHandler.
func (h *TermHandler) Close() error {
	return nil
}

// Read prompts the user for input and returns the trimmed result.
func (h *TermHandler) Read(msg string) (string, error) {
	// Prompt the user for input
	if err := h.Write(msg); err != nil {
		return "", err
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %v", err)
	}

	return strings.TrimSpace(input), nil
}

// ReadHidden prompts the user for hidden input (e.g., for passwords)
// without echoing it back to the terminal.
func (h *TermHandler) ReadHidden(msg string) (string, error) {
	// Windows-specific hidden input
	handle := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return "", fmt.Errorf("unable to get console mode: %v", err)
	}
	defer windows.SetConsoleMode(handle, mode)

	// Disable echoing
	err := windows.SetConsoleMode(handle, mode&^windows.ENABLE_ECHO_INPUT)
	if err != nil {
		return "", fmt.Errorf("unable to set console mode: %v", err)
	}

	input, err := h.Read(msg)
	h.Write("\n\r")

	return input, err
}

// Write writes a message to the console.
func (h *TermHandler) Write(msg string) error {
	_, err := os.Stdout.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write to console: %v", err)
	}
	return nil
}
