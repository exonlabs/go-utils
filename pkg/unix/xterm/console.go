package xterm

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/term"
)

const (
	ESC_REDBRT = "\033[31;1m"
	ESC_BRIGHT = "\033[1m"
	ESC_RESET  = "\033[0m"
)

// parsing and validating input values
type parser = func(string) (any, error)

type Console struct {
	Prompt   string
	Trials   int
	required bool
	hidden   bool
	regex    string
}

func NewConsole() *Console {
	return &Console{
		Prompt: ">>",
		Trials: 3,
	}
}

// set required flag for current operation
func (con *Console) Required() *Console {
	con.required = true
	return con
}

// set hidden flag for current operation
func (con *Console) Hidden() *Console {
	con.hidden = true
	return con
}

// set regex validator for current operation
func (con *Console) Regex(regex string) *Console {
	con.regex = regex
	return con
}

// print error message
func (con *Console) print_error(msg string) {
	msg = fmt.Sprintf(" -- %v", msg)
	// print RED colored error message
	if runtime.GOOS != "windows" {
		msg = ESC_REDBRT + msg + ESC_RESET
	}
	fmt.Println(msg)
}

// read 1 line input from terminal
func (con *Console) read_line(msg string, defval any) (string, error) {
	// get term status to restore when finish
	oldstate, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldstate)

	prompt := fmt.Sprintf("%v %v ", con.Prompt, msg)
	// print BOLD/BRIGHT text prompt
	if runtime.GOOS != "windows" {
		prompt = ESC_BRIGHT + prompt + ESC_RESET
	}

	var val string
	tm := term.NewTerminal(os.Stdin, prompt)
	if con.hidden {
		val, err = tm.ReadPassword(prompt)
	} else {
		val, err = tm.ReadLine()
	}
	if err != nil {
		return "", err
	}
	val = strings.TrimSpace(val)
	if len(val) > 0 {
		return val, nil
	}

	if defval != nil {
		return fmt.Sprint(defval), nil
	}
	return "", nil
}

// get user input with multiple trials
func (con *Console) get_input(msg string, defval any, pr parser) (any, error) {
	defer func() {
		con.required = false
		con.hidden = false
		con.regex = ""
	}()

	msg += ":"
	if defval != nil {
		msg += fmt.Sprintf(" [%v]", defval)
	} else if !con.required {
		msg += " []"
	}

	for i := 0; i < con.Trials; i++ {
		input, err := con.read_line(msg, defval)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				fmt.Println()
				return nil, err
			}
			con.print_error("internal error, " + err.Error())
			continue
		}
		if len(input) == 0 {
			if con.required {
				con.print_error("input required, please enter value")
				continue
			}
			return defval, nil
		}
		if pr != nil {
			if val, err := pr(input); err != nil {
				con.print_error(err.Error())
				continue
			} else {
				return val, nil
			}
		}
		return input, nil
	}

	return nil, fmt.Errorf("failed to get valid input")
}

// read general string value
func (con *Console) ReadValue(msg string, defval any) (string, error) {
	var pr parser
	if len(con.regex) > 0 {
		pr = func(input string) (any, error) {
			return con.parse_regex(input, con.regex)
		}
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// confirm imput value
func (con *Console) ConfirmValue(msg, chkval string) error {
	defer func() {
		con.required = false
		con.hidden = false
		con.regex = ""
	}()

	msg += ":"
	for i := 0; i < con.Trials; i++ {
		input, err := con.read_line(msg, nil)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				fmt.Println()
				return err
			}
			con.print_error("internal error, " + err.Error())
			continue
		}
		if len(input) == 0 {
			con.print_error("empty input, please confirm value")
			continue
		}
		if input == chkval {
			return nil
		}
		con.print_error("value not matching, please try again")
	}

	return fmt.Errorf("failed to confirm value")
}

// read numeric integer value
func (con *Console) ReadNumber(msg string, defval any) (int64, error) {
	pr := func(input string) (any, error) {
		return con.parse_number(input, nil, nil)
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}
func (con *Console) ReadNumberWLimit(
	msg string, defval any, vmin, vmax int64) (int64, error) {
	pr := func(input string) (any, error) {
		return con.parse_number(input, &vmin, &vmax)
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}

// read numeric decimal value
func (con *Console) ReadDecimal(
	msg string, decimals int, defval any) (float64, error) {
	pr := func(input string) (any, error) {
		return con.parse_decimal(input, decimals, nil, nil)
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}
func (con *Console) ReadDecimalWLimit(msg string, decimals int,
	defval any, vmin, vmax float64) (float64, error) {
	pr := func(input string) (any, error) {
		return con.parse_decimal(input, decimals, &vmin, &vmax)
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}

// select string value from certain values
func (con *Console) SelectValue(
	msg string, values []string, defval any) (string, error) {
	msg += fmt.Sprintf(" {%v}", strings.Join(values, "|"))
	pr := func(input string) (any, error) {
		for i := range values {
			if input == values[i] {
				return input, nil
			}
		}
		return "", fmt.Errorf("invalid value")
	}
	val, err := con.get_input(msg, defval, pr)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}
func (con *Console) SelectYesNo(msg string, defval any) (bool, error) {
	val, err := con.SelectValue(msg, []string{"y", "n"}, defval)
	if err != nil {
		return false, err
	}
	if val == "y" {
		return true, nil
	}
	return false, nil
}

/////////////////////////////////////////////////

// parse input string against regex
func (con *Console) parse_regex(
	input string, regex string) (string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	if !re.MatchString(input) {
		return "", fmt.Errorf("invalid input format")
	}
	return input, nil
}

// parse numbers
func (con *Console) parse_number(
	input string, vmin, vmax *int64) (int64, error) {
	v, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number format")
	}
	val := int64(v)
	if vmin != nil && vmax != nil {
		if val < *vmin || val > *vmax {
			return 0, fmt.Errorf("value out of range")
		}
	}
	return val, nil
}

// parse decimals
func (con *Console) parse_decimal(
	input string, decimals int, vmin, vmax *float64) (float64, error) {
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal format")
	}
	n := math.Pow(10, float64(decimals))
	val := float64(math.Round(v*n)) / n
	if vmin != nil && vmax != nil {
		if val < *vmin || val > *vmax {
			return 0, fmt.Errorf("value out of range")
		}
	}
	return val, nil
}
