// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package fsx_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/exonlabs/go-utils/pkg/abc/fsx"
)

func ExampleParsePath() {
	paths := []string{
		"",
		" ",
		"/",
		"/tmp/..",
		"/tmp/dir1/../",
		"/tmp/dir1/../srcfile.txt",
	}
	for _, path := range paths {
		p, err := fsx.ParsePath(path)
		if err == nil {
			fmt.Printf("\"%s\" --> \"%s\"\n", path, p)
		} else {
			fmt.Printf("\"%s\" --> err: %s\n", path, err.Error())
		}
	}

	// Results:
	// "" --> err: invalid path
	// " " --> err: invalid path
	// "/" --> "/"
	// "/tmp/.." --> "/"
	// "/tmp/dir1/../" --> "/tmp"
	// "/tmp/dir1/../srcfile.txt" --> "/tmp/srcfile.txt"
}

func ExampleCopy() {
	tmpDir := os.TempDir()
	srcPath := filepath.Join(tmpDir, "srcfile.txt")
	dstPath := filepath.Join(tmpDir, "srcfile_copy.txt")

	err := fsx.Copy(srcPath, dstPath)
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleCopyDir() {
	tmpDir := os.TempDir()
	srcDir := filepath.Join(tmpDir, "srcdir")
	dstDir := filepath.Join(tmpDir, "srcdir_copy")

	err := fsx.CopyDir(srcDir, dstDir)
	if err != nil {
		fmt.Println(err)
	}
}
