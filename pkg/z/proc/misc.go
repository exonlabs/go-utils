package xputil

import (
	"fmt"
	"os"
	"runtime"
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
