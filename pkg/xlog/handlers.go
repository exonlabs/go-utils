package xlog

import "os"

type Handler interface {
	HandleRecord(string) error
}

// Write log messages to Stdout
type StdoutHandler struct {
}

func NewStdoutHandler() *StdoutHandler {
	return &StdoutHandler{}
}

func (h *StdoutHandler) HandleRecord(r string) error {
	_, err := os.Stdout.WriteString(r + "\n")
	return err
}

// Write log messages to file
type FileHandler struct {
	FilePath string
}

func NewFileHandler(path string) *FileHandler {
	return &FileHandler{
		FilePath: path,
	}
}

func (h *FileHandler) HandleRecord(r string) error {
	fh, err := os.OpenFile(
		h.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.WriteString(r + "\n")
	return err
}
