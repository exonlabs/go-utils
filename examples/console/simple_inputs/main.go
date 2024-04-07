package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/exonlabs/go-utils/pkg/unix/xterm"
)

// represent results
func print_val(value any, err error) {
	if err != nil {
		print_err(err)
		return
	}
	val := fmt.Sprint(value)
	if len(val) == 0 {
		val = "<empty>"
	}
	fmt.Printf("  * Value: %v\n", val)
}

// represent error or exit
func print_err(err error) {
	if strings.Contains(err.Error(), "EOF") {
		fmt.Print("\n--exit--\n\n")
		os.Exit(0)
	}
	fmt.Printf("Error: %v\n", err)
}

func main() {
	con := xterm.NewConsole()

	fmt.Println()
	print_val(con.Required().ReadValue("Enter required string", nil))
	fmt.Println()
	print_val(con.Required().ReadValue(
		"Enter required string with default", "default val"))

	fmt.Println()
	print_val(con.ReadValue("Enter optional string", ""))
	fmt.Println()
	print_val(con.ReadValue(
		"Enter optional string with default", "default val"))

	fmt.Println()
	print_val(con.Required().
		Regex("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+$").
		ReadValue("[input validation] Enter email (user@domain)", nil))

	fmt.Println()
	print_val(con.Required().
		Regex("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}"+
			"(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$").
		ReadValue("[input validation] Enter IPv4 (x.x.x.x)", nil))

	fmt.Println()
	print_val(con.Hidden().ReadValue("Test hidden input text", ""))
	fmt.Println()
	print_val(con.Required().Hidden().
		ReadValue("Test required hidden input", nil))
	fmt.Println()
	print_val(con.Hidden().
		ReadValue("Test hidden input with default", "default val"))

	fmt.Println()
	if res, err := con.Required().Hidden().ReadValue(
		"[input with confirm] Enter password", nil); err != nil {
		print_err(err)
	} else {
		if err := con.Required().Hidden().
			ConfirmValue("Confirm password", res); err != nil {
			print_err(err)
		} else {
			fmt.Printf("  * Password: %v\n", res)
		}
	}

	fmt.Println()
	print_val(con.Required().ReadNumber("Enter required number", nil))
	fmt.Println()
	print_val(con.ReadNumber("Enter optional number", 1234))
	fmt.Println()
	print_val(con.Required().ReadNumberWLimit(
		"Enter required number (-100 <= N <= 100)", nil, -100, 100))
	fmt.Println()
	print_val(con.Required().ReadNumberWLimit(
		"Enter required number with default (-100 <= N <= 100)",
		0, -100, 100))

	fmt.Println()
	print_val(con.Required().ReadDecimal(
		"Enter required decimal", 3, nil))
	fmt.Println()
	print_val(con.ReadDecimal("Enter optional decimal", 3, 1234))
	fmt.Println()
	print_val(con.Required().ReadDecimalWLimit(
		"Enter required decimal (-10.55 <= N <= 10.88)",
		2, nil, -10.55, 10.88))
	fmt.Println()
	print_val(con.Required().ReadDecimalWLimit(
		"Enter required decimal with default (0 <= N <= 100)",
		2, 20.0, 0, 100))

	fmt.Println()
	print_val(con.Required().SelectValue("Select from list",
		[]string{"val1", "val2", "val3"}, nil))
	fmt.Println()
	print_val(con.Required().SelectValue(
		"Select from list with default",
		[]string{"val1", "val2", "val3"}, "val2"))
	fmt.Println()
	print_val(con.Required().SelectYesNo("Select Yes/No", nil))
	fmt.Println()
	print_val(con.Required().SelectYesNo(
		"Select Yes/No with default", "n"))

	fmt.Println()
}
