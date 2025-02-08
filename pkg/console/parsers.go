// Copyright (c) 2024 ExonLabs, All rights reserved.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package console

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// RegexParser validates the input string using a provided regular expression.
// It returns the input if it matches the regex or an error if the input is invalid.
func RegexParser(input string, regex string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %v", err)
	}

	if !re.MatchString(input) {
		return "", fmt.Errorf("invalid input value")
	}

	return input, nil
}

// NumberParser converts the input string to an integer and validates
// it against optional minimum and maximum limits.
// Returns the parsed integer or an error if the input is invalid or out of range.
func NumberParser(input string, vmin, vmax *int64) (int64, error) {
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format, expected an integer")
	}

	// Validate against minimum and maximum limits if provided
	if (vmin != nil && val < *vmin) || (vmax != nil && val > *vmax) {
		return 0, fmt.Errorf("value out of range")
	}

	return val, nil
}

// DecimalParser converts the input string to a float and rounds it to
// the specified number of decimal places.
// It also validates the parsed float against optional minimum and maximum limits.
// Returns the rounded float or an error if the input is invalid or out of range.
func DecimalParser(input string, decimals int, vmin, vmax *float64) (float64, error) {
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format, expected a decimal")
	}

	// Round the value to the specified number of decimal places
	precision := math.Pow(10, float64(decimals))
	val = math.Round(val*precision) / precision

	// Validate against minimum and maximum limits if provided
	if (vmin != nil && val < *vmin) || (vmax != nil && val > *vmax) {
		return 0, fmt.Errorf("value out of range")
	}

	return val, nil
}
