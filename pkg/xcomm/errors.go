package xcomm

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Define common errors
var (
	ErrError      = errors.New("")
	ErrUri        = fmt.Errorf("%winvalid uri", ErrError)
	ErrConnection = fmt.Errorf("%wconnection failed", ErrError)
	ErrClosed     = fmt.Errorf("%wconnection closed", ErrError)
	ErrBreak      = fmt.Errorf("%woperation break", ErrError)
	ErrTimeout    = fmt.Errorf("%woperation timeout", ErrError)
	ErrRead       = fmt.Errorf("%wread failed", ErrError)
	ErrWrite      = fmt.Errorf("%wwrite failed", ErrError)
)

func errIsClosed(err error) bool {
	str_err := err.Error()
	if errors.Is(err, io.EOF) ||
		strings.Contains(str_err, "broken pipe") ||
		strings.Contains(str_err, "reset by peer") ||
		strings.Contains(str_err, "has been closed") ||
		strings.Contains(str_err, "input/output error") {
		return true
	}
	return false
}
