// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

//go:build !windows

package psutils

import (
	"fmt"
	"os"
)

// SetProcTitle sets the process title in the `/proc` filesystem on Unix-like systems.
// It modifies the content of the `/proc/[pid]/comm` file, where [pid] is the process ID
// of the current process. The title is used to identify the process.
// It returns an error if the title is empty or if it fails to write to the `/proc` file.
func SetProcTitle(title string) error {
	if title == "" {
		return fmt.Errorf("process title should not be empty")
	}

	// Construct the path to the /proc/[pid]/comm file
	path := fmt.Sprintf("/proc/%d/comm", os.Getpid())

	// Open the file for writing
	f, err := os.OpenFile(path, os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open /proc/comm: %v", err)
	}
	defer f.Close()

	// Write the title to the file
	if _, err := f.WriteString(title); err != nil {
		return fmt.Errorf("failed to write process title: %v", err)
	}

	return nil
}
