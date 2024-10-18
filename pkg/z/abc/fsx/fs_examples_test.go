// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package xos_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/xos"
)

func ExampleParsePath() {
	paths := []string{
		"",
		" ",
		"/",
		"/tmp/..",
		"/tmp/dir_1/../",
		"/tmp/dir_1/../file_1",
	}
	for _, path := range paths {
		p, err := xos.ParsePath(path)
		if err == nil {
			fmt.Printf("\"%s\" --> \"%s\"\n", path, p)
		} else {
			fmt.Printf("\"%s\" --> err: %s\n", path, err.Error())
		}
	}

	// Output:
	// "" --> err: invalid path
	// " " --> err: invalid path
	// "/" --> "/"
	// "/tmp/.." --> "/"
	// "/tmp/dir_1/../" --> "/tmp"
	// "/tmp/dir_1/../file_1" --> "/tmp/file_1"
}

func ExampleCopy() {
	src := "/home/user/myfile.txt"
	dst := "/tmp/myfile_copy.txt"

	err := xos.Copy(src, dst)
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleCopyDir() {
	src := "/home/user/mydir"
	dst := "/tmp/mydir_copy"

	err := xos.CopyDir(src, dst)
	if err != nil {
		fmt.Println(err)
	}
}
