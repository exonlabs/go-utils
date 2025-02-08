// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package console

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Console handles input prompts and validation.
type Console struct {
	Prompt string // Prompt is the string used to prompt the user.
	Trials int    // Trials defines how many input attempts are allowed.

	handler Handler // handler is the interface for reading/writing to the console.

	required bool // required marks the input as mandatory.
	hidden   bool // hidden indicates if the input should be masked (e.g., for passwords).

	parser func(string) (any, error) // parser is used to validate and parse input.

	cAsk *color.Color // cAsk is the color used for asking prompts.
	cErr *color.Color // cErr is the color used for showing errors.
}

// New creates a new Console instance with the provided Handler.
// Returns an error if the handler is nil.
func New(hnd Handler) (*Console, error) {
	if hnd == nil {
		return nil, errors.New("console handler cannot be empty")
	}
	return &Console{
		Prompt:  ">>",
		Trials:  3,
		handler: hnd,
		cAsk:    color.New(color.FgWhite, color.Bold),
		cErr:    color.New(color.FgRed, color.Bold),
	}, nil
}

// NewTermConsole creates a Console instance using a terminal handler.
// Returns an error if the terminal handler cannot be created.
func NewTermConsole() (*Console, error) {
	hnd, err := NewTermHandler()
	if err != nil {
		return nil, err
	}
	return New(hnd)
}

// Close closes the console handler.
func (c *Console) Close() error {
	if c.handler != nil {
		return c.handler.Close()
	}
	return nil
}

// Required marks the input as mandatory.
func (c *Console) Required() *Console {
	c.required = true
	return c
}

// Hidden masks the input (useful for sensitive information like passwords).
func (c *Console) Hidden() *Console {
	c.hidden = true
	return c
}

// Regex sets a regular expression to validate the input.
func (c *Console) Regex(regex string) *Console {
	c.parser = func(input string) (any, error) {
		return RegexParser(input, regex)
	}
	return c
}

// resetFlags resets input validation flags to default values.
func (c *Console) resetFlags() {
	c.required = false
	c.hidden = false
	c.parser = nil
}

// getInput reads and validates user input based on the provided message and default value.
// Returns the parsed input or an error if the input cannot be validated after the allowed trials.
func (c *Console) getInput(msg string, defVal any) (any, error) {
	// Format the input prompt with the prompt string and default value
	msg = fmt.Sprintf("%s %s: ", c.Prompt, msg)
	if defVal != nil {
		msg += fmt.Sprintf("[%v] ", defVal)
	} else if !c.required {
		msg += "[] "
	}
	msg = c.cAsk.Sprint(msg)

	// Helper function for retry messages
	showError := func(trial int, errMsg string) {
		if trial > 1 {
			errMsg += ", please try again"
		}
		c.handler.Write(c.cErr.Sprint("-- "+errMsg) + "\n\r")
	}

	// Attempt to get input based on the number of allowed trials
	var input string
	var err error
	for i := c.Trials; i > 0; i-- {
		if c.hidden {
			input, err = c.handler.ReadHidden(msg)
		} else {
			input, err = c.handler.Read(msg)
		}
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				c.handler.Write("\n\r")
				return nil, err
			}
			showError(i, err.Error())
			continue
		}

		if input == "" {
			if defVal != nil {
				return defVal, nil
			} else if c.required {
				showError(i, "input is required")
				continue
			} else {
				return "", nil
			}
		}

		if c.parser != nil {
			if val, err := c.parser(input); err != nil {
				showError(i, err.Error())
				continue
			} else {
				return val, nil
			}
		}

		return input, nil
	}

	return nil, fmt.Errorf("failed to get a valid input")
}

// ReadValue prompts the user for a string value with an optional default.
// If the input is empty and not required, it returns the default.
func (c *Console) ReadValue(msg string, defVal string) (string, error) {
	defer c.resetFlags()

	var v any
	if !c.required || defVal != "" {
		v = defVal
	}

	val, err := c.getInput(msg, v)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// ConfirmValue prompts the user to input a specific value (e.g., for confirmation).
// It will compare the input with the provided check value.
func (c *Console) ConfirmValue(msg, chkVal string) error {
	defer c.resetFlags()

	c.required = true
	c.parser = func(input string) (any, error) {
		if input == chkVal {
			return true, nil
		}
		return false, errors.New("values not matching")
	}

	_, err := c.getInput(msg, nil)
	return err
}

// ReadNumber prompts the user for an integer value with optional minimum and maximum limits.
func (c *Console) ReadNumber(msg string, defVal int64, limits ...int64) (int64, error) {
	defer c.resetFlags()

	if c.parser == nil {
		var vmin, vmax *int64
		if len(limits) >= 1 {
			vmin = &limits[0]
		}
		if len(limits) >= 2 {
			vmax = &limits[1]
		}
		c.parser = func(input string) (any, error) {
			return NumberParser(input, vmin, vmax)
		}
	}

	var v any
	if !c.required || defVal != 0 {
		v = defVal
	}

	val, err := c.getInput(msg, v)
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}

// ReadDecimal prompts the user for a decimal value with optional precision and limits.
func (c *Console) ReadDecimal(msg string, decimals int, defVal float64, limits ...float64) (float64, error) {
	defer c.resetFlags()

	if c.parser == nil {
		var vmin, vmax *float64
		if len(limits) >= 1 {
			vmin = &limits[0]
		}
		if len(limits) >= 2 {
			vmax = &limits[1]
		}
		c.parser = func(input string) (any, error) {
			return DecimalParser(input, decimals, vmin, vmax)
		}
	}

	var v any
	if !c.required || defVal != 0 {
		v = defVal
	}

	val, err := c.getInput(msg, v)
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}

// SelectValue prompts the user to choose from a list of string values.
func (c *Console) SelectValue(msg string, values []string, defVal string) (string, error) {
	defer c.resetFlags()

	strValues := strings.Join(values, "|")
	c.parser = func(input string) (any, error) {
		return RegexParser(input, fmt.Sprintf("^%s$", strValues))
	}

	var v any
	if !c.required || defVal != "" {
		v = defVal
	}

	val, err := c.getInput(fmt.Sprintf("%s {%v}", msg, strValues), v)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// SelectYesNo prompts the user for a yes/no selection.
// Returns true for "y" and false for "n".
func (c *Console) SelectYesNo(msg string, defVal string) (bool, error) {
	val, err := c.SelectValue(msg, []string{"y", "n"}, defVal)
	if err != nil {
		return false, err
	}
	return val == "y", nil
}
