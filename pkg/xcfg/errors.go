package xcfg

import (
	"errors"
	"fmt"
)

var (
	ErrError        = errors.New("")
	ErrLoadFailed   = fmt.Errorf("%wloading failed", ErrError)
	ErrSaveFailed   = fmt.Errorf("%wsaving failed", ErrError)
	ErrFileNotExist = fmt.Errorf("%wfile does not exist", ErrError)
	ErrEncodeFailed = fmt.Errorf("%wencoding failed", ErrError)
	ErrDecodeFailed = fmt.Errorf("%wdecoding failed", ErrError)
)
