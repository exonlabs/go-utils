// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package namedpipes_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/abc/dictx"
	"github.com/exonlabs/go-utils/pkg/unix/namedpipes"
)

func Example() {
	// Path to the pipe (change the path accordingly for your platform)
	pipePath := "/tmp/test_pipe"

	// Create a named pipe
	err := namedpipes.Create(pipePath, 0o666)
	if err != nil {
		fmt.Printf("Failed to create pipe: %v\n", err)
		return
	}
	defer namedpipes.Delete(pipePath) // Ensure the pipe is deleted after use

	// Set up options for the pipe
	options := dictx.Dict{
		"poll_timeout":   0.1,
		"poll_chunksize": 4096,
	}

	// Create a new pipe instance
	pipe := namedpipes.New(pipePath, options)

	// Write data to the pipe
	dataToWrite := []byte("Hello, named pipe!")
	err = pipe.Write(dataToWrite, 1.0)
	if err != nil {
		fmt.Printf("Failed to write to pipe: %v\n", err)
		return
	}

	// Read data from the pipe
	dataRead, err := pipe.Read(1.0)
	if err != nil {
		fmt.Printf("Failed to read from pipe: %v\n", err)
		return
	}

	// Output the read data
	fmt.Printf("Data read from pipe: %s\n", dataRead)
}
