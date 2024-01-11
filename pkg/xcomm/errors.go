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
	ErrInvalidUri = fmt.Errorf("%winvalid uri format", ErrError)
	ErrOpen       = fmt.Errorf("%wopen connection failed", ErrError)
	ErrNotOpend   = fmt.Errorf("%wconnection not opened", ErrError)
	ErrClosed     = fmt.Errorf("%wconnection closed", ErrError)
	ErrBreak      = fmt.Errorf("%woperation break", ErrError)
	ErrTimeout    = fmt.Errorf("%woperation timeout", ErrError)
	ErrRead       = fmt.Errorf("%wread failed", ErrError)
	ErrWrite      = fmt.Errorf("%wwrite failed", ErrError)
)

func errIsClosed(err error) bool {
	if errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "reset by peer") {
		return true
	}
	return false
}
