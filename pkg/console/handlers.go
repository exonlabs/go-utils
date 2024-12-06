// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package console

// Handler defines the interface for reading and writing from/to the console.
// It includes methods to read regular and hidden input, write output, and close the handler.
type Handler interface {
	Read(string) (string, error)       // Read reads input with a prompt.
	ReadHidden(string) (string, error) // ReadHidden reads input without echoing (for passwords).
	Write(string) error                // Write outputs a formatted message to the console.
	Close() error                      // Close cleans up any resources used by the handler.
}
