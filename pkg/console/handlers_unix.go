// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package console

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// TermHandler is a terminal-based implementation of the Handler interface.
// It uses the 'golang.org/x/term' package for reading input from the terminal.
type TermHandler struct {
	tm *term.Terminal
}

// NewTermHandler creates and returns a new TermHandler for reading from
// and writing to the terminal.
func NewTermHandler() (*TermHandler, error) {
	return &TermHandler{
		tm: term.NewTerminal(os.Stdin, ""),
	}, nil
}

// Close implements the Handler interface but does not need to perform any
// action for TermHandler.
func (h *TermHandler) Close() error {
	return nil
}

// Read prompts the user for input and returns the trimmed result.
// It sets the terminal to raw mode while reading.
func (h *TermHandler) Read(msg string) (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to set terminal to raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	h.tm.SetPrompt(msg)
	input, err := h.tm.ReadLine()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// ReadHidden prompts the user for hidden input (e.g., for passwords)
// without echoing it back to the terminal.
func (h *TermHandler) ReadHidden(msg string) (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to set terminal to raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Disable echoing
	input, err := h.tm.ReadPassword(msg)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// Write writes a message to the console.
func (h *TermHandler) Write(msg string) error {
	_, err := os.Stdout.WriteString(msg)
	if err != nil {
		return fmt.Errorf("failed to write to console: %v", err)
	}
	return nil
}
