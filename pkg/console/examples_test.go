// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package console_test

import (
	"fmt"

	"github.com/exonlabs/go-utils/pkg/console"
)

func ExampleConsole_ReadValue() {
	con, _ := console.NewTermConsole()
	defer con.Close()

	// read input with optional value
	name, _ := con.ReadValue("Enter your name", "Anonymous")
	fmt.Println(name)

	// read required input
	name, _ = con.Required().ReadValue("Enter your name", "")
	fmt.Println(name)

	// read a required input with echo off (hidden input)
	passwd, _ := con.Hidden().Required().ReadValue("Enter password", "")
	fmt.Println(passwd)
}

func ExampleConsole_SelectValue() {
	con, _ := console.NewTermConsole()
	defer con.Close()

	color, _ := con.SelectValue("Select color?",
		[]string{"Red", "Blue", "Green"}, "Red")
	fmt.Println(color)
}

func ExampleConsole_SelectYesNo() {
	con, _ := console.NewTermConsole()
	defer con.Close()

	yes, _ := con.SelectYesNo("Continue?", "y")
	fmt.Println(yes)
}
