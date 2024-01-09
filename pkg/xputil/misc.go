package xputil

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// set process title in OS process table, max 16 char
func SetProcTitle(title string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("not supported on %s OS", runtime.GOOS)
	}
	if len(title) == 0 {
		return fmt.Errorf("empty process title")
	}
	path := fmt.Sprintf("/proc/%d/comm", os.Getpid())
	fd, err := os.OpenFile(path, os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(title)
	return err
}

// run function f, return err and trace if function panics
func PanicExcept(f func() error) (err error, trace string) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
			stack := debug.Stack()
			// remove unwanted lines from trace
			indx := bytes.Index(stack, []byte("panic({"))
			trace = string(stack[indx:])
		}
	}()
	err = f()
	return
}
