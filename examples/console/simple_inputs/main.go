// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/exonlabs/go-utils/pkg/console"
)

// represent results
func printValue(value any, err error) {
	if err != nil {
		printError(err)
		return
	}
	val := fmt.Sprint(value)
	if len(val) == 0 {
		val = "<empty>"
	}
	fmt.Printf("  * Value: %v\n\n", val)
}

// represent error or exit
func printError(err error) {
	if strings.Contains(err.Error(), "EOF") {
		fmt.Print("\n--exit--\n\n")
		os.Exit(0)
	}
	fmt.Printf("Error: %v\n\n", err)
}

func main() {
	con, _ := console.NewTermConsole()
	defer con.Close()

	fmt.Println()

	printValue(con.Required().
		ReadValue("Enter required string", ""))
	printValue(con.Required().
		ReadValue("Enter required string with default", "default val"))

	printValue(con.
		ReadValue("Enter optional string", ""))
	printValue(con.
		ReadValue("Enter optional string with default", "default val"))

	printValue(con.Hidden().
		ReadValue("Test hidden input text", ""))
	printValue(con.Required().Hidden().
		ReadValue("Test required hidden input", ""))
	printValue(con.Hidden().
		ReadValue("Test hidden input with default", "default val"))

	if res, err := con.Required().Hidden().
		ReadValue("Enter password", ""); err != nil {
		printError(err)
	} else {
		if err := con.Required().Hidden().
			ConfirmValue("Confirm password", res); err != nil {
			printError(err)
		} else {
			fmt.Printf("  * Password: %v\n\n", res)
		}
	}

	printValue(con.Required().
		Regex("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+$").
		ReadValue("Enter email (user@domain)", ""))

	printValue(con.Required().
		Regex("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}"+
			"(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$").
		ReadValue("Enter IPv4 (x.x.x.x)", ""))

	printValue(con.Required().
		ReadNumber("Enter required number", 0))
	printValue(con.
		ReadNumber("Enter optional number", 1234))
	printValue(con.Required().
		ReadNumber("Enter required number (-100 <= N <= 100)",
			0, -100, 100))
	printValue(con.
		ReadNumber("Enter number with default (-100 <= N <= 100)",
			0, -100, 100))

	printValue(con.Required().
		ReadDecimal("Enter required decimal", 3, 0))
	printValue(con.
		ReadDecimal("Enter optional decimal", 3, 1234))
	printValue(con.Required().
		ReadDecimal("Enter required decimal (-10.55 <= N <= 10.88)",
			2, 0, -100.55, 100.88))
	printValue(con.
		ReadDecimal("Enter decimal with default (0 <= N <= 100)",
			2, 20.0, 0, 100))

	printValue(con.Required().
		SelectValue("Select from list", []string{"val1", "val2", "val3"}, ""))
	printValue(con.Required().
		SelectValue("Select from list with default",
			[]string{"val1", "val2", "val3"}, "val2"))
	printValue(con.Required().
		SelectYesNo("Select Yes/No", ""))
	printValue(con.Required().
		SelectYesNo("Select Yes/No with default", "n"))

	fmt.Println()
}
